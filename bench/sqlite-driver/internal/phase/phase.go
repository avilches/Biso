// Package phase mide en que se va el tiempo dentro del proceso, para poder
// separar lo que cuesta abrir la base de datos de lo que cuesta la consulta.
//
// Solo escribe algo cuando la variable de entorno BENCH_PHASES vale 1, para que
// la medicion del reloj de pared no pague ningun trabajo extra.
package phase

import (
	"fmt"
	"os"
	"time"
)

// Timer acumula las marcas de tiempo de un proceso.
type Timer struct {
	on     bool
	start  time.Time
	last   time.Time
	marks  []mark
	binary string
}

type mark struct {
	name string
	d    time.Duration
}

// New crea el medidor. `binary` es el nombre que sale en la salida de error.
func New(binary string) *Timer {
	now := time.Now()
	return &Timer{
		on:     os.Getenv("BENCH_PHASES") == "1",
		start:  now,
		last:   now,
		binary: binary,
	}
}

// Mark cierra una fase con el nombre dado.
func (t *Timer) Mark(name string) {
	if !t.on {
		return
	}
	now := time.Now()
	t.marks = append(t.marks, mark{name, now.Sub(t.last)})
	t.last = now
}

// Report escribe por la salida de error una linea por fase y el total interno,
// en milisegundos con tres decimales.
func (t *Timer) Report() {
	if !t.on {
		return
	}
	total := time.Since(t.start)
	for _, m := range t.marks {
		fmt.Fprintf(os.Stderr, "phase\t%s\t%s\t%.3f\n", t.binary, m.name, float64(m.d.Microseconds())/1000.0)
	}
	fmt.Fprintf(os.Stderr, "phase\t%s\t%s\t%.3f\n", t.binary, "total_in_main", float64(total.Microseconds())/1000.0)
}
