// Package board contiene el modelo minimo de tarea, el calculo de urgencia, la
// regla de orden y el formato de columnas de `biso ls`, tal y como los definen
// docs/spec/modelo-de-datos/index.md (con su pagina "La urgencia" en urgencia.md) y docs/spec/cmd/ls.md.
//
// Este paquete es identico para los cuatro controladores que se comparan: lo
// unico que cambia entre binarios es como se abre la base de datos y como se
// recorren las filas. Todo lo que viene despues de tener las filas en memoria
// es el mismo codigo, para que la medicion compare controladores y no
// implementaciones.
package board

import (
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// TerminalStatus y ActiveStatus son dos de los tres papeles de estado que
// docs/spec/cmd/config.md declara en su seccion "Las claves" (initial_status,
// active_status, terminal_status). El banco los deja fijos porque el generador
// crea siempre el mismo vocabulario.
const (
	TerminalStatus = "Done"
	ActiveStatus   = "In Progress"
)

// Task es la tarea reducida a los campos que `biso ls` necesita para imprimir
// sus ocho columnas y para calcular la urgencia. No es el modelo completo de
// docs/spec/modelo-de-datos/index.md a proposito: la primera regla de "El presupuesto
// de arranque" (docs/spec/presupuestos.md) prohibe leer lo que la invocacion no
// va a usar.
type Task struct {
	ID         int64
	Title      string
	Status     string
	Type       string
	Priority   string
	Due        string
	Ordinal    int64
	HasOrdinal bool
	CreatedAt  time.Time

	// OpenQuestion vale cierto cuando la tarea tiene una pregunta sin
	// responder, que es lo que anula el termino `activa` de la urgencia.
	OpenQuestion bool

	Assignees []string
	AcDone    int
	AcTotal   int

	// Blocked y Blocks son los dos derivados que salen de las dependencias.
	Blocked bool
	Blocks  bool

	Urgency float64
}

// Coefficients son los siete pesos configurables de "La urgencia"
// (docs/spec/modelo-de-datos/urgencia.md).
type Coefficients struct {
	Priority float64
	Active   float64
	Blocking float64
	Blocked  float64
	Due      float64
	Criteria float64
	Age      float64
}

// DefaultCoefficients devuelve los valores por defecto de "La urgencia".
func DefaultCoefficients() Coefficients {
	return Coefficients{
		Priority: 6.0,
		Active:   4.0,
		Blocking: 8.0,
		Blocked:  -5.0,
		Due:      12.0,
		Criteria: 1.0,
		Age:      0.5,
	}
}

// CoefficientsFrom lee los pesos del mapa de configuracion del tablero, que es
// lo que de verdad hace el programa: los valores viven en la tabla `board`.
func CoefficientsFrom(cfg map[string]string) Coefficients {
	c := DefaultCoefficients()
	get := func(key string, dst *float64) {
		if raw, ok := cfg[key]; ok {
			if v, err := strconv.ParseFloat(raw, 64); err == nil {
				*dst = v
			}
		}
	}
	get("urgency.priority", &c.Priority)
	get("urgency.active", &c.Active)
	get("urgency.blocking", &c.Blocking)
	get("urgency.blocked", &c.Blocked)
	get("urgency.due", &c.Due)
	get("urgency.criteria", &c.Criteria)
	get("urgency.age", &c.Age)
	return c
}

func priorityWeight(p string) float64 {
	switch p {
	case "high":
		return 1.0
	case "medium":
		return 0.5
	case "low":
		return 0.0
	default:
		return 0.3
	}
}

// ComputeUrgency aplica la formula de "La urgencia" (docs/spec/modelo-de-datos/urgencia.md).
func ComputeUrgency(t *Task, c Coefficients, now time.Time) float64 {
	if t.Status == TerminalStatus {
		return 0.0
	}
	u := c.Priority * priorityWeight(t.Priority)
	if t.Status == ActiveStatus && !t.OpenQuestion {
		u += c.Active
	}
	if t.Blocks {
		u += c.Blocking
	}
	if t.Blocked {
		u += c.Blocked
	}
	u += c.Due * proximity(t.Due, now)
	if t.AcTotal > 0 {
		u += c.Criteria
	}
	ageDays := now.Sub(t.CreatedAt).Hours() / 24.0
	ageTerm := ageDays / 30.0
	if ageTerm > 4.0 {
		ageTerm = 4.0
	}
	if ageTerm < 0 {
		ageTerm = 0
	}
	u += c.Age * ageTerm
	// Redondeo a un decimal, como manda "La urgencia".
	return float64(int64(u*10+copySign(0.5, u))) / 10.0
}

func copySign(mag, sign float64) float64 {
	if sign < 0 {
		return -mag
	}
	return mag
}

func proximity(due string, now time.Time) float64 {
	if due == "" {
		return 0.0
	}
	d, err := time.Parse("2006-01-02", due)
	if err != nil {
		return 0.0
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	days := d.Sub(today).Hours() / 24.0
	p := (30.0 - days) / 30.0
	if p < 0.0 {
		return 0.0
	}
	if p > 1.0 {
		return 1.0
	}
	return p
}

// SortDefault aplica la tupla de orden por defecto de docs/spec/cmd/ls.md: primero
// las que tienen `ordinal`, ascendente; entre las que no lo tienen, `urgency`
// descendente; y cualquier empate por identificador ascendente.
func SortDefault(tasks []*Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i], tasks[j]
		if a.HasOrdinal != b.HasOrdinal {
			return a.HasOrdinal
		}
		if a.HasOrdinal && b.HasOrdinal {
			if a.Ordinal != b.Ordinal {
				return a.Ordinal < b.Ordinal
			}
			return a.ID < b.ID
		}
		if a.Urgency != b.Urgency {
			return a.Urgency > b.Urgency
		}
		return a.ID < b.ID
	})
}

// cellWidth devuelve la anchura en celdas de terminal de una runa, con la regla
// de docs/spec/cmd/ls.md: las marcas combinantes miden cero, los ideogramas de Asia
// oriental y los emoji miden dos, y todo lo demas mide una.
func cellWidth(r rune) int {
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
		return 0
	}
	switch {
	case r >= 0x1100 && r <= 0x115F,
		r >= 0x2E80 && r <= 0xA4CF,
		r >= 0xAC00 && r <= 0xD7A3,
		r >= 0xF900 && r <= 0xFAFF,
		r >= 0xFE30 && r <= 0xFE6F,
		r >= 0xFF00 && r <= 0xFF60,
		r >= 0xFFE0 && r <= 0xFFE6,
		r >= 0x1F300 && r <= 0x1F64F,
		r >= 0x1F900 && r <= 0x1F9FF,
		r >= 0x20000 && r <= 0x3FFFD:
		return 2
	}
	return 1
}

// StringWidth mide una cadena en celdas de terminal.
func StringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += cellWidth(r)
	}
	return w
}

// TruncateCells recorta a `limit` celdas anadiendo tres puntos si hace falta,
// contando los puntos dentro del limite y sin partir nunca un grafema.
func TruncateCells(s string, limit int) string {
	if StringWidth(s) <= limit {
		return s
	}
	budget := limit - 3
	w := 0
	cut := 0
	for i, r := range s {
		cw := cellWidth(r)
		if w+cw > budget {
			cut = i
			break
		}
		w += cw
		cut = i + len(string(r))
	}
	// No partir un grafema: si lo que sigue al corte es una marca
	// combinante, retroceder hasta la frontera anterior.
	for cut > 0 {
		r := []rune(s[cut:])
		if len(r) == 0 {
			break
		}
		if cellWidth(r[0]) == 0 {
			prev := strings.LastIndexFunc(s[:cut], func(rune) bool { return true })
			if prev <= 0 {
				break
			}
			cut = prev
			continue
		}
		break
	}
	return s[:cut] + "..."
}

const titleCells = 100

// Row es una fila ya formateada en sus ocho valores, antes de rellenar anchos.
type Row [8]string

// BuildRows construye las ocho columnas de docs/spec/cmd/ls.md para las tareas
// dadas, en el orden en que llegan.
func BuildRows(tasks []*Task) []Row {
	rows := make([]Row, 0, len(tasks))
	for _, t := range tasks {
		var r Row
		r[0] = "TASK-" + strconv.FormatInt(t.ID, 10)
		r[1] = t.Status
		r[2] = dash(t.Type)
		r[3] = dash(t.Priority)
		r[4] = TruncateCells(t.Title, titleCells)
		if t.AcTotal > 0 {
			r[5] = "ac " + strconv.Itoa(t.AcDone) + "/" + strconv.Itoa(t.AcTotal)
		} else {
			r[5] = "-"
		}
		switch len(t.Assignees) {
		case 0:
			r[6] = "-"
		case 1:
			r[6] = t.Assignees[0]
		default:
			r[6] = t.Assignees[0] + "+" + strconv.Itoa(len(t.Assignees)-1)
		}
		r[7] = dash(t.Due)
		rows = append(rows, r)
	}
	return rows
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// Render escribe la tabla con dos espacios entre columnas y las columnas 1 a 7
// rellenadas al ancho de su valor mas largo, como manda docs/spec/cmd/ls.md.
func Render(rows []Row, out *strings.Builder) {
	var widths [8]int
	for _, r := range rows {
		for i := 0; i < 7; i++ {
			if w := StringWidth(r[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}
	for _, r := range rows {
		for i := 0; i < 8; i++ {
			out.WriteString(r[i])
			if i == 7 {
				break
			}
			for pad := widths[i] - StringWidth(r[i]); pad > 0; pad-- {
				out.WriteByte(' ')
			}
			out.WriteString("  ")
		}
		out.WriteByte('\n')
	}
}
