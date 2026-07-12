package slicer

import "image"

// gap is a run of quiet rows: a candidate cut point.
type gap struct {
	middle int // center row of the run
	size   int // rows in the run
}

// gapMergeDistance is how close (in rows) two gaps may sit before only the
// larger of the two is kept.
const gapMergeDistance = 20

// findGaps scans the canonical image and returns every gap that could take a
// cut. Rows outside every search window can never take a cut, so their
// comparisons are skipped — with the default heights, about 80% of the work.
// The first and last rows are never part of a gap.
func (slc Slicer) findGaps(img *image.Gray) []gap {
	stride := img.Stride
	pix := img.Pix

	rect := img.Bounds()
	height := rect.Max.Y - rect.Min.Y
	width := rect.Max.X // stride may exceed width due to row padding

	var quietRows []int
	for y := 1; y < height-2; y++ {
		// A row is inside a search window when it lies past MinHeight within
		// its TargetHeight block.
		if y%slc.TargetHeight < slc.MinHeight {
			continue
		}

		rowA := pix[y*stride : y*stride+width]
		rowB := pix[(y+1)*stride : (y+1)*stride+width]
		if match(rowA, rowB) {
			quietRows = append(quietRows, y)
		}
	}

	return filterCloseGaps(groupQuietRows(quietRows), gapMergeDistance)
}

// groupQuietRows merges each run of adjacent quiet rows into a single gap.
func groupQuietRows(rows []int) []gap {
	if len(rows) == 0 {
		return []gap{}
	}

	var gaps []gap
	start := rows[0]
	count := 1
	for i := 1; i < len(rows); i++ {
		if rows[i] == rows[i-1]+1 {
			count++
			continue
		}
		gaps = append(gaps, gap{middle: (start + rows[i-1]) / 2, size: count})
		start = rows[i]
		count = 1
	}

	return append(gaps, gap{middle: (start + rows[len(rows)-1]) / 2, size: count})
}

// filterCloseGaps keeps only the largest gap in each cluster of gaps lying
// within distance rows of one another. Gaps must be ordered by middle,
// low to high.
func filterCloseGaps(gaps []gap, distance int) []gap {
	if len(gaps) == 0 {
		return gaps
	}

	var filtered []gap
	current := gaps[0]
	for _, g := range gaps[1:] {
		if g.middle-current.middle <= distance {
			if g.size > current.size {
				current = g
			}
			continue
		}
		filtered = append(filtered, current)
		current = g
	}

	return append(filtered, current)
}
