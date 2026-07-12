// Package slicer cuts tall images into shorter horizontal slices sized for
// OCR. Cuts land only in visually blank rows, so lines of text are never
// severed.
//
// Analysis uses the luminance channel alone. Each row is compared with the
// row below it; a row whose every pixel matches the row below within a small
// contrast tolerance is quiet, and a run of quiet rows forms a gap — the
// blank space between lines of text. The Nth cut
// is placed at the middle of the biggest gap inside its search window,
// [N*TargetHeight - (TargetHeight - MinHeight), N*TargetHeight]. The biggest
// gap in a window is usually a paragraph break, so paragraphs tend to stay
// whole. A window with no gap is cut at its bottom boundary, and a cut that
// would land at or past the bottom of the image is dropped.
//
// Both heights are soft limits. A cut may land anywhere in its window, and
// the shortfall carries into the next slice, so a slice can be as tall as
// TargetHeight + (TargetHeight - MinHeight); the final slice may be shorter
// than MinHeight.
//
// Basic usage:
//
//	slc := slicer.New(500, 400) // target height, min height (pixels)
//	slices := slc.Slice(img)
package slicer
