// Package ansimate converts ANSI animations into animated WebP images.
//
//nolint:gochecknoglobals
package ansimate

import (
	"image/color"
)

const (
	c00 = 0x00
	c55 = 0x55
	caa = 0xaa
	cff = 0xff
)

// Palette contains the 16 standard CGA/VGA colors ordered by ANSI SGR color indices (0-15).
var Palette = [16]color.NRGBA{
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

// BlinkMode defines how ANSI blink / SGR 5 is interpreted.
type BlinkMode int

const (
	// BlinkOff disables blink and iCE high-intensity backgrounds.
	BlinkOff BlinkMode = iota
	// BlinkText interprets SGR 5 as blinking text (hidden on alternate frames).
	BlinkText
	// BlinkICE interprets SGR 5 as high-intensity background (iCE color mode).
	BlinkICE
)

func (b BlinkMode) String() string {
	switch b {
	case BlinkOff:
		return "off"
	case BlinkText:
		return "blink"
	case BlinkICE:
		return "ice"
	default:
		return "unknown"
	}
}
