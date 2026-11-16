# ANSI-mate

[![Go Reference](https://pkg.go.dev/badge/github.com/bengarrett/ansimate.svg)](https://pkg.go.dev/github.com/bengarrett/ansimate)

`ansimate` converts continuous ANSI animations (ANSimations) from the BBS and MS-DOS era into lightweight, lossless animated WebP images in pure Go.

## Features

- **Pure Go & Portable:** Zero CGo and zero external binary dependencies (no need for ImageMagick, Ansilove, or ffmpeg).
- **Pixel-Perfect DOS VGA Rendering:** Uses an embedded authentic 4,096-byte CP437 IBM VGA 8×16 BIOS ROM font bitmap. Renders sharp, authentic 640×400 canvases without anti-aliasing fuzziness.
- **Modem Baud Rate Pacing:** Emulates realistic BBS streaming transmission rates (e.g., 9,600, 14,400, 19,200, or 28,800 baud) and frame ticks.
- **Efficient Lossless Animated WebP:** Leverages pure-Go WebP encoding with sub-frame diffing (encoding only changed bounding boxes) and automatic identical frame merging.
- **ANSImation Detection:** Fast heuristic and SAUCE detection to distinguish dynamic animations from static ANSI artwork without rasterization.
- **Full ANSI & iCE Support:** Handles standard cursor positioning (`CSI H`), movements (`CSI A-D`), screen/line clearing (`CSI 2J`, `CSI 2K`), SGR colors, and iCE high-intensity backgrounds (`BlinkICE`).

## Quick Usage

### Convert an ANSImation to Animated WebP

```go
package main

import (
	"log"
	"os"

	"github.com/bengarrett/ansimate"
)

func main() {
	input, err := os.Open("animation.ans")
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()

	output, err := os.Create("animation.webp")
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	// Default config: 19,200 baud, 20 FPS, 15s max duration, 2s end delay
	if err := ansimate.Animate(output, input, nil); err != nil {
		log.Fatal(err)
	}
}
```

### Detect if an ANSI File is an ANSImation

```go
isAnimation, err := ansimate.Detect(file)
if err != nil {
	log.Fatal(err)
}
if isAnimation {
	fmt.Println("File is an ANSImation!")
}
```

### Custom Configuration

```go
cfg := &ansimate.Config{
	BaudRate:    14400,           // 14.4k baud modem speed
	FPS:         20,              // 20 frames per second
	MaxDuration: 10 * time.Second,// Cap animation at 10s
	EndDelay:    3 * time.Second, // Hold final completed frame for 3s
	Blink:       ansimate.BlinkICE, // Interpret SGR 5 as iCE high-intensity background
	Lossy:       false,           // Default false (pixel-perfect lossless WebP)
}

err := ansimate.Animate(output, input, cfg)
```

## Command-Line Tool

A standalone CLI is provided for converting files without writing Go code:

```bash
# Build the binary
task build

# Convert an ANSImation to WebP
./ansimate animation.ans

# Check whether a file is an ANSImation
./ansimate -check animation.ans

# Convert with custom baud rate and end pause delay
./ansimate -baud 28800 -fps 20 -delay 3.0s animation.ans output.webp
```

## License

This project is licensed under the [MIT License](LICENSE).

Test artwork files located in `testdata/` remain the copyright of their respective authors and are included strictly for automated regression testing and format conformance. See [testdata/README.md](testdata/README.md) for full attributions.
