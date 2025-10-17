package main

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/dolmen-go/kittyimg"
)

type terminalImage struct {
	w    int
	h    int
	data string
}

func cropToSquare(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	size := min(h, w)
	x0 := b.Min.X + (w-size)/2
	y0 := b.Min.Y + (h-size)/2
	rect := image.Rect(x0, y0, x0+size, y0+size)
	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(rect)
}

func getEncodedImage(data []byte) (terminalImage, error) {
	if len(data) == 0 {
		return terminalImage{}, nil
	}

	r := bytes.NewReader(data)
	img, _, err := image.Decode(r)
	if err != nil {
		return terminalImage{}, err
	}
	square := cropToSquare(img)

	var w bytes.Buffer
	kittyimg.Fprint(&w, square)

	return terminalImage{
		w:    img.Bounds().Dx(),
		h:    img.Bounds().Dy(),
		data: w.String(),
	}, nil
}
