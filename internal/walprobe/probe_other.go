//go:build !unix

package walprobe

// probe answers on a platform this build has no probe for. Saying the
// filesystem is safe is what keeps `biso doctor` from reporting a risk it
// did not measure: the warning of
// docs/spec/cmd/doctor.md#el-sondeo-del-sistema-de-ficheros means the probe
// ran and something failed, and here it never ran.
func probe(string) bool { return true }
