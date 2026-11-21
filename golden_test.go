package ansimate_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bengarrett/ansimate"
	"github.com/nalgeon/be"
)

func golden(tb testing.TB, tname, src, ext string, cfg *ansimate.Config) {
	tb.Helper()

	name := filepath.Join("testdata", src)
	goldenPath := filepath.Join("testdata", tname+ext)

	data, err := os.ReadFile(name)
	if err != nil {
		tb.Skip("testdata not found:", err)
	}

	goldenBytes, err := os.ReadFile(goldenPath)
	be.Err(tb, err, nil)

	// generate webp in the buffer
	var gotBuf bytes.Buffer
	err = ansimate.Animate(&gotBuf, bytes.NewReader(data), cfg)
	be.Err(tb, err, nil)

	got := gotBuf.Len() == len(goldenBytes)
	tb.Log("length of bytes", got)
	be.Equal(tb, got, true)

	// verify exact match against golden file
	got = bytes.Equal(gotBuf.Bytes(), goldenBytes)
	tb.Log("matching of bytes", got)
	be.Equal(tb, got, true)
}

func detect(tb testing.TB, src string, ansimation bool) {
	tb.Helper()

	name := filepath.Join("testdata", src)
	data, err := os.ReadFile(name)
	if err != nil {
		tb.Skip("testdata not found:", err)
	}

	check := ansimate.DetectB(data)
	be.Equal(tb, check, ansimation)
}

func TestGoldenDefaults(t *testing.T) {
	t.Parallel()

	// Chaos2 is the default ansimation offered on web/index.html.
	const src = "CHAOS2.ANS"

	tests := []struct {
		name string
		cfg  *ansimate.Config
	}{
		{
			name: "CHAOS2", cfg: &ansimate.Config{
				BaudRate:    19200,
				FPS:         20,
				MaxDuration: 0,
				EndDelay:    2 * time.Second,
				Columns:     80,
				Rows:        25,
				Lossy:       false,
				CGA:         false,
				Palette:     ansimate.PaletteVGA,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			golden(t, tc.name, src, ".webp", tc.cfg)
		})
	}

	detect(t, src, true)
}

func TestGoldenWebP(t *testing.T) {
	t.Parallel()

	// LS-RRINC.SDA is an ansimation.
	//  It is a test case for testing the ansimate configs as:
	//  - it is tiny at only 4.3K
	//  - the animation loop is brief
	//  - it is a 50 row piece with no scrolling
	//  - but on a 25 row screen it causes an animation artifact
	const src = "LS-RRINC.SDA"
	const base = "LS-RRINC-"

	tests := []struct {
		name string
		cfg  *ansimate.Config
	}{
		// palettes
		{name: base + "VGA", cfg: &ansimate.Config{Palette: ansimate.PaletteVGA}},
		{name: base + "Mono", cfg: &ansimate.Config{Palette: ansimate.PaletteMono}},
		{name: base + "Amber", cfg: &ansimate.Config{Palette: ansimate.PaletteAmber}},
		{name: base + "Green", cfg: &ansimate.Config{Palette: ansimate.PaletteGreen}},
		// baud rates
		{name: base + "1", cfg: &ansimate.Config{BaudRate: 1200}},
		{name: base + "9", cfg: &ansimate.Config{BaudRate: 9600}},
		{name: base + "115", cfg: &ansimate.Config{BaudRate: 115200}},
		// frames-per-second
		{name: base + "f5", cfg: &ansimate.Config{FPS: 5, BaudRate: 9600}},
		{name: base + "f30", cfg: &ansimate.Config{FPS: 30, BaudRate: 9600}},
		{name: base + "f70", cfg: &ansimate.Config{FPS: 70, BaudRate: 9600}},
		// loop delay / end delay
		{name: base + "1s", cfg: &ansimate.Config{EndDelay: 1 * time.Second}},
		{name: base + "10s", cfg: &ansimate.Config{EndDelay: 10 * time.Second}},
		// max duration
		{name: base + "1max", cfg: &ansimate.Config{MaxDuration: 1 * time.Second}},
		// lossy image compression
		{name: base + "lossy", cfg: &ansimate.Config{Lossy: true}},
		// ibm pc cga font
		{name: base + "cga", cfg: &ansimate.Config{CGA: true}},
		// cga font while locking the screen to 10 columns and 10 rows
		{name: base + "x10", cfg: &ansimate.Config{CGA: true, Columns: 10, Rows: 10}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			golden(t, tc.name, src, ".webp", tc.cfg)
		})
	}

	detect(t, src, true)
}

func TestGoldenANSI(t *testing.T) {
	t.Parallel()

	// DARKSPIR.LTD is a 40 row ansi text, but is not an ansimation.
	//  It is a good test case for a few reasons:
	//  - it is small at only 6.9K
	//  - requires scrolling when viewed in a standard 25 row screen
	//  - uses the blinking toggle while scrolling
	const src = "DARKSPIR.LTD"
	const base = "DARKSPIR"

	tests := []struct {
		name string
		cfg  *ansimate.Config
	}{
		{
			name: base + "1",
			cfg:  &ansimate.Config{BaudRate: 115200, FPS: 70, EndDelay: 10 * time.Second},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			golden(t, tc.name, src, ".webp", tc.cfg)
		})
	}

	detect(t, src, false)
}
