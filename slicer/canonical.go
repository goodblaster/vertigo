package slicer

import (
	"image"
	"image/draw"
)

// canonical converts img to the form the analysis works on. That form is an
// implementation detail; currently it is a grayscale (luminance) plane with
// zero-origin bounds, so row comparisons read one byte per pixel.
func canonical(img image.Image) *image.Gray {
	if g, ok := img.(*image.Gray); ok && g.Rect.Min == (image.Point{}) {
		return g
	}

	g := image.NewGray(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(g, g.Bounds(), img, img.Bounds().Min, draw.Src)
	return g
}
