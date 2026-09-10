package board

import (
	"strconv"
	"strings"
	"time"
)

// Collected es el estado que un cargador va llenando mientras recorre las cinco
// consultas. Existe para que los cuatro controladores compartan todo lo que no
// es hablar con SQLite: el ensamblado, la urgencia, el orden y el formato.
type Collected struct {
	Config map[string]string
	Tasks  []*Task
	byID   map[int64]*Task
}

// NewCollected crea un acumulador vacio.
func NewCollected() *Collected {
	return &Collected{
		Config: make(map[string]string, 16),
		Tasks:  make([]*Task, 0, 512),
		byID:   make(map[int64]*Task, 512),
	}
}

// AddConfig guarda una clave de la configuracion del tablero.
func (c *Collected) AddConfig(key, value string) {
	c.Config[key] = value
}

// AddTask registra una tarea leida de la consulta principal.
func (c *Collected) AddTask(id int64, title, status, typ, priority, due string, ordinal int64, hasOrdinal bool, createdAt string, openQuestion bool) {
	t := &Task{
		ID:           id,
		Title:        title,
		Status:       status,
		Type:         typ,
		Priority:     priority,
		Due:          due,
		Ordinal:      ordinal,
		HasOrdinal:   hasOrdinal,
		OpenQuestion: openQuestion,
	}
	if ts, err := time.Parse(time.RFC3339, createdAt); err == nil {
		t.CreatedAt = ts
	}
	c.Tasks = append(c.Tasks, t)
	c.byID[id] = t
}

// AddAssignee anade una persona asignada a la tarea, en el orden de llegada.
func (c *Collected) AddAssignee(taskID int64, who string) {
	if t := c.byID[taskID]; t != nil {
		t.Assignees = append(t.Assignees, who)
	}
}

// AddCriteria fija los recuentos de criterios de aceptacion de una tarea.
func (c *Collected) AddCriteria(taskID int64, total, done int) {
	if t := c.byID[taskID]; t != nil {
		t.AcTotal = total
		t.AcDone = done
	}
}

// AddEdge registra una arista de dependencia. `from` depende de `to`.
func (c *Collected) AddEdge(from int64, fromStatus string, to int64, toStatus string) {
	if toStatus != TerminalStatus {
		if t := c.byID[from]; t != nil {
			t.Blocked = true
		}
	}
	if fromStatus != TerminalStatus {
		if t := c.byID[to]; t != nil {
			t.Blocks = true
		}
	}
}

// Finish calcula la urgencia, ordena, recorta al limite y devuelve la salida
// literal de `biso ls` mas el numero de filas mostradas y el total que encaja.
func (c *Collected) Finish(now time.Time, limit int) (out string, shown, total int) {
	coef := CoefficientsFrom(c.Config)
	for _, t := range c.Tasks {
		t.Urgency = ComputeUrgency(t, coef, now)
	}
	SortDefault(c.Tasks)
	total = len(c.Tasks)
	view := c.Tasks
	if limit >= 0 && limit < len(view) {
		view = view[:limit]
	}
	shown = len(view)
	var sb strings.Builder
	sb.Grow(shown * 160)
	Render(BuildRows(view), &sb)
	return sb.String(), shown, total
}

// TruncationWarning devuelve el aviso de recorte de la seccion 10.4, o la
// cadena vacia si no se ha recortado nada.
func TruncationWarning(shown, total int) string {
	if shown >= total {
		return ""
	}
	return "warning: " + strconv.Itoa(total-shown) + " more tasks match; showing " +
		strconv.Itoa(shown) + " of " + strconv.Itoa(total) + "\n" +
		"hint: narrow with -s, --type or -l, or ask for everything with --all\n"
}
