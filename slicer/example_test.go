package slicer_test

import (
	"fmt"
	"image"

	"github.com/goodblaster/vertigo/slicer"
)

func ExampleSlicer_Slice() {
	// A blank 800x1200 image; real inputs would come from image.Decode.
	img := image.NewRGBA(image.Rect(0, 0, 800, 1200))

	slc := slicer.New(500, 400) // target height, min height (pixels)
	slices := slc.Slice(img)

	fmt.Println(len(slices))
	// Output: 3
}
