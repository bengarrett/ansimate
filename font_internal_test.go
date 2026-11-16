package ansimate

import (
	"image"
	"image/color"
	"testing"

	"github.com/nalgeon/be"
)

func TestFontData(t *testing.T) {
	t.Parallel()
	be.Equal(t, glyphW, 8)
	be.Equal(t, glyphH, 16)
	be.Equal(t, glyphL, 16)
	be.Equal(t, canvasW, 640)
	be.Equal(t, canvasH, 400)
}

func TestPalette(t *testing.T) {
	t.Parallel()
	pal := PaletteVGA
	be.Equal(t, len(pal), 16)
	// Black
	be.Equal(t, pal[0], color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	// White
	be.Equal(t, pal[15], color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	// IBM 8503 VGA Monochrome
	be.Equal(t, len(PaletteMono), 16)
	be.Equal(t, PaletteMono[0], color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	be.Equal(t, PaletteMono[15], color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	// P134 Amber Phosphor
	be.Equal(t, len(PaletteAmber), 16)
	be.Equal(t, PaletteAmber[0], color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	be.Equal(t, PaletteAmber[15], color.NRGBA{R: 255, G: 192, B: 0, A: 255})

	// IBM 5151 P39 Green Phosphor
	be.Equal(t, len(PaletteGreen), 16)
	be.Equal(t, PaletteGreen[0], color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	be.Equal(t, PaletteGreen[15], color.NRGBA{R: 0, G: 255, B: 0, A: 255})
}

func TestDrawCell(t *testing.T) {
	t.Parallel()
	pal := PaletteVGA
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
	for y := range glyphH {
		for x := range glyphW {
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
	for y := range glyphH {
		for x := range glyphW {
			c := img.NRGBAAt(glyphW+x, y)
			be.Equal(t, c, pal[bgBlue])
		}
	}
}

func TestDrawScreen(t *testing.T) {
	t.Parallel()
	pal := PaletteVGA
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
	be.Equal(t, img.Bounds().Dx(), canvasW)
	be.Equal(t, img.Bounds().Dy(), canvasH)

	// Custom size grid
	const customRows, customCols = 10, 20
	custom := make([][]cell, customRows)
	for r := range custom {
		custom[r] = make([]cell, customCols)
	}
	customImg := drawScreen(custom, pal)
	be.Equal(t, customImg.Bounds().Dx(), customCols*glyphW)
	be.Equal(t, customImg.Bounds().Dy(), customRows*glyphH)
}
