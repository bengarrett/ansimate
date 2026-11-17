package ansimate_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/bengarrett/ansimate"
	"github.com/deepteams/webp/animation"
	"github.com/nalgeon/be"
)

const glyphW, glyphH = 8, 16

func TestAnimateNil(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := ansimate.Animate(nil, strings.NewReader("hello"), nil)
	be.Err(t, err, ansimate.ErrNilWriter)

	err = ansimate.Animate(&buf, nil, nil)
	be.Err(t, err, ansimate.ErrNilReader)
}

func TestAnimateEmpty(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := ansimate.Animate(&buf, strings.NewReader(""), nil)
	be.Err(t, err, ansimate.ErrEmpty)
}

func TestAnimateBinary(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := ansimate.Animate(&buf, strings.NewReader("RIFF\x16\xd9\x00\x00WEBPVP8X"), nil)
	be.Err(t, err, ansimate.ErrBinary)
}

func TestAnimateBasic(t *testing.T) {
	t.Parallel()
	// Ansimation stream: write title, clear screen, write frame 2, clear screen, write frame 3
	ansiData := strings.Join([]string{
		"\x1b[2J\x1b[1;1HFrame 1: Loading...",
		"\x1b[2J\x1b[1;1HFrame 2: Processing...",
		"\x1b[2J\x1b[1;1HFrame 3: Completed!",
	}, "\r\n")

	var out bytes.Buffer
	cfg := &ansimate.Config{
		BaudRate:    9600,
		FPS:         10,
		MaxDuration: 5 * time.Second,
		EndDelay:    1 * time.Second,
	}

	err := ansimate.Animate(&out, strings.NewReader(ansiData), cfg)
	be.Err(t, err, nil)
	be.True(t, out.Len() > 0)

	// Decode the generated animated WebP using deepteams/webp/animation
	anim, err := animation.DecodeBytes(out.Bytes())
	be.Err(t, err, nil)
	be.Equal(t, anim.CanvasWidth, ansimate.Cols*glyphW)
	be.Equal(t, anim.CanvasHeight, ansimate.Rows*glyphH)
	be.Equal(t, anim.LoopCount, 0)
	be.True(t, len(anim.Frames) >= 1)
}

func TestAnimateBlink(t *testing.T) {
	t.Parallel()
	// Static ANSI screen with blinking text
	ansiData := "\x1b[2J\x1b[1;1HWelcome to the BBS! Call \x1b[5m555-1234\x1b[0m today!"

	var out bytes.Buffer
	err := ansimate.Animate(&out, strings.NewReader(ansiData), nil)
	be.Err(t, err, nil)
	be.True(t, out.Len() > 0)

	// Verify it generated exactly a 2-frame infinite loop animation
	anim, err := animation.DecodeBytes(out.Bytes())
	be.Err(t, err, nil)
	be.Equal(t, anim.LoopCount, 0)
	be.Equal(t, len(anim.Frames), 2)
}

func TestAnimateCustomDimensions(t *testing.T) {
	t.Parallel()
	const customCols, customRows = 40, 15
	ansiData := "\x1b[1;1HHello Custom Terminal!\r\n\x1b[2;1HRow 2 Content"

	var out bytes.Buffer
	cfg := &ansimate.Config{
		Columns:  customCols,
		Rows:     customRows,
		FPS:      10,
		BaudRate: 19200,
		EndDelay: 500 * time.Millisecond,
	}

	err := ansimate.Animate(&out, strings.NewReader(ansiData), cfg)
	be.Err(t, err, nil)

	anim, err := animation.DecodeBytes(out.Bytes())
	be.Err(t, err, nil)
	be.Equal(t, anim.CanvasWidth, customCols*glyphW)
	be.Equal(t, anim.CanvasHeight, customRows*glyphH)
}

func TestAnimateClampedBounds(t *testing.T) {
	t.Parallel()
	const (
		oversizedCols, oversizedRows = 500, 500
		clampedCols, clampedRows     = 160, 100
	)
	ansiData := "\x1b[1;1HClamped bounds"

	var out bytes.Buffer
	cfg := &ansimate.Config{
		Columns: oversizedCols,
		Rows:    oversizedRows,
	}

	err := ansimate.Animate(&out, strings.NewReader(ansiData), cfg)
	be.Err(t, err, nil)

	anim, err := animation.DecodeBytes(out.Bytes())
	be.Err(t, err, nil)
	be.Equal(t, anim.CanvasWidth, clampedCols*glyphW)
	be.Equal(t, anim.CanvasHeight, clampedRows*glyphH)
}
