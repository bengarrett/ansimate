# About Ansimate: Specifications, Standards & Citations

`ansimate` is a pure Go package and CLI tool designed to detect ANSI animations (ANSImations) and render them into lossless, authentic IBM VGA animated WebP files.

This document details the historical specifications, technical standards, external software libraries, and reference tools consulted during the design and implementation of `ansimate`.

---

## 1. Font & Character Encoding

- **IBM PC BIOS ROM VGA 8×16 Font (`font_vga_8x16.bin`):**
  - **Source Extraction:** Extracted directly from **SeaVGABIOS** (`/usr/share/seavgabios/vgabios-stdvga.bin`, from the open-source [SeaBIOS](https://www.seabios.org/SeaBIOS) project).
  - **Binary Offset:** Offset `0x7360` through `0x8360` (exactly 4,096 bytes: 256 characters × 16 scanlines per character).
  - **SHA-256:** `c7dd5990fa5401649cf69e9bbb4c8a51f86865939b9318281a27de7f314eda3a`
  - **Historical Provenance:** The 8×16 VGA character font arrays in SeaVGABIOS (`vgasrc/vgafonts.c`) and Bochs VGABios originate from the classic `fntcol16.zip` collection compiled by Joseph Gil on Simtel.Net.
  - **License Status:**
    - **Raw Bitmap Data:** Designated as **Public Domain** by Joseph Gil in `fntcol16.zip`. Furthermore, under US Copyright Law (37 C.F.R. § 202.1(e) and *Eltra Corp. v. Ringer*, 579 F.2d 294), raw bitmap font data and typeface designs are not eligible for copyright protection (unlike scalable vector font software code such as TrueType bytecode).
    - **SeaVGABIOS Project:** The surrounding SeaVGABIOS wrapper and build infrastructure are licensed under **GNU LGPLv3**.
  - **Specification:** IBM Code Page 437 (DOS Latin US) character set mapping ([IBM CP437 Reference](https://en.wikipedia.org/wiki/Code_page_437)).
  - **Implementation:** Embedded directly into the package binary via `//go:embed font_vga_8x16.bin` and rendered via a zero-allocation bitwise blitter (`font.go`).

---

## 2. Palette & Color Standards

- **CGA / EGA / VGA 16-Color RGB DAC Palette:**
  - **Standard:** Standard IBM 16-color PC palette (DAC color registers 0–15).
  - **Color 6 (Brown):** Utilizes the authentic IBM 5153 color display monitor DAC circuit adjustment (`#AA5500` / `RGB(170, 85, 0)`) rather than raw dark yellow (`#AAAA00` / `RGB(170, 170, 0)`).
- **IBM VGA BIOS Monochrome & Phosphor Palettes:**
  - **Gray-Scale Summing (`INT 10h, AH=10h, AL=1Bh`):** When driving a monochrome VGA monitor (e.g., IBM 8503), the VGA BIOS sums the 16 DAC registers using NTSC luminance weights ($30\%\text{ R} + 59\%\text{ G} + 11\%\text{ B}$), mapping each SGR color index to a distinct 64-level grayscale intensity rather than 1-bit black-and-white.
  - **`PaletteMono`:** IBM 8503 paper-white monochrome grayscale (`RGB(Y, Y, Y)`).
  - **`PaletteAmber`:** Classic P134 Amber phosphor CRT monochrome (`RGB(Y, 0.75Y, 0)`).
  - **`PaletteGreen`:** Classic IBM 5151 P39 Green phosphor CRT monochrome (`RGB(0, Y, 0)`).

---

## 3. Terminal Emulation & Escape Sequences

- **ANSI X3.64-1979 / ECMA-48:**
  - **Standard:** *Control Functions for Coded Character Sets* ([ECMA-48 5th Edition](https://www.ecma-international.org/publications-and-standards/standards/ecma-48/)).
  - Defines the Control Sequence Introducer (`CSI`, `\x1b[`).
- **MS-DOS `ANSI.SYS` Technical Specification:**
  - Standard reference for PC terminal emulation behavior:
    - **Cursor positioning:** `CSI <row>;<col>H` and `CSI <row>;<col>f`
    - **Directional movement:** `CSI <n>A` (Up), `CSI <n>B` (Down), `CSI <n>C` (Right), `CSI <n>D` (Left)
    - **Erasure:** `CSI 2J` (Clear entire screen) and `CSI 2K` (Clear line)
    - **Cursor state:** `CSI s` (Save cursor position) and `CSI u` (Restore cursor position)
    - **Line wrapping:** `CSI ?7h` / `CSI =7h` (Wrap enable), `CSI ?7l` / `CSI =7l` (Wrap disable)
    - **SGR attributes:** `CSI <n>m` (0 = reset, 1 = bold/bright, 30–37 = foreground, 40–47 = background).

---

## 4. SAUCE Metadata Specification

- **SAUCE (Standard Architecture for Universal Comment Extensions):**
  - **Specification:** SAUCE v00.5 (March 1994) by Tasmaniac / ACiD Productions (Lori Gardner & Richard Elson).
  - **Fields used in detection:**
    - 128-byte record parsing at EOF (`SAUCE` magic tag).
    - `DataType == 1` (`Character`).
    - `FileType == 2` (`AnsiMation` — explicitly flagged ANSI animation).
  - **Cross-reference:** [bengarrett/sauce](https://github.com/bengarrett/sauce).

---

## 5. WebP Container & Animation Encoding

- **Google WebP Container Specification:**
  - [WebP Container Specification (RIFF Container)](https://developers.google.com/speed/webp/docs/riff_container):
    - `VP8X` extended header chunk (dimensions, canvas flags, animation flag).
    - `ANIM` global animation parameters (background color, loop count).
    - `ANMF` frame chunks (sub-frame coordinates `Offset_X`/`Offset_Y`, frame dimensions, duration, dispose method, alpha blend method).

---

## 6. Modem Timing & BBS Stream Simulation

- **Serial Framing (UART / RS-232):**
  - Standard asynchronous BBS serial communication uses 10 bits per transmitted character byte (1 start bit, 8 data bits, 1 stop bit, no parity).
  - Pacing formulas:
    $$\text{Bytes Per Second} = \frac{\text{Baud Rate}}{10}$$
    $$\text{Bytes Per Frame Tick} = \frac{\text{Bytes Per Second}}{\text{FPS}}$$

---

## 7. Software Libraries & Reference Implementations

### Direct Go Dependencies
- [**`github.com/deepteams/webp`**](https://github.com/deepteams/webp):
  - Used for pure Go WebP encoding and RIFF animation container muxing (`github.com/deepteams/webp/animation`).
  - Provides the `AnimEncoder` for frame encoding (`VP8L` lossless / `VP8` lossy), sub-frame dirty-rect diffing, loop headers, and frame disposal/blending.
- [**`github.com/nalgeon/be`**](https://github.com/nalgeon/be):
  - Used throughout unit tests and golden regression suites for clear, zero-dependency assertions (`be.True`, `be.Equal`, `be.Err`).

### Sibling Ecosystem Packages Consulted
- [**`github.com/bengarrett/sauce`**](https://github.com/bengarrett/sauce):
  - Referenced for exact SAUCE record byte offsets, data type definitions (`DataType 1`), and file type values (`FileType 2`).
- [**`github.com/bengarrett/ansibump`**](https://github.com/bengarrett/ansibump):
  - Referenced for ANSI parsing conventions.

### External Tools & Architecture References
- [**`AnsiLove/C`**](https://github.com/ansilove/ansilove) *(Stefan Vogt, Brian Cassidy, Frederic Cambus)*:
  - Referenced conceptually for standard 80×25 VGA character grid rendering and CP437 font bit blitting conventions.
- **Google `libwebp` (`webpinfo`, `webpmux`, `cwebp`)**:
  - The `webpinfo` utility was used during development to verify generated WebP RIFF chunk trees (`VP8X`, `ANIM`, `ANMF`), frame counts, sub-frame bounding boxes, and timestamp durations.
- **DOSBox & FreeDOS**:
  - Consulted for historical `ANSI.SYS` driver escape sequence specifications and ROM BIOS font definitions.

---

## 8. Golden Test Fixtures

Authentic 1990s scene ANSImations (see [testdata/README.md](testdata/README.md) for full attributions and copyright details):
- `CHAOS2.ANS` by Chaos / ACiD Productions
- `MAXIMUM3.ANS` by Maximum / ACiD Productions
- `SC-N0393.ANS` by Soulcatcher / ACiD Productions
- `SSACID4.ANS` by Seven Sisters / ACiD Productions
