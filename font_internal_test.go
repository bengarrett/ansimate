package ansimate

import (
	"image"
	"image/color"
	"testing"

	"github.com/nalgeon/be"
)

func TestFontData(t *testing.T) {
	t.Parallel()
	be.Equal(t, GlyphW, 8)
	be.Equal(t, GlyphH, 16)
	be.Equal(t, GlyphL, 16)
	be.Equal(t, CanvasW, 640)
	be.Equal(t, CanvasH, 400)
}

func TestPalette(t *testing.T) {
	t.Parallel()
	pal := Palette
	be.Equal(t, len(pal), 16)
	// Black
	be.Equal(t, pal[0], color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	// White
	be.Equal(t, pal[15], color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	be.Equal(t, BlinkOff.String(), "off")
	be.Equal(t, BlinkText.String(), "blink")
	be.Equal(t, BlinkICE.String(), "ice")
	be.Equal(t, BlinkMode(99).String(), "unknown")
}

func TestDrawCell(t *testing.T) {
	t.Parallel()
	pal := Palette
	img := image.NewNRGBA(image.Rect(0, 0, 16, 32))

	const fgWhite = 15
	const bgBlue = 1

	// Draw full block (0xDB)
	drawCell(img, 0, 0, cell{
		Char: 0xdb,
		FG:   fgWhite,
		BG:   bgBlue,
	}, pal)

	// All 8x16 pixels of cell (0,0) should be foreground (White)
	for y := range GlyphH {
		for x := range GlyphW {
			c := img.NRGBAAt(x, y)
			be.Equal(t, c, pal[fgWhite])
		}
	}

	// Draw space (0x20)
	drawCell(img, 1, 0, cell{
		Char: ' ',
		FG:   fgWhite,
		BG:   bgBlue,
	}, pal)

	// All 8x16 pixels of cell (1,0) should be background (Blue)
	for y := range GlyphH {
		for x := range GlyphW {
			c := img.NRGBAAt(GlyphW+x, y)
			be.Equal(t, c, pal[bgBlue])
		}
	}
}

func TestDrawScreen(t *testing.T) {
	t.Parallel()
	pal := Palette
	grid := make([][]cell, Rows)
	for r := range grid {
		grid[r] = make([]cell, Cols)
		for c := range grid[r] {
			grid[r][c] = cell{
				Char: ' ',
				FG:   7,
				BG:   0,
			}
		}
	}

	img := drawScreen(grid, pal)
	be.Equal(t, img.Bounds().Dx(), CanvasW)
	be.Equal(t, img.Bounds().Dy(), CanvasH)

	// Custom size grid
	const customRows, customCols = 10, 20
	custom := make([][]cell, customRows)
	for r := range custom {
		custom[r] = make([]cell, customCols)
	}
	customImg := drawScreen(custom, pal)
	be.Equal(t, customImg.Bounds().Dx(), customCols*GlyphW)
	be.Equal(t, customImg.Bounds().Dy(), customRows*GlyphH)
}
