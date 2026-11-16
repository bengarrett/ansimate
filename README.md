# ANSImate

[![Go Reference](https://pkg.go.dev/badge/github.com/bengarrett/ansimate.svg)](https://pkg.go.dev/github.com/bengarrett/ansimate)

ANSImate converts continuous ANSI streams (ANSimations) found in the PC BBS and MS-DOS era into lightweight, lossless animated WebP images in pure Go.

#### Features

- Pure Go with no external binary dependencies.
- DOS VGA rendering using an authentic IBM VGA, 8×16 BIOS ROM font bitmap.
- Emulates BBS streaming transmission rates and frame ticks.
- Lossless animated WebP output.
- Heuristic and SAUCE detection to distinguish dynamic animations from static ANSI artworks.
- ANSI.SYS terminal support including cursor positioning, movements (`CSI A-D`), screen/line clearing, and SGR colors.

#### Quick Usage

Convert an ANSImation ANSI file to an web-friendly, animated WebP image.

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

	if err := ansimate.Animate(output, input, nil); err != nil {
		log.Fatal(err)
	}
}
```

#### Detect if a file is an ANSImation

```go
ok, err := ansimate.Detect(file)
if err != nil {
	log.Fatal(err)
}
if ok {
	fmt.Println("File is an ANSImation!")
}
```

#### Custom Configuration

```go
cfg := &ansimate.Config{
	BaudRate:    14400,                  // 14.4k baud modem speed
	FPS:         20,                     // 20 frames per second
	MaxDuration: 10 * time.Second,       // Cap animation at 10s
	EndDelay:    3 * time.Second,        // Hold final completed frame for 3s
	Lossy:       true,                   // Use lossy image compression, but it is not recommended
	Palette:     ansimate.PaletteMono,   // Or PaletteAmber, PaletteGreen, PaletteVGA
}

err := ansimate.Animate(output, input, cfg)
```

#### Command-Line Tool

A standalone CLI is provided for converting files without writing Go code:

```bash
# Build the binary
cd ansimate
go build .

# Convert an ANSImation to WebP
./ansimate animation.ans

# Check whether a file is an ANSImation
./ansimate -check animation.ans

# Convert with custom baud rate, end pause delay, and monochrome palette
./ansimate -baud 28800 -fps 20 -delay 3.0s -palette mono animation.ans output.webp
```

### License

This project is licensed under the [MIT License](LICENSE).

Test artwork files located in `testdata/` remain the copyright of their respective authors and are included strictly for automated regression testing and format conformance. See [testdata/README.md](testdata/README.md) for full attributions.
