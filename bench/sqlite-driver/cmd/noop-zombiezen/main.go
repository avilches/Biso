// noop-zombiezen carga la envoltura de zombiezen y no la usa. No puede mirar la
// lista de controladores de database/sql, porque este controlador no se registra
// ahi, asi que referencia una constante suya para forzar el enlazado.
package main

import (
	"os"

	"zombiezen.com/go/sqlite"
)

func main() {
	if sqlite.OpenReadWrite == 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
