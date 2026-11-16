package ansimate

import (
	_ "embed"
	"image"
	"image/color"
)

//go:embed font_vga_8x16.bin
var fontVGA8x16 []byte

const (
	// GlyphWidth is the pixel width of each 8x16 font character.
	GlyphWidth = 8
	// GlyphHeight is the pixel height of each 8x16 font character.
	GlyphHeight = 16
	// GlyphBytes is the byte length of each 8x16 font character bitmap.
	GlyphBytes = 16
	// FontGlyphs is the total number of characters in the CP437 font ROM.
	FontGlyphs = 256
	// DefaultColumns is the standard terminal width in character cells.
	DefaultColumns = 80
	// DefaultRows is the standard terminal height in character cells.
	DefaultRows = 25
	// CanvasWidth is the pixel width of a standard 80x25 text screen.
	CanvasWidth = DefaultColumns * GlyphWidth
	// CanvasHeight is the pixel height of a standard 80x25 text screen.
	CanvasHeight = DefaultRows * GlyphHeight

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

	startX := col * GlyphWidth
	startY := row * GlyphHeight
	glyphOffset := int(cell.Char) * GlyphBytes
	glyphData := fontVGA8x16[glyphOffset : glyphOffset+GlyphBytes]

	for y := range GlyphHeight {
		rowBits := glyphData[y]
		imgY := startY + y
		rowOffset := imgY * img.Stride

		for x := range GlyphWidth {
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
	cols := DefaultColumns
	rows := DefaultRows
	if len(grid) > 0 {
		rows = len(grid)
		if len(grid[0]) > 0 {
			cols = len(grid[0])
		}
	}

	img := image.NewNRGBA(image.Rect(0, 0, cols*GlyphWidth, rows*GlyphHeight))
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
