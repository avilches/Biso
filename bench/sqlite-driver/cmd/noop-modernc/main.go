// noop-modernc carga el controlador traducido a Go y no lo usa. Es el binario
// que ensena el peaje que ese controlador cobra antes de llegar a main.
package main

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
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
