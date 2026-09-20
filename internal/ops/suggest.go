package ops

import "biso/internal/match"

// Suggest is the "did you mean" algorithm of
// docs/spec/vocabularios.md#el-algoritmo-de-sugerencias-más-parecidas offered
// to the layer above this one.
//
// `biso help` is the one place where the vocabulary being suggested from is
// not a board's but this program's own: the list of commands, which lives in
// internal/cli because the help texts do. The layer rule of section 3 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md says
// internal/cli reaches internal/match through here and never around it, and
// the error of a command name that does not exist is not reason enough to
// make the one exception, so the algorithm comes to it instead.
func Suggest(value string, candidates []string, n int) []string {
	return match.Suggest(value, candidates, n)
}
