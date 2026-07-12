# Vertigo

Cut tall images into shorter horizontal slices sized for OCR.

Each cut lands at the largest visually blank gap in its search window, so
lines of text are never severed — and because the biggest gaps tend to be
paragraph breaks, whole paragraphs usually stay together.

Available as both a Go package and a standalone CLI.

---

## CLI

### Install / build

```bash
$ go install github.com/goodblaster/vertigo@latest
# or, from a checkout:
$ go build -o vertigo .   # or: make build
```

### Usage

```bash
$ vertigo [flags] <image> [<image> ...]
```

Each input produces one or more slices named `<input>-000.<format>`,
`<input>-001.<format>`, ... written next to the input file (or to `-out`).
The written filenames are printed to stdout, one per line.

```bash
$ vertigo page.png
page-000.jpg
page-001.jpg
page-002.jpg

# Write PNG slices into ./out
$ vertigo -format png -out out page.png

# Just print the cut rows as JSON
$ vertigo -cuts page.png
[483,974]
```

### Piping

Stdout carries only the written slice filenames (the `-debug` overlay's name
goes to stderr), so the output feeds straight into the next command. An input
of `-` reads one image from standard input and names its slices
`stdin-000.<format>`, ...

```bash
# OCR each slice as it is written
$ vertigo page.png | while read -r f; do tesseract "$f" "${f%.*}"; done

# Slice an image straight from a URL
$ curl -s https://example.com/page.png | vertigo -
stdin-000.jpg
stdin-001.jpg

# Cut rows for a piped image, as JSON
$ curl -s https://example.com/page.png | vertigo -cuts -
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-target-height` | `500` | Target slice height in pixels. A soft limit: a cut may land above its boundary and the shortfall carries into the next slice, so a slice can be up to `target-height + (target-height - min-height)` tall. |
| `-min-height` | `400` | Minimum slice height in pixels. Also a soft limit: the final slice may be shorter. The rows between `min-height` and `target-height` form the window searched for a clean cut. |
| `-format` | `jpg` | Output image format: `jpg`, `png`, or `tiff`. |
| `-out` | | Directory to write slices to. Defaults to the directory of each input file. |
| `-cuts` | `false` | Print the chosen cut rows as a JSON array instead of writing image slices. |
| `-debug` | `false` | Also write `<input>-cuts.jpg`: the input image with a pink line drawn at each cut row. Its filename is printed to stderr. |
| `-verbose` | `false` | Log image dimensions and timing to stderr. |
| `-version` | `false` | Print version and build information, then exit. |

---

## Go package

```bash
$ go get github.com/goodblaster/vertigo
```

```go
package main

import (
	"image"

	"github.com/goodblaster/vertigo/slicer"
)

func slicePage(img image.Image) []image.Image {
	slc := slicer.New(500, 400) // target height, min height (pixels)
	return slc.Slice(img)
}
```

- `slicer.Default()` — a slicer with the default heights (500/400).
- `slc.Slice(img)` — cut an image into slices; an image already within
  `TargetHeight` comes back unchanged as a single slice.
- `slc.Cuts(img)` — just the cut rows, if you want to inspect or store them.
- `slicer.SliceAt(img, cuts)` — cut at rows you choose yourself.

The library depends only on the Go standard library.

---

## How it works

1. The image is converted to grayscale and analyzed on luminance alone.
2. Each row is compared with the row below it. A row whose every pixel
   matches the row below within a small contrast tolerance is *quiet*, and
   runs of quiet rows form *gaps* — the blank space between lines of text.
   Only rows inside a search window are compared, which skips about 80% of
   the work at the default heights.
3. The Nth search window is the last `target-height - min-height` rows
   before row `N * target-height`. The cut lands at the middle of the
   biggest gap in the window; a window with no gap is cut at its bottom
   boundary.
4. The image is cut at the chosen rows.
