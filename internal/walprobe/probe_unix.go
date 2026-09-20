//go:build unix

package walprobe

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// probeSize is how much the test file holds: 64 KiB of zeros, written and
// synced before anything else is tried.
const probeSize = 64 * 1024

// pattern is the eight bytes the mapping writes and an ordinary read has to
// see afterwards.
var pattern = []byte{0xb1, 0x50, 0x01, 0xde, 0xad, 0xbe, 0xef, 0x01}

// probe runs the three functional steps of
// docs/spec/cmd/doctor.md#el-sondeo-del-sistema-de-ficheros in the board's
// own directory, because what is being asked about is the filesystem of
// that directory and not of a temporary one somewhere else.
func probe(dir string) bool {
	path := filepath.Join(dir, fmt.Sprintf(".biso-wal-probe-%d-%08x", os.Getpid(), rand.Uint32()))
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		// Not being able to write there at all is itself a filesystem
		// where WAL cannot work, so the probe stops here.
		return false
	}
	defer os.Remove(path)
	defer f.Close()

	if _, err := f.Write(make([]byte, probeSize)); err != nil {
		return false
	}
	if err := f.Sync(); err != nil {
		return false
	}
	if !locksAreEnforced(path, f) {
		return false
	}
	return mappingIsShared(f)
}

// locksAreEnforced takes an exclusive lock on the first byte from one
// descriptor and checks that a second, independent descriptor cannot take
// the same one. A filesystem that grants it is not enforcing the lock,
// which is the classic symptom of an NFS mount locking locally without
// coordinating with the other end.
func locksAreEnforced(path string, f *os.File) bool {
	held := unix.Flock_t{Type: unix.F_WRLCK, Whence: 0, Start: 0, Len: 1}
	if err := unix.FcntlFlock(f.Fd(), unix.F_SETLK, &held); err != nil {
		// ENOLCK, ENOSYS or EINVAL: the lock is not there to be had.
		return false
	}
	defer func() {
		release := unix.Flock_t{Type: unix.F_UNLCK, Whence: 0, Start: 0, Len: 1}
		_ = unix.FcntlFlock(f.Fd(), unix.F_SETLK, &release)
	}()

	second, err := os.OpenFile(path, os.O_RDWR, 0o600)
	if err != nil {
		return false
	}
	defer second.Close()

	conflicting := unix.Flock_t{Type: unix.F_WRLCK, Whence: 0, Start: 0, Len: 1}
	err = unix.FcntlFlock(second.Fd(), unix.F_SETLK, &conflicting)
	if err == nil {
		// It was granted, so the first lock is not being enforced.
		release := unix.Flock_t{Type: unix.F_UNLCK, Whence: 0, Start: 0, Len: 1}
		_ = unix.FcntlFlock(second.Fd(), unix.F_SETLK, &release)
		return false
	}
	// Refused is the answer a real exclusive lock gives; any other error
	// is fcntl failing, which counts as a failure too.
	return err == unix.EAGAIN || err == unix.EACCES
}

// mappingIsShared writes a pattern through a shared mapping, syncs it, and
// reads the same bytes back with an ordinary read of the file. A MAP_SHARED
// that does not really share memory with the file gives itself away here.
func mappingIsShared(f *os.File) bool {
	data, err := unix.Mmap(int(f.Fd()), 0, probeSize,
		unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		// ENODEV, ENOTSUP or EOPNOTSUPP: no shared mapping here.
		return false
	}
	defer func() { _ = unix.Munmap(data) }()

	copy(data, pattern)
	if err := unix.Msync(data, unix.MS_SYNC); err != nil {
		return false
	}
	read := make([]byte, len(pattern))
	if _, err := f.ReadAt(read, 0); err != nil {
		return false
	}
	return bytes.Equal(read, pattern)
}
