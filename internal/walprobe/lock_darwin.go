//go:build darwin

package walprobe

// ofdSetLock is macOS's F_OFD_SETLK, from <sys/fcntl.h>, where it has been
// defined as 90 since OS X 10.10. It is written out here because x/sys/unix
// generates its darwin constants from a list that leaves the three F_OFD_*
// commands out, and the number is part of the kernel's stable interface, not
// something this program chooses.
const ofdSetLock = 90
