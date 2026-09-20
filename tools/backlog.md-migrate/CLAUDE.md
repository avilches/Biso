# backlog.md-migrate

Importador y exportador entre Backlog.md y `biso`. Lee [`README.md`](README.md) para saber qué hace,
[`docs/especificacion.md`](docs/especificacion.md) para el comportamiento exacto y
[`docs/decisiones.md`](docs/decisiones.md) antes de cambiar una regla que parezca arbitraria.

## Es un proyecto aparte

Vive en `tools/backlog.md-migrate/` del repositorio de `biso` por comodidad, y **no hereda nada de
`biso`**:

- Tiene su propio módulo Go (`go.mod` en esta carpeta). El `go vet ./...` y el `go test ./...` de la
  raíz no entran aquí, y la dependencia de YAML no toca el `go.mod` de `biso`.
- **No importa ningún paquete de `internal/` de `biso`.** Habla con `biso` solo como lo haría
  cualquier usuario: ejecutando el binario (`biso config get`, `biso ls`, `biso new --from`). Si
  necesita una regla de `biso` (el algoritmo de coincidencia de vocabularios, por ejemplo), la
  reimplementa y la describe en su propia especificación.
- Su documentación no está en `docs/` de `biso`, no se publica en su sitio de MkDocs y `docs-doctor`
  no la comprueba. Al revés, la especificación de `biso` no se modifica para acomodar a esta utilidad.
- **Una pega conocida:** el `make fmt-check` de la raíz ejecuta `gofmt -l .` sobre todo el árbol y por
  tanto también ve estos ficheros. Hay que dejarlos formateados con `gofmt`.
- Sus pruebas que necesiten `biso` usan el binario `bin/biso` que compila el `Makefile` de la raíz,
  y se saltan con un mensaje claro si no existe.

## Es una herramienta general

No está hecha para ningún tablero concreto. Que se use con el tablero de Backlog.md de este proyecto
es circunstancial, y **ninguna decisión puede apoyarse en lo que ese tablero contenga o deje de
contener**. Que ninguna tarea real use un campo (`due_date`, los comentarios, la definición de hecho)
no es un motivo para no cubrirlo: la herramienta cubre lo que Backlog.md puede guardar, y eso se mide
creando tareas con el CLI de Backlog.md, no mirando un tablero. El tablero de este proyecto sirve como
una prueba real más, no como el modelo.

## Reglas

- **La documentación va en español**, el código y todo lo que imprime la herramienta (ayuda, mensajes,
  informes) en inglés, sin una palabra en español en un `.go`.
- **Nunca em-dash**, en ningún texto.
- **Nada se pierde en silencio.** Lo que no se puede mapear se informa, con el fichero y el campo. Un
  campo o una sección de Backlog.md que el convertidor no reconoce es un hallazgo, no algo que se
  ignora.
- **Si al implementar aparece un caso que la especificación no cubre, primero se añade a la
  especificación** y luego se implementa.
- Los mensajes de commit y las descripciones de PR no llevan coautoría ni mención de haber sido
  generados por un agente.
