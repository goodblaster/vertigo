package slicer

import "image"

// Heights (in pixels) used when none are configured.
const (
	DefaultTargetHeight = 500
	DefaultMinHeight    = 400
)

// Slicer chooses where to cut tall images. The zero value uses the default
// heights; use New to pick your own.
type Slicer struct {
	// TargetHeight is the preferred slice height. It is a soft limit: a cut
	// may land above its boundary, and the shortfall carries into the next
	// slice, so a slice can be as tall as
	// TargetHeight + (TargetHeight - MinHeight).
	TargetHeight int

	// MinHeight is the shortest a slice should be, also a soft limit: the
	// final slice may be shorter, though no cut is made once the remaining
	// rows fit in one slice. The rows between MinHeight and TargetHeight form
	// the window searched for a clean cut.
	MinHeight int
}

// New returns a Slicer with the given heights. Values that are zero,
// negative, or out of order fall back to defaults, so the result is always
// usable.
func New(targetHeight, minHeight int) Slicer {
	if targetHeight <= 0 {
		targetHeight = DefaultTargetHeight
	}
	if minHeight <= 0 || minHeight >= targetHeight {
		minHeight = targetHeight * DefaultMinHeight / DefaultTargetHeight
	}
	return Slicer{
		TargetHeight: targetHeight,
		MinHeight:    minHeight,
	}
}

// Default returns a Slicer with the default heights.
func Default() Slicer {
	return New(DefaultTargetHeight, DefaultMinHeight)
}

// windowHeight is the size of the search window: how far above each
// TargetHeight boundary a cut may land.
func (slc Slicer) windowHeight() int {
	return slc.TargetHeight - slc.MinHeight
}

// maxHeight is the tallest slice the heights allow: a full TargetHeight plus
// the shortfall of a cut placed at the very top of its window.
func (slc Slicer) maxHeight() int {
	return slc.TargetHeight + slc.windowHeight()
}

// Cuts returns the rows at which img should be cut, counted from its top
// edge, low to high, every one strictly inside the image. It returns nil if
// the image already fits in a single slice, which may be as tall as
// TargetHeight + (TargetHeight - MinHeight).
func (slc Slicer) Cuts(img image.Image) []int {
	// Sanitize so a zero-value or hand-built Slicer behaves like one from New.
	slc = New(slc.TargetHeight, slc.MinHeight)

	height := img.Bounds().Dy()
	if height <= slc.maxHeight() {
		return nil
	}

	return slc.chooseCuts(height, slc.findGaps(canonical(img)))
}

// Slice cuts img into horizontal slices at the rows returned by Cuts. An
// image that already fits in a single slice comes back as one.
func (slc Slicer) Slice(img image.Image) []image.Image {
	return SliceAt(img, slc.Cuts(img))
}

// chooseCuts places one cut per search window, at the middle of the window's
// biggest gap, in an image height rows tall. A window with no gap is cut at
// its bottom boundary. Cutting stops as soon as the rows below the previous
// cut fit in one slice, so the image never ends in a needless sliver.
func (slc Slicer) chooseCuts(height int, gaps []gap) []int {
	var cuts []int

	prev := 0
	for hi := slc.TargetHeight; height-prev > slc.maxHeight(); hi += slc.TargetHeight {
		lo := hi - slc.windowHeight()

		// Ties go to the lower gap, which yields taller slices.
		best := hi
		bestSize := 0
		for _, g := range gaps {
			if g.middle > lo && g.middle < hi && g.size >= bestSize {
				best = g.middle
				bestSize = g.size
			}
		}

		cuts = append(cuts, best)
		prev = best
	}

	return cuts
}
