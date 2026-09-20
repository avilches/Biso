//go:build unix && !linux && !darwin

package walprobe

// ofdSetLock is zero on a Unix that has no open file description locks, such
// as the BSDs. The probe then stops after checking that byte-range locks work
// at all and skips the enforcement step, because a process cannot conflict
// with its own traditional record locks and a probe that asked anyway would
// report every filesystem as unsafe (see probe_unix.go).
const ofdSetLock = 0
