#!/bin/bash
# Mide lo que cuesta el `init` de cada paquete antes de que main empiece, con el
# GODEBUG=inittrace=1 que trae el runtime de Go.
#
# Lo lanza run.sh, y tambien vale a mano. Existe porque el reloj de pared no dice
# donde se va el tiempo cuando se va antes de main: el medidor de fases de
# internal/phase arranca dentro de main y no puede ver nada de lo que pasa antes.
#
# Un `init` caro se paga en cada una de las muchas invocaciones que un agente hace
# a lo largo de una sesion, y no hay ningun flag del programa que lo evite, asi
# que es la clase de coste que decide una eleccion.
#
# Las columnas que imprime son: milisegundos que tarda ese init, el paquete, y la
# memoria que reserva mientras lo hace, que es lo que explica por que tarda.

set -uo pipefail

cd "$(dirname "$0")"
BIN="$PWD/bin"
WORK="$PWD/work"
TOP="${TOP:-4}"

for d in mattn modernc ncruces zombiezen; do
  echo "--- ls-$d"
  GODEBUG=inittrace=1 "$BIN/ls-$d" "$WORK/board.db" 2>&1 >/dev/null \
    | grep 'ms clock' \
    | awk '{
        pkg = $2
        clock = $5
        # La linea tiene la forma:
        #   init <paquete> @<t> ms, <clock> ms clock, <n> bytes, <m> allocs
        bytes = $8
        allocs = $10
        printf "%s\t%s\t%s\t%s\n", clock, pkg, bytes, allocs
      }' \
    | sort -rn \
    | head -"$TOP" \
    | awk -F'\t' '{ printf "    %8s ms  %-44s %12s bytes  %8s allocs\n", $1, $2, $3, $4 }'
  total=$(GODEBUG=inittrace=1 "$BIN/ls-$d" "$WORK/board.db" 2>&1 >/dev/null \
    | grep 'ms clock' \
    | awk '{ s += $5 } END { printf "%.3f", s }')
  echo "    suma de todos los init: $total ms"
done
