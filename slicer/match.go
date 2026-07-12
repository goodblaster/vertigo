package slicer

// contrastThreshold is the smallest luminance difference (out of 255) that
// counts as a visual change between vertically adjacent pixels. Smaller
// differences are treated as noise — scan grain, compression artifacts,
// anti-aliasing fringes — rather than content.
const contrastThreshold = 16

// match reports whether two vertically adjacent rows are visually identical:
// no pixel pair differs by more than contrastThreshold. A row that matches
// the row below it is quiet — blank space between lines of text.
func match(rowA, rowB []byte) bool {
	for i := range rowA {
		d := int(rowA[i]) - int(rowB[i])
		if d < -contrastThreshold || d > contrastThreshold {
			return false
		}
	}
	return true
}
