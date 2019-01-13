// vcs-torture/gsos/console_windows.go

// +build windows

package gsos

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Note: All DLL loads and proc lookups are in os_windows.go

// smallRect defines a rectangle with 16-bit coordinates.
// See https://docs.microsoft.com/en-us/windows/console/small-rect-str
type smallRect struct {
	Left, Top, Right, Bottom int16
}

// coordinates is the xy coordinates of a character cell in a console
// screen buffer; (0,0) is the top-left cell of the buffer.
// See https://docs.microsoft.com/en-us/windows/console/coord-str
type coordinates struct {
	X, Y int16
}

type word int16

// consoleScreenBufferInfo holds console screen buffer information.
// See https://docs.microsoft.com/en-us/windows/console/console-screen-buffer-info-str
type consoleScreenBufferInfo struct {
	dwSize              coordinates
	dwCursorPosition    coordinates
	wAttributes         word
	srWindow            smallRect
	dwMaximumWindowSize coordinates
}

// TerminalWidth returns the width of the terminal (uses 0 as error return value)
func TerminalWidth() int {
	var info consoleScreenBufferInfo
	r1, _, _ := procGetConsoleScreenBufferInfo.Call(uintptr(windows.Stdout), uintptr(unsafe.Pointer(&info)))
	if r1 == 0 {
		return 0 // something went wrong
	}
	return int(info.dwSize.X)
}
