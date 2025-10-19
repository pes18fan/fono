package termimg

import (
	"math"
	"os"

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
