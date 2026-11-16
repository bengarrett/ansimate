// Package ansimate converts ANSI animations into animated WebP images.
//
//nolint:gochecknoglobals,dupl
package ansimate

import (
	"image/color"
)

const (
	c00 = 0x00
	c55 = 0x55
	caa = 0xaa
	cff = 0xff

	// IBM VGA BIOS NTSC grayscale luminance summing levels (30% R + 59% G + 11% B).
	m00 = 0x00
	m01 = 0x33
	m02 = 0x64
	m03 = 0x65
	m04 = 0x13
	m05 = 0x46
	m06 = 0x77
	m07 = 0xaa
	m08 = 0x55
	m09 = 0x88
	m10 = 0xb9
	m11 = 0xec
	m12 = 0x68
	m13 = 0x9b
	m14 = 0xcc
	m15 = 0xff

	// P134 Amber phosphor green channel luminance scaling (75% of red).
	a01 = 0x26
	a02 = 0x4b
	a03 = 0x4c
	a04 = 0x0e
	a05 = 0x35
	a06 = 0x59
	a07 = 0x80
	a08 = 0x40
	a09 = 0x66
	a10 = 0x8b
	a11 = 0xb1
	a12 = 0x4e
	a13 = 0x74
	a14 = 0x99
	a15 = 0xc0
)

var (
	// PaletteVGA contains the 16 standard IBM CGA/VGA colors.
	PaletteVGA = palettevga

	// PaletteMono contains the 16-level IBM 8503 VGA monochrome paper-white grayscale palette.
	PaletteMono = palettemono

	// PaletteAmber contains the 16-level P134 amber phosphor CRT monochrome palette.
	PaletteAmber = paletteamber

	// PaletteGreen contains the 16-level IBM 5151 P39 green phosphor CRT monochrome palette.
	PaletteGreen = palettegreen
)

var palettevga = [16]color.NRGBA{
	{R: c00, G: c00, B: c00, A: cff}, // 0: Black
	{R: caa, G: c00, B: c00, A: cff}, // 1: Red
	{R: c00, G: caa, B: c00, A: cff}, // 2: Green
	{R: caa, G: c55, B: c00, A: cff}, // 3: Brown
	{R: c00, G: c00, B: caa, A: cff}, // 4: Blue
	{R: caa, G: c00, B: caa, A: cff}, // 5: Magenta
	{R: c00, G: caa, B: caa, A: cff}, // 6: Cyan
	{R: caa, G: caa, B: caa, A: cff}, // 7: Light Gray
	{R: c55, G: c55, B: c55, A: cff}, // 8: Dark Gray
	{R: cff, G: c55, B: c55, A: cff}, // 9: Bright Red
	{R: c55, G: cff, B: c55, A: cff}, // 10: Bright Green
	{R: cff, G: cff, B: c55, A: cff}, // 11: Yellow
	{R: c55, G: c55, B: cff, A: cff}, // 12: Bright Blue
	{R: cff, G: c55, B: cff, A: cff}, // 13: Bright Magenta
	{R: c55, G: cff, B: cff, A: cff}, // 14: Bright Cyan
	{R: cff, G: cff, B: cff, A: cff}, // 15: White
}

var palettemono = [16]color.NRGBA{
	{R: m00, G: m00, B: m00, A: cff}, // 0: Black
	{R: m01, G: m01, B: m01, A: cff}, // 1: Red
	{R: m02, G: m02, B: m02, A: cff}, // 2: Green
	{R: m03, G: m03, B: m03, A: cff}, // 3: Brown
	{R: m04, G: m04, B: m04, A: cff}, // 4: Blue
	{R: m05, G: m05, B: m05, A: cff}, // 5: Magenta
	{R: m06, G: m06, B: m06, A: cff}, // 6: Cyan
	{R: m07, G: m07, B: m07, A: cff}, // 7: Light Gray
	{R: m08, G: m08, B: m08, A: cff}, // 8: Dark Gray
	{R: m09, G: m09, B: m09, A: cff}, // 9: Bright Red
	{R: m10, G: m10, B: m10, A: cff}, // 10: Bright Green
	{R: m11, G: m11, B: m11, A: cff}, // 11: Yellow
	{R: m12, G: m12, B: m12, A: cff}, // 12: Bright Blue
	{R: m13, G: m13, B: m13, A: cff}, // 13: Bright Magenta
	{R: m14, G: m14, B: m14, A: cff}, // 14: Bright Cyan
	{R: m15, G: m15, B: m15, A: cff}, // 15: White
}

var paletteamber = [16]color.NRGBA{
	{R: m00, G: c00, B: c00, A: cff}, // 0: Black
	{R: m01, G: a01, B: c00, A: cff}, // 1: Red
	{R: m02, G: a02, B: c00, A: cff}, // 2: Green
	{R: m03, G: a03, B: c00, A: cff}, // 3: Brown
	{R: m04, G: a04, B: c00, A: cff}, // 4: Blue
	{R: m05, G: a05, B: c00, A: cff}, // 5: Magenta
	{R: m06, G: a06, B: c00, A: cff}, // 6: Cyan
	{R: m07, G: a07, B: c00, A: cff}, // 7: Light Gray
	{R: m08, G: a08, B: c00, A: cff}, // 8: Dark Gray
	{R: m09, G: a09, B: c00, A: cff}, // 9: Bright Red
	{R: m10, G: a10, B: c00, A: cff}, // 10: Bright Green
	{R: m11, G: a11, B: c00, A: cff}, // 11: Yellow
	{R: m12, G: a12, B: c00, A: cff}, // 12: Bright Blue
	{R: m13, G: a13, B: c00, A: cff}, // 13: Bright Magenta
	{R: m14, G: a14, B: c00, A: cff}, // 14: Bright Cyan
	{R: m15, G: a15, B: c00, A: cff}, // 15: White
}

var palettegreen = [16]color.NRGBA{
	{R: c00, G: m00, B: c00, A: cff}, // 0: Black
	{R: c00, G: m01, B: c00, A: cff}, // 1: Red
	{R: c00, G: m02, B: c00, A: cff}, // 2: Green
	{R: c00, G: m03, B: c00, A: cff}, // 3: Brown
	{R: c00, G: m04, B: c00, A: cff}, // 4: Blue
	{R: c00, G: m05, B: c00, A: cff}, // 5: Magenta
	{R: c00, G: m06, B: c00, A: cff}, // 6: Cyan
	{R: c00, G: m07, B: c00, A: cff}, // 7: Light Gray
	{R: c00, G: m08, B: c00, A: cff}, // 8: Dark Gray
	{R: c00, G: m09, B: c00, A: cff}, // 9: Bright Red
	{R: c00, G: m10, B: c00, A: cff}, // 10: Bright Green
	{R: c00, G: m11, B: c00, A: cff}, // 11: Yellow
	{R: c00, G: m12, B: c00, A: cff}, // 12: Bright Blue
	{R: c00, G: m13, B: c00, A: cff}, // 13: Bright Magenta
	{R: c00, G: m14, B: c00, A: cff}, // 14: Bright Cyan
	{R: c00, G: m15, B: c00, A: cff}, // 15: White
}
