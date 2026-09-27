# backlog.md-migrate

Importa un tablero de [Backlog.md](https://github.com/MrLesk/Backlog.md) a un tablero de `biso`, y
(más adelante) exporta un tablero de `biso` a Backlog.md. Son dos órdenes independientes, una
dirección cada vez: no es un sincronizador.

```
backlog.md-migrate import <backlog-dir> --project <dir> [--out <file>] [--biso <path>] [--strict]
backlog.md-migrate export ...        # por diseñar, ver TASK-7
```

**Estado: implementado.** Las seis fases de `import` están terminadas, revisadas y con sus pruebas de
extremo a extremo pasando contra el binario real de `biso`. La especificación completa está en
[`docs/especificacion.md`](docs/especificacion.md), y el porqué de cada regla, incluidas las que se
precisaron durante la implementación, en [`docs/decisiones.md`](docs/decisiones.md). Lo que queda es
`export`, sin diseñar todavía (TASK-7), y las mediciones sueltas de la sección D de
[`docs/pendientes.md`](docs/pendientes.md).

Es un proyecto aparte, con su propia especificación y sus propias decisiones. Vive dentro del
repositorio de `biso` por comodidad, pero no cuelga de él: ver [`CLAUDE.md`](CLAUDE.md).

## Uso de `import`

El tablero destino se crea antes, con el prefijo y el vocabulario que se quieran. Se ejecuta desde el
directorio del proyecto, donde `biso` deja su fichero puntero:

```
biso init my-project --prefix BISO
```

Después se convierte el tablero de origen y se ensaya la importación sin escribir nada:

```
backlog.md-migrate import ./backlog --project . --out tasks.ndjson
biso new --from tasks.ndjson --dry-run
```

Lo que el convertidor no ha podido mapear, o ha tenido que cambiar (un id ocupado, un estado que el
destino no declara), sale por la salida de errores, una línea por hallazgo con el fichero y el campo.
El código de salida 5 avisa de que hubo hallazgos. Si el ensayo pasa, se repite `biso new --from` sin
`--dry-run`: el lote entra entero o no entra nada.

## Documentación

- [`docs/especificacion.md`](docs/especificacion.md): qué hace cada orden, campo a campo, con sus
  salidas y sus códigos de salida.
- [`docs/decisiones.md`](docs/decisiones.md): por qué se mapea cada cosa como se mapea, y las
  alternativas que se descartaron.
