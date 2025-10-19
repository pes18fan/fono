package termimg

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32               = windows.NewLazyDLL("kernel32.dll")
	procGetCurrentConsoleFont = modkernel32.NewProc("GetCurrentConsoleFont")
)

type consoleFontInfo struct {
	nFont      uint32
	dwFontSize windows.Coord
}

// Change the dimensions of an image to be based on columns
func (t *TerminalImage) fixImageDimensions() {
	handle := windows.Handle(os.Stdout.Fd())

	var csbi windows.ConsoleScreenBufferInfo
	windows.GetConsoleScreenBufferInfo(handle, &csbi)

	cols := int(csbi.Window.Right - csbi.Window.Left + 1)
	rows := int(csbi.Window.Top - csbi.Window.Bottom + 1)

	var cfi consoleFontInfo
	procGetCurrentConsoleFont.Call(uintptr(handle), 0, uintptr(unsafe.Pointer(&cfi)))

	t.W = cols * int(cfi.dwFontSize.X)
	t.H = rows * int(cfi.dwFontSize.Y)
}
