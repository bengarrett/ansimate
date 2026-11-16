ANSImate was mostly prompted and then manually cleaned up and refactored by myself in November 2026, using a combination of Gemini 4 Argon, Gemini 3.8 Flash and neovim. The models used a number of resources and references created by humans and these are listed here.

## ANSImate references, specifications and citations

This document details the historical specifications, technical standards, external software libraries, and reference tools consulted during the design and implementation of `ansimate`.

---

- **IBM PC BIOS ROM VGA 8×16 Font (`font_vga_8x16.bin`)**
  - Extracted directly from **SeaVGABIOS**"
    `/usr/share/seavgabios/vgabios-stdvga.bin`, from the open-source [SeaBIOS](https://www.seabios.org/SeaVGABIOS.html) project)
  - Originates from the public domain [`fntcol16.zip`](https://github.com/dokkaebi-project/font-archive) collection compiled by Joseph Gil on Simtel.Net.
- **ANSI X3.64-1979 / ECMA-48:**
  - **Standard:** _Control Functions for Coded Character Sets_ ([ECMA-48 5th Edition](https://www.ecma-international.org/publications-and-standards/standards/ecma-48/)).
  - Defines the Control Sequence Introducer (`CSI`, `\x1b[`).
- **MS-DOS `ANSI.SYS` Technical Specification:**
  - [Wikipedia `ANSI.SYS` Escape Sequences](https://en.wikipedia.org/wiki/ANSI.SYS#Escape_sequences)
  - [Microsoft MS-DOS Reference Archive](https://www.pcjs.org/documents/books/mspl13/msdos/)
  - [FreeDOS `NANSI.SYS`](https://gitlab.com/FreeDOS/drivers/nansi)
- **SAUCE (Standard Architecture for Universal Comment Extensions):**
  - **Specification:** SAUCE v00.5 (November, 2013) by Tasmaniac and ACiD Productions
  - [SAUCE specification](https://www.acid.org/info/sauce/sauce.htm)
  - **Cross-reference:** [bengarrett/sauce](https://github.com/bengarrett/sauce)
- **Google WebP Container Specification:**
  - [WebP Container Specification (RIFF Container)](https://developers.google.com/speed/webp/docs/riff_container)
- [**`AnsiLove/C`**](https://github.com/ansilove/ansilove) _(Stefan Vogt, Brian Cassidy, Frederic Cambus)_:
  - Referenced conceptually for standard 80×25 VGA character grid rendering and CP437 font bit blitting conventions.
- **[DOSBox](https://www.dosbox.com/) and [FreeDOS](https://freedos.org/)**:
  - Consulted for historical `ANSI.SYS` driver escape sequence specifications and ROM BIOS font definitions.

### Sibling Packages

- [**`github.com/bengarrett/sauce`**](https://github.com/bengarrett/sauce):
  - Referenced for exact SAUCE record byte offsets, data type definitions (`DataType 1`), and file type values (`FileType 2`).
- [**`github.com/bengarrett/ansibump`**](https://github.com/bengarrett/ansibump):
  - Referenced for ANSI parsing conventions.
