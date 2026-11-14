package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goodblaster/vertigo/internal/build"
	"github.com/goodblaster/vertigo/slicer"
)

func main() {
	flag.Usage = usage

	targetHeight := flag.Int("target-height", slicer.DefaultTargetHeight,
		"Target slice height in pixels. A soft limit: a cut may land above its\n"+
			"boundary and the shortfall carries into the next slice, so a slice can be\n"+
			"up to target-height plus (target-height - min-height) pixels tall.")
	minHeight := flag.Int("min-height", slicer.DefaultMinHeight,
		"Minimum slice height in pixels. Also a soft limit: the final slice may be\n"+
			"shorter. The rows between min-height and target-height form the window\n"+
			"searched for a clean cut.")
	format := flag.String("format", "jpg",
		"Output image format: jpg, png, or tiff.")
	outDir := flag.String("out", "",
		"Directory to write slices to, created if missing. Defaults to the\n"+
			"directory of each input file.")
	cutsOnly := flag.Bool("cuts", false,
		"Print the chosen cut rows as a JSON array instead of writing image slices.\n"+
			"An image that already fits in one slice prints [].")
	debug := flag.Bool("debug", false,
		"Also write <input>-cuts.jpg: a copy of the input image with a pink line\n"+
			"drawn at each chosen cut row. Its filename is printed to stderr.")
	verbose := flag.Bool("verbose", false,
		"Log image dimensions and timing to stderr.")
	version := flag.Bool("version", false,
		"Print version and build information, then exit.")
	flag.Parse()

	if *version {
		printVersion()
		return
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	outFormat, err := normalizeFormat(*format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vertigo: %v\n", err)
		os.Exit(2)
	}

	if err := validateHeights(*targetHeight, *minHeight); err != nil {
		fmt.Fprintf(os.Stderr, "vertigo: %v\n", err)
		os.Exit(2)
	}

	job := &job{
		slc:      slicer.New(*targetHeight, *minHeight),
		outDir:   *outDir,
		format:   outFormat,
		cutsOnly: *cutsOnly,
		debug:    *debug,
		verbose:  *verbose,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}

	failed := false
	for _, path := range flag.Args() {
		if err := job.sliceFile(path); err != nil {
			fmt.Fprintf(os.Stderr, "vertigo: %s: %v\n", path, err)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}

// validateHeights rejects height flags the slicer would otherwise silently
// replace with defaults.
func validateHeights(targetHeight, minHeight int) error {
	if targetHeight <= 0 {
		return fmt.Errorf("-target-height must be positive, got %d", targetHeight)
	}
	if minHeight <= 0 {
		return fmt.Errorf("-min-height must be positive, got %d", minHeight)
	}
	if minHeight >= targetHeight {
		return fmt.Errorf("-min-height (%d) must be less than -target-height (%d)", minHeight, targetHeight)
	}
	return nil
}

// job holds the settings shared by every input of one run.
type job struct {
	slc      slicer.Slicer
	outDir   string // empty: write next to each input
	format   string
	cutsOnly bool
	debug    bool
	verbose  bool
	stdout   io.Writer
	stderr   io.Writer

	// claimed maps each output name prefix (directory plus base name) to the
	// input that writes there, so two inputs never overwrite each other.
	claimed map[string]string
}

// sliceFile slices one input image, printing each written filename (or, with
// cutsOnly, the cut rows as JSON) to stdout. The path "-" reads the image
// from standard input and names its slices stdin-000.<format>, ...
func (j *job) sliceFile(path string) error {
	dir := filepath.Dir(path)
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if path == "-" {
		dir = "."
		base = "stdin"
	}
	if j.outDir != "" {
		dir = j.outDir
	}

	writes := !j.cutsOnly || j.debug
	if writes {
		prefix := filepath.Join(dir, base)
		if other, taken := j.claimed[prefix]; taken {
			return fmt.Errorf("output names %s-* are already used by %s; skipped", prefix, other)
		}
		if j.claimed == nil {
			j.claimed = map[string]string{}
		}
		j.claimed[prefix] = path
	}

	img, err := loadImage(path)
	if err != nil {
		return err
	}

	if writes && j.outDir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	start := time.Now()
	cuts := j.slc.Cuts(img)
	if j.verbose {
		b := img.Bounds()
		fmt.Fprintf(j.stderr, "%s: %dx%d, %d slices, cuts found in %v\n", path, b.Dx(), b.Dy(), len(cuts)+1, time.Since(start))
	}

	if j.debug {
		name := filepath.Join(dir, base+"-cuts.jpg")
		if err := writeDebugImage(img, cuts, name); err != nil {
			return err
		}
		// To stderr: stdout carries only slice filenames, safe to pipe.
		fmt.Fprintln(j.stderr, name)
	}

	if j.cutsOnly {
		if cuts == nil {
			cuts = []int{}
		}
		out, err := json.Marshal(cuts)
		if err != nil {
			return err
		}
		fmt.Fprintln(j.stdout, string(out))
		return nil
	}

	for i, im := range slicer.SliceAt(img, cuts) {
		name := filepath.Join(dir, fmt.Sprintf("%s-%03d.%s", base, i, j.format))
		if err := saveImage(im, name, j.format); err != nil {
			return err
		}
		fmt.Fprintln(j.stdout, name)
	}
	return nil
}

func normalizeFormat(format string) (string, error) {
	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		return "jpg", nil
	case "png":
		return "png", nil
	case "tiff", "tif":
		return "tiff", nil
	}
	return "", fmt.Errorf("unsupported format %q (supported: jpg, png, tiff)", format)
}

func printVersion() {
	info := build.Info()
	fmt.Printf("vertigo %s", info.Version)
	if info.Commit != "" {
		fmt.Printf(" (%s)", info.Commit)
	}
	if info.Date != "" {
		fmt.Printf(" built %s", info.Date)
	}
	fmt.Println()
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `vertigo %s

Cut tall images into shorter horizontal slices sized for OCR. Each cut lands
at the largest visually blank gap in its search window, so lines of text are
never severed and cuts tend to fall at paragraph breaks.

Usage:
  vertigo [flags] <image> [<image> ...]

Each input produces one or more slices named <input>-000.<format>,
<input>-001.<format>, ... written next to the input file (or to -out).
The written filenames are printed to stdout, one per line, ready to pipe
into the next command. An input of - reads one image from standard input
and writes slices named stdin-000.<format> to the current directory
(or to -out).

Flags:
`, build.Info().Version)
	flag.PrintDefaults()
}
