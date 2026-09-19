package vcs

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"sync"
)

// outcome of launching one program.
type execResult struct {
	// started is false when the program could not be launched at all, which is
	// what "the configured system is not installed" looks like from here.
	started bool
	// exit is the exit code, meaningful only when started is true.
	exit int
	// stdout is the standard output, captured only for questions.
	stdout string
}

// ok reports whether the program ran and ended well.
func (e execResult) ok() bool { return e.started && e.exit == 0 }

// collector gathers the lines that the orders write, from both streams, in the
// order each stream produced them. The two streams interleave at line
// granularity and their relative order is not guaranteed, which is exactly
// what docs/spec/cmd/snapshot.md promises.
type collector struct {
	mu    sync.Mutex
	lines []string
}

// take returns the lines gathered so far.
func (c *collector) take() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.lines...)
}

// stream returns a writer that splits what it receives into lines and adds
// them to the collector. Blank lines are discarded, and a last line without a
// final newline counts as a line once the writer is flushed.
func (c *collector) stream() *lineWriter { return &lineWriter{to: c} }

type lineWriter struct {
	to  *collector
	buf []byte
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		w.emit(w.buf[:i])
		w.buf = w.buf[i+1:]
	}
	return len(p), nil
}

// flush emits whatever is left without a final newline.
func (w *lineWriter) flush() {
	if len(w.buf) > 0 {
		w.emit(w.buf)
		w.buf = nil
	}
}

func (w *lineWriter) emit(line []byte) {
	text := strings.TrimSuffix(string(line), "\r")
	if strings.TrimSpace(text) == "" {
		return
	}
	w.to.mu.Lock()
	defer w.to.mu.Unlock()
	w.to.lines = append(w.to.lines, text)
}

// ask runs one of the questions of a recipe: its output is the answer and is
// never forwarded, and a non-zero exit code is a legitimate answer too.
func ask(dir string, name string, args ...string) execResult {
	var out bytes.Buffer
	cmd := command(dir, name, args)
	cmd.Stdout = &out
	cmd.Stderr = nil
	result := wait(cmd)
	result.stdout = out.String()
	return result
}

// order runs one of the orders of a recipe: what it writes on either stream is
// forwarded, line by line, to the collector.
func order(c *collector, dir string, name string, args ...string) execResult {
	out, errs := c.stream(), c.stream()
	cmd := command(dir, name, args)
	cmd.Stdout = out
	cmd.Stderr = errs
	result := wait(cmd)
	out.flush()
	errs.flush()
	return result
}

// command builds the process. Every question and every order runs with the
// board directory as its working directory and with the standard input closed
// and no terminal, so that an order waiting for a password fails instead of
// hanging forever. There is no timeout, deliberately.
func command(dir string, name string, args []string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = nil
	return cmd
}

func wait(cmd *exec.Cmd) execResult {
	err := cmd.Run()
	if err == nil {
		return execResult{started: true}
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return execResult{started: true, exit: exit.ExitCode()}
	}
	return execResult{}
}
