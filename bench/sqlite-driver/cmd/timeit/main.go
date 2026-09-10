// timeit mide el reloj de pared de un proceso completo desde fuera, repitiendolo
// muchas veces. Es el sustituto de hyperfine, que no esta instalado en la
// maquina de medida, y hace lo mismo que hace falta aqui: lanzar el binario,
// esperar a que termine, y quedarse con la distribucion de los tiempos.
//
// Lo que mide incluye el fork y el exec, el arranque del runtime de Go y el
// trabajo del programa, que es exactamente el "reloj de pared" del que habla la
// seccion 4.13 de docs/SPEC.md. La salida y los errores del proceso medido se
// descartan, para no pagar el coste de la tuberia.
//
// Uso: timeit -n 200 -warmup 20 -label ls-modernc -- ./bin/ls-modernc board.db
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"time"
)

type result struct {
	Label    string  `json:"label"`
	Runs     int     `json:"runs"`
	MinMs    float64 `json:"min_ms"`
	P50Ms    float64 `json:"p50_ms"`
	P90Ms    float64 `json:"p90_ms"`
	MaxMs    float64 `json:"max_ms"`
	MeanMs   float64 `json:"mean_ms"`
	StddevMs float64 `json:"stddev_ms"`
}

func main() {
	n := flag.Int("n", 200, "repeticiones que se miden")
	warmup := flag.Int("warmup", 20, "repeticiones previas que se descartan")
	label := flag.String("label", "", "nombre de la medida")
	asJSON := flag.Bool("json", false, "imprime el resultado en JSON")
	flag.Parse()

	argv := flag.Args()
	if len(argv) == 0 {
		fmt.Fprintln(os.Stderr, "usage: timeit [-n N] [-warmup N] [-label L] [-json] -- cmd args...")
		os.Exit(2)
	}
	if *label == "" {
		*label = argv[0]
	}

	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "timeit: %v\n", err)
		os.Exit(1)
	}
	defer devnull.Close()

	run := func() (time.Duration, error) {
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Stdout = devnull
		cmd.Stderr = devnull
		start := time.Now()
		err := cmd.Run()
		return time.Since(start), err
	}

	for i := 0; i < *warmup; i++ {
		if _, err := run(); err != nil {
			fmt.Fprintf(os.Stderr, "timeit: warmup failed: %v\n", err)
			os.Exit(1)
		}
	}

	samples := make([]float64, 0, *n)
	for i := 0; i < *n; i++ {
		d, err := run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "timeit: run %d failed: %v\n", i, err)
			os.Exit(1)
		}
		samples = append(samples, float64(d.Microseconds())/1000.0)
	}
	sort.Float64s(samples)

	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		variance += (s - mean) * (s - mean)
	}
	variance /= float64(len(samples))

	r := result{
		Label:    *label,
		Runs:     len(samples),
		MinMs:    samples[0],
		P50Ms:    percentile(samples, 0.50),
		P90Ms:    percentile(samples, 0.90),
		MaxMs:    samples[len(samples)-1],
		MeanMs:   mean,
		StddevMs: math.Sqrt(variance),
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(r)
		return
	}
	fmt.Printf("%-16s n=%d  min=%.2f  p50=%.2f  p90=%.2f  max=%.2f  media=%.2f  desv=%.2f  (ms)\n",
		r.Label, r.Runs, r.MinMs, r.P50Ms, r.P90Ms, r.MaxMs, r.MeanMs, r.StddevMs)
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Round(p * float64(len(sorted)-1)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
