// Package walprobe answers whether a directory is on a filesystem where
// SQLite's WAL mode is safe, by exercising the two primitives that mode
// needs rather than by recognizing the filesystem's name
// (docs/spec/cmd/doctor.md#el-sondeo-del-sistema-de-ficheros).
//
// Recognizing names would fall short of a new FUSE filesystem or a network
// disk that announces itself with an ordinary type, so the probe takes a
// byte-range lock and a shared memory mapping in the very directory being
// asked about, and reports whether both behaved.
//
// Only `biso doctor` calls it: it writes a file and waits up to two
// seconds, which no command on the startup budget can pay for.
package walprobe

import (
	"context"
	"time"
)

// Timeout is the single deadline the three functional steps run inside. No
// system call the probe makes has a timeout of its own, and a hung network
// mount blocking in fcntl is the case this covers, so the probe runs in a
// goroutine of its own and the deadline is waited on instead.
const Timeout = 2 * time.Second

// Safe answers whether WAL mode is safe in dir. Anything that fails, is
// unsupported or does not finish in time answers false, because every one
// of those is a filesystem where the mode cannot be relied on.
//
// A false is never a claim that the board is damaged: it is the cheapest
// signal there is that the guarantees SQLite needs are not all there, which
// is why `biso doctor` reports it as a warning and not as an error.
func Safe(dir string) bool {
	done := make(chan bool, 1)
	go func() {
		// The goroutine is abandoned if the deadline runs out, because Go
		// cannot cancel a system call that is blocked; it cleans its own
		// test file up if it ever returns, and nothing here waits for
		// that.
		done <- probe(dir)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	select {
	case ok := <-done:
		return ok
	case <-ctx.Done():
		return false
	}
}
