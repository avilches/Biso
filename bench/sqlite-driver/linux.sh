#!/bin/bash
# Repite la medida en Linux, dentro de un contenedor, porque la maquina de
# referencia de la seccion 4.13 de docs/SPEC.md es la que ejecuta la integracion
# continua y esa no es macOS.
#
# El motivo concreto por el que hace falta, y no se puede dar por supuesto: en
# macOS y en los BSD, modernc.org/libc importa sin condiciones un paquete que
# parsea /etc/services al arrancar el proceso, y en Linux no lo importa. Eso mueve
# milisegundos y cambia el orden de los candidatos.
#
# Se ejecuta con `./linux.sh`. Necesita un Docker que funcione (en esta maquina es
# Dory). Todo, incluido compilar, pasa dentro del contenedor, asi que tambien es
# la unica forma de medir el controlador de C en Linux: es el unico de los cuatro
# que no se puede compilar de forma cruzada desde macOS.
#
# El reloj de pared que sale de aqui NO es comparable con el de macOS: el
# contenedor corre sobre una maquina virtual con su propio nucleo. Lo que si es
# solido es el tiempo de `init` por paquete, porque es reloj dentro del proceso y
# no depende del fork ni del exec, y lo son las comparaciones entre los cuatro
# controladores, porque los cuatro se miden en el mismo contenedor.

set -uo pipefail

cd "$(dirname "$0")"
ROOT="$PWD"
RESULTS="$ROOT/results"
IMAGE="${IMAGE:-golang:1.27-alpine}"

mkdir -p "$RESULTS"

if ! docker version >/dev/null 2>&1; then
  echo "linux.sh: no hay un Docker que responda, y sin el no se puede medir en Linux" >&2
  exit 1
fi

{
  echo "# Medido dentro de un contenedor $IMAGE, sobre Dory en un Apple M3 Max."
  echo "# Los cuatro controladores se compilan y se miden en el mismo contenedor."
  echo
  docker run --rm -e "RUNS=${RUNS:-200}" -v "$ROOT:/bench" -w /bench "$IMAGE" sh ./linux-inside.sh
} 2>&1 | tee "$RESULTS/10-linux.txt"
