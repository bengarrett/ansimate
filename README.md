# ANSImate

[![Go Reference](https://pkg.go.dev/badge/github.com/bengarrett/ansimate.svg)](https://pkg.go.dev/github.com/bengarrett/ansimate)

ANSImate converts continuous ANSI streams (ANSImations) found in the PC BBS and MS-DOS era into lightweight, lossless animated WebP images. [Check out the live web demo](https://bengarrett.github.io/ansimate/).

#### Features

- Pure Go with no external binary dependencies.
- DOS VGA/CGA rendering using an authentic IBM PC BIOS ROM fonts.
- Emulates BBS streaming transmission rates and frame ticks.
- Lossless animated WebP output.
- Heuristic and SAUCE detection to distinguish dynamic animations from regular ANSI artworks.
- MS-DOS, ANSI.SYS terminal support including cursor positioning, movements, screen/line clearing, and SGR colors.
- WebAssembly support for standalone runtimes and browser client-side conversion.

#### ANSImation?

ANSI escape codes are a terminal standard that goes back to minicomputers of the late 1970s. On 16-bit microcomputers, it became popular on online bulletin boards later in the 1980s. This allowed boards to offer color, cursor movement, and overall more striking user interfaces, though it came at the expense of complexity. The creators of these interfaces evolved into artists drawing adverts for these bulletin boards, and popular drawing tools of the era like [TheDraw](https://winworldpc.com/product/the-draw/3x) allowed for the creation of streaming ANSI art files, to display the artwork onscreen as they were downloaded over slow, copper telephone lines.

We first started seeing ANSI artists using the blink character options to add visual movement to their otherwise static art. And I would guess this evolved into exploring the ability to control the text cursor movement to generate animations. These animations initially started as TV and movie-style screen wipes and later expanded in complexity into moving adverts and animated comic strips, and these became known as ANSImations (sometimes styled with a lowercase i).

[![Blinking text example](readme/BBS.webp)](https://defacto2.net/f/ba2bcbb)

_1985 advert for The Game Gallery BBS_

[![Early ANSImation example](readme/WELCOME.webp)](https://defacto2.net/f/b21a2ed)

_1988 advert for the Pass It On BBS_

[![ANSImation](readme/HOLIC2.webp)](https://defacto2.net/f/ac2c5c0)

_1991 advert for the BBS-a-Holic BBS by RaD Man_

[![Then they went animated comic book](readme/KS-RGRK1.webp)](https://defacto2.net/f/b02d6c9)

_1993 'A Tribute Ansimation' By Kronos_

<small>
  The example images were created using ANSImate, however, the &copy; belongs to their authors.
  These authors do not endorse ANSImate and usage permission in this article was never asked.
</small>

#### Terminal Utility

A small, standalone program is provided for converting files without the need for writing Go code.

#### [Download the latest release](https://github.com/bengarrett/ansimate/releases/latest)

```bash
# Or, build the binary
cd ansimate
go build .

# Convert an ANSImation to WebP
./ansimate animation.ans

# Check whether a file is an ANSImation
./ansimate -check animation.ans

# Convert with custom baud rate, end pause delay, and monochrome palette
./ansimate -baud 28800 -fps 20 -delay 3.0s -palette mono animation.ans output.webp
```

#### WebAssembly

#### [Download binaries](https://github.com/bengarrett/ansimate/releases/latest)

`ansimate.wasm` can be used with [Wasmtime](https://wasmtime.dev/), and other runtimes:

```bash
wasmtime run --dir=. ansimate.wasm -- input.ans output.webp
```

`ansimate-web.wasm` can be combined with JS for use in web browsers. See [ansimate-web/README](cmd/ansimate-web/README.md).

And, [there is also the live demo](https://bengarrett.github.io/ansimate/).

#### Go Package, Quick Usage

Convert an ANSImation ANSI file to a web friendly, animated WebP image.

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
  MaxDuration: 10 * time.Second,       // Crop/Cap animation to 10 seconds
  EndDelay:    3 * time.Second,        // Hold the last frame for 3 seconds to create a pause
  Columns:     80,                     // Standard MS-DOS number of characters per line
  Rows:        25,                     // Standard MS-DOS number of lines on the screen
  Lossy:       true,                   // Use lossy image compression, not recommended
  CGA:         false,                  // Do not use the CGA 8x8 font
  Palette:     ansimate.PaletteMono,   // Use monochrome output,
                                       // or swap to PaletteAmber, PaletteGreen, PaletteVGA
}

err := ansimate.Animate(output, input, cfg)
```

#### Animated PNG (APNG) or animated GIF Support

To keep this package small and targeted, ANSImate only supports WebP.
While you can convert these animations into legacy formats using [ImageMagick](https://imagemagick.org/),
the tool can suffer from dependency dramas and the animations do balloon in size.

```sh
cd testdata/

# webp to gif
magick CHAOS2.webp CHAOS2.gif

# webp to apng
magick CHAOS2.webp -coalesce APNG:CHAOS2.png

# list the byte file sizes of CHAOS2
stat -c '%-12n %s' CHAOS2.*
CHAOS2.ANS   19426
CHAOS2.webp  55582
CHAOS2.gif   249387
CHAOS2.png   1176224
```

#### Limitations

To keep this package small and concise, there is only one character set,
the original IBM CP-437. And only two font choices.

#### Similar Projects

- [AnsiLove/C](https://www.ansilove.org/) is the classic, long standing tool for converting ANSI to PNG images.
- [AnsiLove.js](https://ansilove.github.io/ansilove.js/) is a JS port by the same team and with ANSImation support! But requires programming to use. You can try it out at [ansitube](https://ansilove.github.io/ansilove.js/ansitube/). Both tools also have far more options for customizations with fonts, CodePages, and languages.

#### Art

- There is a large collection of easy to navigate art on [Defacto2](https://defacto2.net/files/ansi), ordered by year.
- [16colo.rs](https://16colo.rs/tags/content/ansimation) has hundreds of art tagged as ansimation.

Just remember, many of these works were by edgy or nerd kids and young adults of the 1980-90s, online using computers, long before it was mainstream. So plenty of drawings lean heavily into cartoons, comic book characters, and teen boy humour.

### License

This project is licensed under the [MIT License](LICENSE).

Test artwork files located in `testdata/` remain the copyright of their respective authors and are included strictly for automated regression testing and format conformance. See [testdata/README.md](testdata/README.md) for full attributions.

```

```
