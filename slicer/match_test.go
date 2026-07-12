package slicer

import "testing"

func TestMatch(t *testing.T) {
	base := []byte{100, 100, 100, 100}

	// Differences at or below contrastThreshold are noise, not content.
	noisy := []byte{100, 100 + contrastThreshold, 100 - contrastThreshold, 100}
	if !match(base, noisy) {
		t.Error("rows differing within contrastThreshold should match")
	}

	// A single difference beyond the threshold, in either direction, is content.
	if match(base, []byte{100, 100 + contrastThreshold + 1, 100, 100}) {
		t.Error("rows with a pixel brighter than the threshold allows should not match")
	}
	if match(base, []byte{100, 100 - contrastThreshold - 1, 100, 100}) {
		t.Error("rows with a pixel darker than the threshold allows should not match")
	}

	if !match(base, base) {
		t.Error("identical rows should match")
	}
}
