// Command backlog.md-migrate imports a Backlog.md board into a biso board
// (and, later, exports the other way around). It hands argv to
// internal/cli and exits with the code that comes back: everything else,
// from parsing the command line to the conversion itself, happens there.
package main

import (
	"os"

	"backlog.md-migrate/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
