package ansimate

import (
	_ "embed"
	"image"
	"image/color"
)

//go:embed font_vga_8x16.bin
var fontVGA8x16 []byte

const (
	GlyphW  = 8             // GlyphW is the pixel width of each 8x16 font character.
	GlyphH  = 16            // GlyphH is the pixel height of each 8x16 font character.
	GlyphL  = 16            // GlyphL is the byte length of each 8x16 font character bitmap.
	Cols    = 80            // Cols is the standard terminal width in character cells.
	Rows    = 25            // Rows is the standard terminal height in character cells.
	CanvasW = Cols * GlyphW // CanvasW is the pixel width of a standard 80x25 text screen.
	CanvasH = Rows * GlyphH // CanvasH is the pixel height of a standard 80x25 text screen.

	bytesPerPixel = 4
	offsetR       = 0
	offsetG       = 1
	offsetB       = 2
	offsetA       = 3
	bitHigh       = 0x80
	maxColorIdx   = 15
)

// Cell represents a single character position on the screen.
type Cell struct {
	Char byte
	FG   byte
	BG   byte
}

// DrawCell renders a single character cell into the target image at the specified column and row.
func DrawCell(img *image.NRGBA, col, row int, cell Cell, pal [16]color.NRGBA) {
	fgIdx := cell.FG & maxColorIdx
	bgIdx := cell.BG & maxColorIdx
	fg := pal[fgIdx]
	bg := pal[bgIdx]

	startX := col * GlyphW
	startY := row * GlyphH
	glyphOffset := int(cell.Char) * GlyphL
	glyphData := fontVGA8x16[glyphOffset : glyphOffset+GlyphL]

	for y := range GlyphH {
		rowBits := glyphData[y]
		imgY := startY + y
		rowOffset := imgY * img.Stride

		for x := range GlyphW {
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

// DrawScreen renders an entire grid of cells into a new NRGBA image.
func DrawScreen(grid [][]Cell, pal [16]color.NRGBA) *image.NRGBA {
	cols := Cols
	rows := Rows
	if len(grid) > 0 {
		rows = len(grid)
		if len(grid[0]) > 0 {
			cols = len(grid[0])
		}
	}

	img := image.NewNRGBA(image.Rect(0, 0, cols*GlyphW, rows*GlyphH))
	DrawScreenBuf(img, grid, pal)

	return img
}

// DrawScreenBuf renders an entire grid of cells into an existing NRGBA image.
func DrawScreenBuf(img *image.NRGBA, grid [][]Cell, pal [16]color.NRGBA) {
	for row, line := range grid {
		for col, cell := range line {
			DrawCell(img, col, row, cell, pal)
		}
	}
}
