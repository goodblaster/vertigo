package slicer

import (
	"image"
	"image/draw"
)

// SliceAt cuts img into horizontal slices at the given rows, which Slice
// obtains from Cuts. Rows outside the image are clamped, and an empty cut
// list yields the whole image as a single slice.
func SliceAt(img image.Image, cuts []int) []image.Image {
	slices := make([]image.Image, 0, len(cuts)+1)

	last := 0
	for _, cut := range cuts {
		slices = append(slices, subImage(img, image.Rect(0, last, img.Bounds().Max.X, cut)))
		last = cut
	}

	if last < img.Bounds().Max.Y {
		slices = append(slices, subImage(img, image.Rect(0, last, img.Bounds().Max.X, img.Bounds().Max.Y)))
	}

	return slices
}

// subImage returns the portion of img inside rect, sharing pixels when img's
// concrete type provides SubImage and copying when it does not.
func subImage(img image.Image, rect image.Rectangle) image.Image {
	if subImager, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		return subImager.SubImage(rect)
	}

	rect = rect.Intersect(img.Bounds())
	out := image.NewRGBA(rect)
	draw.Draw(out, rect, img, rect.Min, draw.Src)
	return out
}
