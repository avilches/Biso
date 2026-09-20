// Command biso is the task board of a project. It starts up, hands argv to
// internal/cli and exits with the code that comes back: everything else,
// from the analysis of the command line to the text and the JSON of the
// answer, happens there.
package main

import (
	"fmt"
	"os"

	"biso/internal/cli"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		// Without a working directory there is no board to resolve from,
		// and that is the environment failing and not the request: exit
		// code 8 of docs/spec/codigos-de-salida.md.
		fmt.Fprintf(os.Stderr, "error: the working directory cannot be read: %s\n", err)
		os.Exit(8)
	}
	home, _ := os.UserHomeDir()

	os.Exit(cli.Run(os.Args[1:], cli.Streams{
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
		Stdin:            os.Stdin,
		Getenv:           os.LookupEnv,
		Dir:              dir,
		Home:             home,
		StdoutIsTerminal: cli.Terminal(os.Stdout),
		StderrIsTerminal: cli.Terminal(os.Stderr),
	}))
}
