#!/bin/bash
# Mide los cuatro controladores de SQLite candidatos para `biso`, de principio a
# fin, y deja los resultados en results/.
#
# Se ejecuta con `./run.sh` desde este directorio. No necesita ningun argumento
# ni ninguna herramienta que no venga con Go y con macOS, y no toca la cache de
# compilacion del usuario: la medida en frio usa una cache propia dentro de
# work/.
#
# Lo que hace, en orden:
#   1. Comprueba que hay un Go y un compilador de C.
#   2. Compila los binarios de medida.
#   3. Genera dos tableros, de 300 y de 3.000 tareas.
#   4. Comprueba que los cuatro controladores dan la misma salida byte a byte.
#   5. Mide el reloj de pared de cada uno desde fuera del proceso.
#   6. Separa lo que cuesta abrir la base de datos de lo que cuesta consultarla.
#   7. Mide tamano del binario, dependencias del sistema y compilacion cruzada.
#   8. Comprueba que cada controlador soporta lo que el esquema da por hecho.

set -uo pipefail

cd "$(dirname "$0")"
ROOT="$PWD"
BIN="$ROOT/bin"
WORK="$ROOT/work"
RESULTS="$ROOT/results"

# Repeticiones de la medida del reloj de pared. Se pueden bajar para una prueba
# rapida con RUNS=30 ./run.sh
RUNS="${RUNS:-300}"
WARMUP="${WARMUP:-30}"
PHASE_RUNS="${PHASE_RUNS:-50}"

DRIVERS="mattn modernc ncruces zombiezen"

mkdir -p "$BIN" "$WORK" "$RESULTS"

log() { printf '\n== %s\n' "$*"; }

# ---------------------------------------------------------------- 1. entorno
log "Entorno"
{
  echo "fecha: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "go: $(go version)"
  echo "cc: $(cc --version 2>/dev/null | head -1)"
  echo "os: $(sw_vers -productName) $(sw_vers -productVersion) ($(sw_vers -buildVersion))"
  echo "kernel: $(uname -srm)"
  echo "cpu: $(sysctl -n machdep.cpu.brand_string 2>/dev/null || sysctl -n hw.model)"
  echo "cores: $(sysctl -n hw.ncpu) ($(sysctl -n hw.perflevel0.logicalcpu 2>/dev/null || echo '?') de rendimiento)"
  echo "ram: $(( $(sysctl -n hw.memsize) / 1024 / 1024 / 1024 )) GiB"
  echo "cgo: $(go env CGO_ENABLED)"
  echo
  echo "versiones de los controladores:"
  go list -m all | grep -E 'sqlite|wazero|libc' | sed 's/^/  /'
} | tee "$RESULTS/00-entorno.txt"

# ------------------------------------------------------------- 2. compilacion
log "Compilando los binarios de medida"
go build -o "$BIN/gen" ./cmd/gen || exit 1
go build -o "$BIN/timeit" ./cmd/timeit || exit 1
go build -o "$BIN/noop-bare" ./cmd/noop-bare || exit 1
for d in $DRIVERS; do
  go build -o "$BIN/ls-$d" "./cmd/ls-$d" || exit 1
  go build -o "$BIN/noop-$d" "./cmd/noop-$d" || exit 1
done

# ---------------------------------------------------------------- 3. tableros
log "Generando los tableros de prueba"
# Las rutas van relativas a proposito: el fichero de resultados se commitea, y
# una ruta absoluta le metria dentro el nombre del worktree de quien midio.
"$BIN/gen" -out work/board.db -tasks 300 | tee "$RESULTS/01-tableros.txt"
"$BIN/gen" -out work/board-3000.db -tasks 3000 | tee -a "$RESULTS/01-tableros.txt"

# ------------------------------------------------ 4. la salida es la misma
log "Comprobando que los cuatro dan la misma salida"
{
  for d in $DRIVERS; do
    "$BIN/ls-$d" "$WORK/board.db" 300 > "$WORK/out-$d.txt" 2> "$WORK/err-$d.txt"
    printf '%-12s codigo=%d  filas=%s  md5=%s\n' \
      "$d" "$?" "$(wc -l < "$WORK/out-$d.txt" | tr -d ' ')" \
      "$(md5 -q "$WORK/out-$d.txt")"
  done
  distinct=$(md5 -q "$WORK"/out-*.txt | sort -u | wc -l | tr -d ' ')
  if [ "$distinct" = "1" ]; then
    echo "veredicto: las cuatro salidas son identicas byte a byte"
  else
    echo "veredicto: LAS SALIDAS NO COINCIDEN, la medida no vale"
  fi
} | tee "$RESULTS/02-salida-identica.txt"

# --------------------------------------------------------- 5. reloj de pared
log "Reloj de pared, tablero de 300 tareas ($RUNS repeticiones por binario)"
{
  echo "# Reloj de pared del proceso completo, medido desde fuera con cmd/timeit."
  echo "# Incluye fork, exec, arranque del runtime de Go y todo el trabajo."
  echo
  echo "## Suelo: procesos que no consultan nada"
  "$BIN/timeit" -n "$RUNS" -warmup "$WARMUP" -label "noop-bare" -- "$BIN/noop-bare"
  for d in $DRIVERS; do
    "$BIN/timeit" -n "$RUNS" -warmup "$WARMUP" -label "noop-$d" -- "$BIN/noop-$d"
  done
  echo
  echo "## biso ls sobre 300 tareas, limite 30 (el caso del presupuesto de 4.13)"
  for d in $DRIVERS; do
    "$BIN/timeit" -n "$RUNS" -warmup "$WARMUP" -label "ls-$d" -- "$BIN/ls-$d" "$WORK/board.db"
  done
  echo
  echo "## biso ls --all sobre 300 tareas (imprime las 300 filas)"
  for d in $DRIVERS; do
    "$BIN/timeit" -n "$RUNS" -warmup "$WARMUP" -label "all-$d" -- "$BIN/ls-$d" "$WORK/board.db" 300
  done
  echo
  echo "## biso ls sobre 3.000 tareas, limite 30 (diez veces el tablero de referencia)"
  for d in $DRIVERS; do
    "$BIN/timeit" -n "$RUNS" -warmup "$WARMUP" -label "big-$d" -- "$BIN/ls-$d" "$WORK/board-3000.db"
  done
} | tee "$RESULTS/03-reloj-de-pared.txt"

# ------------------------------------------------------------- 6. las fases
log "Fases internas: abrir la base de datos frente a consultarla"
{
  echo "# Mediana de $PHASE_RUNS ejecuciones, en milisegundos, con BENCH_PHASES=1."
  echo "# 'open' es abrir el fichero y leer el primer PRAGMA. 'query' son las"
  echo "# cinco consultas de biso ls. 'compute' es urgencia, orden y formato."
  echo "# 'total_in_main' es la suma, y lo que falta hasta el reloj de pared es"
  echo "# el arranque del proceso, que ningun controlador puede evitar."
  echo
  printf '%-12s %8s %8s %8s %8s %14s\n' driver args open query compute total_in_main
  for d in $DRIVERS; do
    rm -f "$WORK/phases-$d.txt"
    for _ in $(seq "$PHASE_RUNS"); do
      BENCH_PHASES=1 "$BIN/ls-$d" "$WORK/board.db" >/dev/null 2>> "$WORK/phases-$d.txt"
    done
    awk -v drv="$d" '
      $1 == "phase" { v[$3] = v[$3] " " $4; n[$3]++ }
      END {
        split("args open query compute total_in_main", order, " ")
        printf "%-12s", drv
        for (i = 1; i <= 5; i++) {
          k = order[i]
          c = split(v[k], a, " ")
          for (x = 1; x < c; x++) for (y = x + 1; y <= c; y++) if (a[y] + 0 < a[x] + 0) { t = a[x]; a[x] = a[y]; a[y] = t }
          med = a[int((c + 1) / 2)]
          if (k == "total_in_main") printf " %14.3f", med; else printf " %8.3f", med
        }
        printf "\n"
      }' "$WORK/phases-$d.txt"
  done
} | tee "$RESULTS/04-fases.txt"

log "Coste del init de cada paquete, antes de que main empiece"
{
  echo "# GODEBUG=inittrace=1 dice cuanto tarda el init de cada paquete antes de"
  echo "# llegar a main. El medidor de fases de internal/phase arranca ya dentro de"
  echo "# main, asi que esto es lo unico que ve lo que pasa antes, y es donde se"
  echo "# explica la mayor parte de la diferencia entre los controladores."
  echo
  TOP=4 ./inittrace.sh
} | tee "$RESULTS/05-init.txt"

# ------------------------------------------------------ 7. tamano y enlazado
log "Tamano del binario y dependencias del sistema"
{
  printf '%-14s %12s  %s\n' binario bytes "bibliotecas del sistema (otool -L)"
  printf '%-14s %12d  %s\n' "noop-bare" "$(stat -f%z "$BIN/noop-bare")" \
    "$(otool -L "$BIN/noop-bare" | tail -n +2 | awk '{print $1}' | xargs echo)"
  for d in $DRIVERS; do
    printf '%-14s %12d  %s\n' "ls-$d" "$(stat -f%z "$BIN/ls-$d")" \
      "$(otool -L "$BIN/ls-$d" | tail -n +2 | awk '{print $1}' | xargs echo)"
  done
} | tee "$RESULTS/06-tamano-y-enlazado.txt"

log "Tiempo de compilacion"
{
  echo "# 'frio' es con una cache de compilacion vacia y propia del banco, asi que"
  echo "# incluye compilar la biblioteca estandar y todas las dependencias. Es el"
  echo "# coste de la primera compilacion en una maquina de CI."
  echo "# 'caliente' es tocar el main.go del propio programa y volver a compilar,"
  echo "# que es la vuelta que paga un agente en cada iteracion."
  echo
  printf '%-14s %10s %10s\n' binario frio caliente
  for d in $DRIVERS; do
    cold_cache="$WORK/gocache-cold-$d"
    rm -rf "$cold_cache"
    mkdir -p "$cold_cache"
    t0=$(python3 -c 'import time;print(time.time())')
    GOCACHE="$cold_cache" go build -o "$WORK/cold-$d" "./cmd/ls-$d" >/dev/null 2>&1
    t1=$(python3 -c 'import time;print(time.time())')
    rm -rf "$cold_cache"

    go build -o "$WORK/warm-$d" "./cmd/ls-$d" >/dev/null 2>&1
    touch "./cmd/ls-$d/main.go"
    t2=$(python3 -c 'import time;print(time.time())')
    go build -o "$WORK/warm-$d" "./cmd/ls-$d" >/dev/null 2>&1
    t3=$(python3 -c 'import time;print(time.time())')

    python3 -c "print('%-14s %9.2fs %9.2fs' % ('ls-$d', $t1-$t0, $t3-$t2))"
  done
} | tee "$RESULTS/07-compilacion.txt"

log "Compilacion cruzada desde esta maquina, sin instalar nada"
{
  echo '# Lo que se prueba es si go build produce un binario para otra plataforma'
  echo '# con lo que ya hay instalado. Un fallo aqui es lo que de verdad decide la'
  echo '# eleccion, porque distribuir el programa depende de ello.'
  echo
  printf '%-14s %-18s %-8s %s\n' controlador plataforma CGO resultado
  for d in $DRIVERS; do
    for target in linux/amd64 linux/arm64 windows/amd64; do
      goos="${target%%/*}"
      goarch="${target##*/}"
      out="$WORK/cross-$d-$goos-$goarch"
      err=$(CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -o "$out" "./cmd/ls-$d" 2>&1)
      if [ $? -eq 0 ]; then
        r="ok, $(stat -f%z "$out") bytes"
      else
        r="FALLA: $(echo "$err" | head -1 | cut -c1-90)"
      fi
      printf '%-14s %-18s %-8s %s\n' "$d" "$target" "0" "$r"
    done
  done
  echo
  echo "# Y lo mismo con cgo activado, que es lo que mattn necesita para funcionar."
  for target in linux/amd64 windows/amd64; do
    goos="${target%%/*}"
    goarch="${target##*/}"
    err=$(CGO_ENABLED=1 GOOS="$goos" GOARCH="$goarch" go build -o "$WORK/cross-cgo-$goos" ./cmd/ls-mattn 2>&1)
    if [ $? -eq 0 ]; then r="ok"; else r="FALLA: $(echo "$err" | head -1 | cut -c1-90)"; fi
    printf '%-14s %-18s %-8s %s\n' "mattn" "$target" "1" "$r"
  done
  echo
  echo "# Un binario compilado sin cgo con el controlador de C se enlaza, pero no"
  echo "# tiene controlador registrado. Esto es lo que pasa al ejecutarlo:"
  CGO_ENABLED=0 go build -o "$WORK/nocgo-mattn" ./cmd/ls-mattn 2>&1 | head -3
  if [ -x "$WORK/nocgo-mattn" ]; then
    echo "  compila sin cgo: si"
    echo "  al ejecutarlo: $("$WORK/nocgo-mattn" "$WORK/board.db" 2>&1 >/dev/null | head -1)"
  else
    echo "  compila sin cgo: no"
  fi
} | tee "$RESULTS/08-compilacion-cruzada.txt"

# -------------------------------------------------------- 8. lo que soportan
log "Lo que cada controlador soporta del esquema"
{
  echo "# Cada controlador trae su propia copia de SQLite compilada con sus propias"
  echo "# opciones, asi que esto no es una formalidad: una extension puede estar en"
  echo "# uno y faltar en otro. internal/features/features.go dice de donde sale la"
  echo "# exigencia de cada linea."
  echo
  for d in $DRIVERS; do
    BENCH_FEATURES=1 "$BIN/ls-$d" 2>&1
    echo
  done
} | tee "$RESULTS/09-soporte.txt"

log "Falta la medida en Linux, que va aparte porque necesita Docker: ./linux.sh"
log "Hecho. Los resultados de macOS estan en results/"
ls -la "$RESULTS"
