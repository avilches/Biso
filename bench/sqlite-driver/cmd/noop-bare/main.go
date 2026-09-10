// noop-bare no importa ningun controlador. Es el suelo contra el que se mide lo
// que cuesta el `init` de cada uno: arrancar un proceso de Go y salir.
package main

import "os"

func main() { os.Exit(0) }
