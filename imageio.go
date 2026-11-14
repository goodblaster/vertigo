package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"github.com/hhrutter/tiff"
)

// loadImage reads and decodes an image file (jpeg, png, or tiff). The
// name "-" reads from standard input instead.
func loadImage(filename string) (image.Image, error) {
	f := os.Stdin
	if filename != "-" {
		var err error
		f, err = os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()
	}

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}

// encodeImage writes img to w in the given format: "png", "tiff", or "jpg".
// Unrecognized formats are written as jpg.
func encodeImage(w io.Writer, img image.Image, format string) error {
	switch format {
	case "png":
		return png.Encode(w, img)
	case "tiff":
		return tiff.Encode(w, img, &tiff.Options{Compression: tiff.LZW})
	default:
		return jpeg.Encode(w, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
	}
}

// saveImage writes img to filename in the given format.
func saveImage(img image.Image, filename, format string) error {
	return writeFile(filename, func(w io.Writer) error {
		return encodeImage(w, img, format)
	})
}

// writeFile creates filename and fills it with encode. A file that could not
// be written in full, including one whose close fails, is removed rather
// than left behind truncated.
func writeFile(filename string, encode func(io.Writer) error) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", filename, err)
	}

	err = encode(f)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(filename)
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}
	return nil
}

// writeDebugImage saves a JPEG copy of img with a pink line drawn at each
// cut row.
func writeDebugImage(img image.Image, rows []int, filename string) error {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	pink := color.RGBA{R: 255, G: 105, B: 180, A: 255}
	for _, row := range rows {
		if row < bounds.Min.Y || row >= bounds.Max.Y {
			continue
		}
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, row, pink)
		}
	}

	return writeFile(filename, func(w io.Writer) error {
		return jpeg.Encode(w, rgba, nil)
	})
}
