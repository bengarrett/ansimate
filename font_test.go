package ansimate_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/bengarrett/ansimate"
	"github.com/nalgeon/be"
)

func TestFontData(t *testing.T) {
	t.Parallel()
	be.Equal(t, ansimate.GlyphW, 8)
	be.Equal(t, ansimate.GlyphH, 16)
	be.Equal(t, ansimate.GlyphL, 16)
	be.Equal(t, ansimate.CanvasW, 640)
	be.Equal(t, ansimate.CanvasH, 400)
}

func TestPalette(t *testing.T) {
	t.Parallel()
	pal := ansimate.Palette
	be.Equal(t, len(pal), 16)
	// Black
	be.Equal(t, pal[0], color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	// White
	be.Equal(t, pal[15], color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	be.Equal(t, ansimate.BlinkOff.String(), "off")
	be.Equal(t, ansimate.BlinkText.String(), "blink")
	be.Equal(t, ansimate.BlinkICE.String(), "ice")
	be.Equal(t, ansimate.BlinkMode(99).String(), "unknown")
}

func TestDrawCell(t *testing.T) {
	t.Parallel()
	pal := ansimate.Palette
	img := image.NewNRGBA(image.Rect(0, 0, 16, 32))

	const fgWhite = 15
	const bgBlue = 1

	// Draw full block (0xDB)
	ansimate.DrawCell(img, 0, 0, ansimate.Cell{
		Char: 0xdb,
		FG:   fgWhite,
		BG:   bgBlue,
	}, pal)

	// All 8x16 pixels of cell (0,0) should be foreground (White)
	for y := range ansimate.GlyphH {
		for x := range ansimate.GlyphW {
			c := img.NRGBAAt(x, y)
			be.Equal(t, c, pal[fgWhite])
		}
	}

	// Draw space (0x20)
	ansimate.DrawCell(img, 1, 0, ansimate.Cell{
		Char: ' ',
		FG:   fgWhite,
		BG:   bgBlue,
	}, pal)

	// All 8x16 pixels of cell (1,0) should be background (Blue)
	for y := range ansimate.GlyphH {
		for x := range ansimate.GlyphW {
			c := img.NRGBAAt(ansimate.GlyphW+x, y)
			be.Equal(t, c, pal[bgBlue])
		}
	}
}

func TestDrawScreen(t *testing.T) {
	t.Parallel()
	pal := ansimate.Palette
	grid := make([][]ansimate.Cell, ansimate.Rows)
	for r := range grid {
		grid[r] = make([]ansimate.Cell, ansimate.Cols)
		for c := range grid[r] {
			grid[r][c] = ansimate.Cell{
				Char: ' ',
				FG:   7,
				BG:   0,
			}
		}
	}

	img := ansimate.DrawScreen(grid, pal)
	be.Equal(t, img.Bounds().Dx(), ansimate.CanvasW)
	be.Equal(t, img.Bounds().Dy(), ansimate.CanvasH)

	// Custom size grid
	const customRows, customCols = 10, 20
	custom := make([][]ansimate.Cell, customRows)
	for r := range custom {
		custom[r] = make([]ansimate.Cell, customCols)
	}
	customImg := ansimate.DrawScreen(custom, pal)
	be.Equal(t, customImg.Bounds().Dx(), customCols*ansimate.GlyphW)
	be.Equal(t, customImg.Bounds().Dy(), customRows*ansimate.GlyphH)
}
