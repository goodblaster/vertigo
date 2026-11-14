package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goodblaster/vertigo/slicer"
)

// writePage saves a synthetic 200x1500 page of "text" lines as a PNG.
func writePage(t *testing.T, name string) {
	t.Helper()

	img := image.NewGray(image.Rect(0, 0, 200, 1500))
	for y := range 1500 {
		for x := range 200 {
			c := color.Gray{Y: 255}
			if y%25 < 12 && x%7 != 0 {
				c = color.Gray{Y: 20}
			}
			img.SetGray(x, y, c)
		}
	}

	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func newJob(stdout, stderr *bytes.Buffer) *job {
	return &job{slc: slicer.Default(), format: "png", stdout: stdout, stderr: stderr}
}

func TestSliceFileWritesNextToInput(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "page.png")
	writePage(t, in)

	var stdout, stderr bytes.Buffer
	if err := newJob(&stdout, &stderr).sliceFile(in); err != nil {
		t.Fatal(err)
	}

	names := strings.Fields(stdout.String())
	if len(names) != 3 {
		t.Fatalf("printed %d filenames, want 3: %q", len(names), stdout.String())
	}
	total := 0
	for i, name := range names {
		if want := filepath.Join(dir, "page-00"+string(rune('0'+i))+".png"); name != want {
			t.Errorf("slice %d named %s, want %s", i, name, want)
		}
		img, err := loadImage(name)
		if err != nil {
			t.Fatal(err)
		}
		total += img.Bounds().Dy()
	}
	if total != 1500 {
		t.Errorf("slice heights sum to %d, want 1500", total)
	}
}

func TestSliceFileCreatesOutDir(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "page.png")
	writePage(t, in)

	var stdout, stderr bytes.Buffer
	j := newJob(&stdout, &stderr)
	j.outDir = filepath.Join(dir, "nested", "out")
	if err := j.sliceFile(in); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(j.outDir, "page-000.png")); err != nil {
		t.Errorf("slice not written to the new output directory: %v", err)
	}
}

func TestSliceFileRefusesToOverwriteAnotherInput(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a", "page.png")
	b := filepath.Join(dir, "b", "page.png")
	writePage(t, a)
	writePage(t, b)

	var stdout, stderr bytes.Buffer
	j := newJob(&stdout, &stderr)
	j.outDir = filepath.Join(dir, "out")
	if err := j.sliceFile(a); err != nil {
		t.Fatal(err)
	}
	printed := stdout.String()

	err := j.sliceFile(b)
	if err == nil {
		t.Fatal("second input with the same base name was sliced; want an error")
	}
	if !strings.Contains(err.Error(), a) {
		t.Errorf("error %q does not name the input it collides with", err)
	}
	if stdout.String() != printed {
		t.Errorf("colliding input still printed filenames: %q", stdout.String()[len(printed):])
	}
}

func TestSliceFileCutsOnly(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "page.png")
	writePage(t, in)

	var stdout, stderr bytes.Buffer
	j := newJob(&stdout, &stderr)
	j.cutsOnly = true
	j.verbose = true
	j.outDir = filepath.Join(dir, "out")
	if err := j.sliceFile(in); err != nil {
		t.Fatal(err)
	}

	if got := strings.TrimSpace(stdout.String()); !strings.HasPrefix(got, "[") || strings.Count(got, ",") != 1 {
		t.Errorf("stdout = %q, want a JSON array of 2 cuts", got)
	}
	if !strings.Contains(stderr.String(), "200x1500") {
		t.Errorf("-verbose logged nothing useful with -cuts: %q", stderr.String())
	}
	if _, err := os.Stat(j.outDir); !os.IsNotExist(err) {
		t.Errorf("-cuts created the output directory though it writes no files")
	}
}

func TestValidateHeights(t *testing.T) {
	for _, tc := range []struct {
		target, min int
		ok          bool
	}{
		{500, 400, true},
		{500, 1, true},
		{0, 400, false},
		{-5, 400, false},
		{500, 0, false},
		{500, 500, false},
		{500, 900, false},
	} {
		if err := validateHeights(tc.target, tc.min); (err == nil) != tc.ok {
			t.Errorf("validateHeights(%d, %d) = %v, want ok=%v", tc.target, tc.min, err, tc.ok)
		}
	}
}

func TestWriteFileRemovesPartialOutput(t *testing.T) {
	name := filepath.Join(t.TempDir(), "out.png")
	err := saveImage(image.NewGray(image.Rect(0, 0, 0, 0)), name, "png") // png rejects empty images
	if err == nil {
		t.Skip("encoder accepted an empty image")
	}
	if _, statErr := os.Stat(name); !os.IsNotExist(statErr) {
		t.Errorf("failed write left %s behind", name)
	}
}
