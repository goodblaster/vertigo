package slicer

import (
	"image"
	"image/color"
	"reflect"
	"testing"
)

func TestFindGaps(t *testing.T) {
	// A 10x100 grayscale image whose rows alternate black and white — every
	// row visually differs from the next — except for a uniform band at rows
	// 40–49. With heights 50/40 the search windows are rows 40–49 and 90–99,
	// so the band should surface as the only gap.
	img := image.NewGray(image.Rect(0, 0, 10, 100))
	for y := range 100 {
		v := uint8(0)
		if y%2 == 0 {
			v = 255
		}
		if y >= 40 && y < 50 {
			v = 200
		}
		for x := range 10 {
			img.SetGray(x, y, color.Gray{Y: v})
		}
	}

	got := New(50, 40).findGaps(img)
	want := []gap{{middle: 44, size: 9}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("findGaps() = %v, want %v", got, want)
	}
}

func TestGroupQuietRows(t *testing.T) {
	got := groupQuietRows([]int{5, 6, 7, 30, 50, 51})
	want := []gap{
		{middle: 6, size: 3},
		{middle: 30, size: 1},
		{middle: 50, size: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("groupQuietRows() = %v, want %v", got, want)
	}
}

func TestFilterCloseGaps(t *testing.T) {
	in := []gap{
		{middle: 100, size: 2},
		{middle: 110, size: 5}, // within 20 rows of the first and larger, so it wins
		{middle: 200, size: 1},
	}
	want := []gap{
		{middle: 110, size: 5},
		{middle: 200, size: 1},
	}
	if got := filterCloseGaps(in, 20); !reflect.DeepEqual(got, want) {
		t.Errorf("filterCloseGaps() = %v, want %v", got, want)
	}
}
