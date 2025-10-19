package termimg

import (
	"math"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

// Change the dimensions of an image to be based on columns
func (t *TerminalImage) fixImageDimensions() {
	ws, _ := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	termCols := int(ws.Col)

	termPxWidth := int(ws.Xpixel)
	cellWidth := float64(termPxWidth) / float64(termCols)
	t.W /= int(math.Floor(cellWidth))

	termPxHeight := int(ws.Ypixel)
	cellHeight := float64(termPxHeight) / float64(termCols)
	t.H /= int(math.Floor(cellHeight))
}

// Clear the screen using the OS-native command.
// This is necessary to avoid visual glitches caused by showing the images.
func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
