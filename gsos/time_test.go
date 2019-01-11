// vcs-torture/time_test.go
//
// Some hints on testing
// run tests: "go test ."
// skip long tests: "go test -short ."
// code coverage: "go test . -cover"

package gsos

import (
	"testing"

	"math"
	"time"
)

// TestHighresTime makes sure that the resolution of HighresTime matches
// that of time.Time to within a certain precision. This is a very loose
// test, it's just making sure we didn't screw up horribly.
// This test is not run in -short mode.
func TestHighresTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip HighresTime test in short mode")
	}
	start := time.Now()
	hiStart := HighresTime()
	time.Sleep(500*time.Millisecond)
	hiEnd := HighresTime()
	end := time.Now()

	hiresDeltaMs := (hiEnd - hiStart).Duration().Seconds() * 1000.0
	osTimeDeltaMs := end.Sub(start).Seconds() * 1000.0

	if math.Abs(hiresDeltaMs - osTimeDeltaMs) > 1.0 {
		t.Errorf("HighresTime doesn't match time.Time: got %.2f, want close to %.2f", hiresDeltaMs, osTimeDeltaMs)
	}
}
