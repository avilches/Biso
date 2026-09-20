//go:build !race

package main

// raceDetector says whether this build carries the race detector, which is
// what decides whether the startup budget is asserted or skipped.
const raceDetector = false
