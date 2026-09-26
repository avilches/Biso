# backlog.md-migrate

Importa un tablero de [Backlog.md](https://github.com/MrLesk/Backlog.md) a un tablero de `biso`, y
(más adelante) exporta un tablero de `biso` a Backlog.md. Son dos órdenes independientes, una
dirección cada vez: no es un sincronizador.

```
backlog.md-migrate import <backlog-dir> --project <dir> [--out <file>] [--biso <path>] [--strict]
backlog.md-migrate export ...        # por diseñar, ver TASK-7
```

**Estado: sin implementar.** La especificación y las decisiones se repasaron entera con quien encarga el
proyecto, y ese repaso ya terminó: las propuestas de [`docs/pendientes.md`](docs/pendientes.md) quedan
ratificadas sin ninguna decisión pendiente, TASK-73 (los ids de subtarea con punto) incluida, que se
resolvió en el sentido de que `biso` no adopta esa forma, así que este proyecto decidió su propio
mecanismo para conservarla (ver [`docs/decisiones.md`](docs/decisiones.md)). Lo que queda antes de
escribir código no es ya ninguna decisión: son las mediciones pendientes de la sección D de
[`docs/pendientes.md`](docs/pendientes.md) y el diseño completo de la exportación (TASK-7).

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
