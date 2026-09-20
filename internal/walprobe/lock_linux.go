//go:build linux

package walprobe

import "golang.org/x/sys/unix"

// ofdSetLock is the fcntl command that takes a byte-range lock bound to the
// open file description instead of to the process, which is what makes the
// enforcement step of the probe answerable from one process at all (see
// probe_unix.go). Linux has had it since 3.15 and x/sys/unix carries the
// constant.
const ofdSetLock = unix.F_OFD_SETLK
