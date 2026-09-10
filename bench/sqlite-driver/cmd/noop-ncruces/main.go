// noop-ncruces carga el controlador de WebAssembly y no lo usa. Aqui se ve si
// preparar el modulo de wazero cuesta algo antes de llegar a main, o si ese coste
// se paga mas tarde, al abrir la primera conexion.
package main

import (
	"database/sql"
	"os"

	_ "github.com/ncruces/go-sqlite3/driver"
)

// main no consulta nada. Lo unico que hace es mirar la lista de controladores
// registrados, y eso hace falta: sin una referencia de verdad al paquete, el
// enlazador se lleva por delante un import en blanco que nadie usa y el binario
// dejaria de medir lo que dice medir. Se comprueba mirando el tamano: si sale
// parecido al de noop-bare, el controlador no esta dentro.
func main() {
	if len(sql.Drivers()) == 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
