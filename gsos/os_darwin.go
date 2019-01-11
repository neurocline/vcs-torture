// vcs-torture/os_windows.go
// -- Mac OS X-specific code for vcs-torture

// +build darwin

package gsos

// #include <mach/mach_time.h>
import "C"

import (
	"time"
)

// ----------------------------------------------------------------------------------------------
// Mach (Mac OS X) timing code (adapted from https://github.com/ScaleFT/monotime)
// We don't actually need this, because the Go library has high-resolution time for
// some operating systems. But because the Go library lacks high-resolution time for Windows,
// we have this so that our code is consistent.
// Currently, we make no attempt to implement time.Time-compatible values.
//
// Get rid of this code once the Go library supports nanosecond-level timing for all
// relevant operating systems.

// hiresTimestamp is a high-resolution time counter. On Windows 10, this has a resolution of
// 1 to 20 nanoseconds.
type HighresTimestamp uint64

// getHiresTimestamp returns the current time as a HighresTimestamp
func HighresTime() HighresTimestamp {
	return HighresTimestamp(C.mach_absolute_time())
}

// HighresTimestamp.Duration() converts a HighresTimestamp into a time.Duration value
// (this looks horrible, but matches the Mach library code)
func (t HighresTimestamp) Duration() time.Duration {
	return time.Duration(uint64(t) * uint64(tbinfo.numer) / uint64(tbinfo.denom))
}

// ------------------------

var tbinfo C.struct_mach_timebase_info

func init() {
	C.mach_timebase_info(&tbinfo)
}
