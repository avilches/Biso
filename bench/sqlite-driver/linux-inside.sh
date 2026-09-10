#!/bin/sh
# Se ejecuta DENTRO de un contenedor de Linux con Go y un compilador de C.
# Lo lanza linux.sh, y no tiene sentido ejecutarlo a mano en macOS.
#
# Compila y mide los cuatro controladores en el MISMO contenedor, porque el suelo
# de arrancar un proceso cambia de una imagen a otra y comparar cifras de dos
# contenedores distintos no dice nada.
#
# Aqui esta la unica forma de tener la cifra del controlador de C en Linux: no se
# puede compilar de forma cruzada desde macOS, asi que se construye aqui.

set -u

export GOFLAGS=-mod=mod
export GOCACHE=/tmp/gocache
export GOMODCACHE=/tmp/gomodcache
RUNS="${RUNS:-200}"

echo "== Entorno del contenedor"
grep PRETTY_NAME /etc/os-release
uname -srm
echo "go: $(go version)"
apk add --no-cache gcc musl-dev >/dev/null 2>&1
echo "cc: $(gcc --version | head -1)"
if [ -f /etc/services ]; then
  echo "lineas de /etc/services: $(wc -l < /etc/services)"
else
  echo "/etc/services: no existe en esta imagen"
fi
echo

echo "== Compilacion dentro del contenedor"
start=$(date +%s)
CGO_ENABLED=1 go build -o /tmp/ls-mattn ./cmd/ls-mattn || exit 1
end=$(date +%s)
echo "ls-mattn     con cgo,  en frio: $((end - start))s, $(stat -c%s /tmp/ls-mattn) bytes"
for d in modernc ncruces zombiezen; do
  start=$(date +%s)
  CGO_ENABLED=0 go build -o "/tmp/ls-$d" "./cmd/ls-$d" || exit 1
  end=$(date +%s)
  echo "ls-$d  sin cgo,  en frio: $((end - start))s, $(stat -c%s "/tmp/ls-$d") bytes"
done
CGO_ENABLED=0 go build -o /tmp/gen ./cmd/gen || exit 1
CGO_ENABLED=0 go build -o /tmp/timeit ./cmd/timeit || exit 1
CGO_ENABLED=0 go build -o /tmp/noop-bare ./cmd/noop-bare || exit 1
echo

echo "== El tablero se regenera dentro del contenedor"
/tmp/gen -out /tmp/board.db -tasks 300
/tmp/gen -out /tmp/board-3000.db -tasks 3000
echo

echo "== La salida sigue siendo la misma en Linux"
for d in mattn modernc ncruces zombiezen; do
  "/tmp/ls-$d" /tmp/board.db 300 > "/tmp/out-$d.txt" 2>/dev/null
  echo "$d $(md5sum < "/tmp/out-$d.txt")"
done
echo

echo "== Tiempo de init por paquete, el motivo principal de venir a Linux"
for d in mattn modernc ncruces zombiezen; do
  echo "--- $d"
  GODEBUG=inittrace=1 "/tmp/ls-$d" /tmp/board.db 2>&1 >/dev/null \
    | grep clock \
    | sed 's/.*@\([0-9.]*\) ms, \([0-9.]*\) ms clock.*/\2 &/' \
    | sort -rn | head -3
done
echo

echo "== Reloj de pared dentro del contenedor"
/tmp/timeit -n "$RUNS" -warmup 20 -label noop-bare -- /tmp/noop-bare
for d in mattn modernc ncruces zombiezen; do
  /tmp/timeit -n "$RUNS" -warmup 20 -label "ls-$d" -- "/tmp/ls-$d" /tmp/board.db
done
echo "-- tablero de 3.000 tareas"
for d in mattn modernc ncruces zombiezen; do
  /tmp/timeit -n "$RUNS" -warmup 20 -label "big-$d" -- "/tmp/ls-$d" /tmp/board-3000.db
done
echo

echo "== Con que se enlaza cada binario"
for d in mattn modernc ncruces zombiezen; do
  echo "--- ls-$d"
  ldd "/tmp/ls-$d" 2>&1 | head -4 | sed 's/^/    /'
done
