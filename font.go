package ansimate

import (
	_ "embed"
	"image"
	"image/color"
)

//go:embed font_vga_8x16.bin
var fontVGA8x16 []byte

const (
	glyphW  = 8             // glyphW is the pixel width of each 8x16 font character.
	glyphH  = 16            // glyphH is the pixel height of each 8x16 font character.
	glyphL  = 16            // glyphL is the byte length of each 8x16 font character bitmap.
	canvasW = Cols * glyphW // canvasW is the pixel width of a standard 80x25 text screen.
	canvasH = Rows * glyphH // canvasH is the pixel height of a standard 80x25 text screen.

	bytesPerPixel = 4
	offsetR       = 0
	offsetG       = 1
	offsetB       = 2
	offsetA       = 3
	bitHigh       = 0x80
	maxColorIdx   = 15
)

// cell represents a single character position on the screen.
type cell struct {
	Char  byte
	FG    byte
	BG    byte
	Blink bool
}

// drawCell renders a single character cell into the target image at the specified column and row.
func drawCell(img *image.NRGBA, col, row int, c cell, pal [16]color.NRGBA) {
	drawCellBlink(img, col, row, c, pal, true)
}

// drawCellBlink renders a single character cell, optionally blanking blinking characters.
func drawCellBlink(img *image.NRGBA, col, row int, c cell, pal [16]color.NRGBA, blinkVisible bool) {
	fgIdx := c.FG & maxColorIdx
	bgIdx := c.BG & maxColorIdx
	fg := pal[fgIdx]
	bg := pal[bgIdx]

	if c.Blink && !blinkVisible {
		fg = bg
	}

	startX := col * glyphW
	startY := row * glyphH
	glyphOffset := int(c.Char) * glyphL
	glyphData := fontVGA8x16[glyphOffset : glyphOffset+glyphL]

	for y := range glyphH {
		rowBits := glyphData[y]
		imgY := startY + y
		rowOffset := imgY * img.Stride

		for x := range glyphW {
			imgX := startX + x
			pixOffset := rowOffset + imgX*bytesPerPixel

			c := bg
			if (rowBits & (bitHigh >> x)) != 0 {
				c = fg
			}

			img.Pix[pixOffset+offsetR] = c.R
			img.Pix[pixOffset+offsetG] = c.G
			img.Pix[pixOffset+offsetB] = c.B
			img.Pix[pixOffset+offsetA] = c.A
		}
	}
}

// drawScreen renders an entire grid of cells into a new NRGBA image.
func drawScreen(grid [][]cell, pal [16]color.NRGBA) *image.NRGBA {
	return drawScreenBlink(grid, pal, true)
}

// drawScreenBlink renders an entire grid of cells into a new NRGBA image with blink visibility.
func drawScreenBlink(grid [][]cell, pal [16]color.NRGBA, blinkVisible bool) *image.NRGBA {
	cols := Cols
	rows := Rows
	if len(grid) > 0 {
		rows = len(grid)
		if len(grid[0]) > 0 {
			cols = len(grid[0])
		}
	}

	img := image.NewNRGBA(image.Rect(0, 0, cols*glyphW, rows*glyphH))
	for row, line := range grid {
		for col, c := range line {
			drawCellBlink(img, col, row, c, pal, blinkVisible)
		}
	}

	return img
}
