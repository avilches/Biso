//go:build !race

package board

// raceDetector says whether this build carries the race detector, which
// is what decides which of the two limits of budget_test.go applies.
const raceDetector = false
