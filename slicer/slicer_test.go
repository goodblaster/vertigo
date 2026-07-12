package slicer_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/goodblaster/vertigo/slicer"
)

// textImage builds a synthetic page: 12px lines of "text" every 25px on a
// white background. Each blankBand is a [start, end) row range left entirely
// blank, standing in for a paragraph break.
func textImage(w, h int, blankBands ...[2]int) *image.RGBA {
	inBand := func(y int) bool {
		for _, b := range blankBands {
			if y >= b[0] && y < b[1] {
				return true
			}
		}
		return false
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			c := color.RGBA{R: 255, G: 255, B: 255, A: 255}
			if !inBand(y) && y%25 < 12 && x%7 != 0 {
				c = color.RGBA{R: 20, G: 20, B: 20, A: 255}
			}
			img.Set(x, y, c)
		}
	}
	return img
}

func TestCutsProperties(t *testing.T) {
	slc := slicer.New(500, 400)
	img := textImage(800, 2350)

	cuts := slc.Cuts(img)
	if len(cuts) == 0 {
		t.Fatal("expected cuts for an image taller than TargetHeight")
	}

	maxHeight := slc.TargetHeight + (slc.TargetHeight - slc.MinHeight)
	prev := 0
	for i, cut := range cuts {
		if cut <= prev {
			t.Fatalf("cuts not strictly increasing: %v", cuts)
		}
		if cut >= img.Bounds().Dy() {
			t.Errorf("cut %d lies at or beyond the image height %d", cut, img.Bounds().Dy())
		}
		if h := cut - prev; h < slc.MinHeight || h > maxHeight {
			t.Errorf("slice %d is %dpx tall, want within [%d, %d]", i, h, slc.MinHeight, maxHeight)
		}
		prev = cut
	}
	if rem := img.Bounds().Dy() - prev; rem < 1 || rem > maxHeight {
		t.Errorf("final slice is %dpx tall, want within [1, %d]", rem, maxHeight)
	}
}

func TestCutsShortImage(t *testing.T) {
	slc := slicer.Default()
	img := textImage(800, slc.TargetHeight)

	if cuts := slc.Cuts(img); cuts != nil {
		t.Errorf("Cuts() = %v for an image within TargetHeight, want nil", cuts)
	}

	slices := slc.Slice(img)
	if len(slices) != 1 {
		t.Fatalf("Slice() returned %d slices for an image within TargetHeight, want 1", len(slices))
	}
	if !slices[0].Bounds().Eq(img.Bounds()) {
		t.Errorf("single slice bounds = %v, want %v", slices[0].Bounds(), img.Bounds())
	}
}

func TestCutsPreferParagraphBreaks(t *testing.T) {
	// Blank bands inside the first two search windows ([400,500) and [900,1000)).
	para1 := [2]int{420, 450}
	para2 := [2]int{920, 950}
	img := textImage(800, 1500, para1, para2)

	cuts := slicer.New(500, 400).Cuts(img)
	if len(cuts) != 3 {
		t.Fatalf("Cuts() = %v, want 3 cuts for a 1500px image", cuts)
	}
	if cuts[0] < para1[0] || cuts[0] >= para1[1] {
		t.Errorf("cut %d missed the paragraph break %v", cuts[0], para1)
	}
	if cuts[1] < para2[0] || cuts[1] >= para2[1] {
		t.Errorf("cut %d missed the paragraph break %v", cuts[1], para2)
	}
}

func TestCutsNonDivisibleHeights(t *testing.T) {
	// 500 - 300 = 200 does not divide 500; the window arithmetic once
	// misaligned for such heights and missed gaps near window tops.
	para := [2]int{440, 470} // inside the first search window (300, 500)
	img := textImage(800, 1000, para)

	cuts := slicer.New(500, 300).Cuts(img)
	if len(cuts) == 0 || cuts[0] < para[0] || cuts[0] >= para[1] {
		t.Errorf("cuts = %v, want the first cut inside the paragraph break %v", cuts, para)
	}
	for _, cut := range cuts {
		if cut >= img.Bounds().Dy() {
			t.Errorf("cut %d lies at or beyond the image height %d", cut, img.Bounds().Dy())
		}
	}
}

func TestSliceHeightsSumToImageHeight(t *testing.T) {
	img := textImage(800, 2350, [2]int{430, 460}, [2]int{1410, 1440})
	slices := slicer.Default().Slice(img)

	sum := 0
	for _, s := range slices {
		sum += s.Bounds().Dy()
	}
	if sum != img.Bounds().Dy() {
		t.Errorf("slice heights sum to %d, want %d", sum, img.Bounds().Dy())
	}
}

func TestZeroValueSlicerIsUsable(t *testing.T) {
	var slc slicer.Slicer
	cuts := slc.Cuts(textImage(800, 1500))
	if len(cuts) == 0 {
		t.Error("zero-value Slicer produced no cuts; want default heights to apply")
	}
}

func TestNewSanitizesInputs(t *testing.T) {
	if slc := slicer.New(0, 0); slc.TargetHeight != slicer.DefaultTargetHeight || slc.MinHeight != slicer.DefaultMinHeight {
		t.Errorf("New(0, 0) = %+v, want defaults", slc)
	}
	if slc := slicer.New(1000, 2000); slc.MinHeight >= slc.TargetHeight {
		t.Errorf("New(1000, 2000) = %+v, want MinHeight below TargetHeight", slc)
	}
}

// opaqueImage hides the concrete image type so it exposes no SubImage
// method, forcing Slice down the copy fallback path.
type opaqueImage struct {
	image.Image
}

func TestSliceWithoutSubImageMethod(t *testing.T) {
	img := textImage(800, 1200)
	slices := slicer.Default().Slice(opaqueImage{img})

	sum := 0
	for _, s := range slices {
		if s.Bounds().Dx() != img.Bounds().Dx() {
			t.Errorf("slice width = %d, want %d", s.Bounds().Dx(), img.Bounds().Dx())
		}
		sum += s.Bounds().Dy()
	}
	if sum != img.Bounds().Dy() {
		t.Errorf("slice heights sum to %d, want %d", sum, img.Bounds().Dy())
	}
}

func TestSliceAt(t *testing.T) {
	img := textImage(800, 1000)

	slices := slicer.SliceAt(img, []int{300, 700})
	if len(slices) != 3 {
		t.Fatalf("SliceAt() returned %d slices, want 3", len(slices))
	}
	for i, want := range []int{300, 400, 300} {
		if got := slices[i].Bounds().Dy(); got != want {
			t.Errorf("slice %d is %dpx tall, want %d", i, got, want)
		}
	}

	if slices := slicer.SliceAt(img, nil); len(slices) != 1 || !slices[0].Bounds().Eq(img.Bounds()) {
		t.Errorf("SliceAt() with no cuts should return the whole image as one slice")
	}
}

func BenchmarkCuts(b *testing.B) {
	img := textImage(1600, 3000, [2]int{430, 460}, [2]int{1410, 1440})
	slc := slicer.Default()

	b.ReportAllocs()
	for b.Loop() {
		slc.Cuts(img)
	}
}
