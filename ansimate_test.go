package ansimate_test

import (
	"bytes"
	"image"
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

func TestAnimateCGAEndDelay(t *testing.T) {
	t.Parallel()
	// Multi-frame animation across multiple ticks.
	ansiData := "\x1b[2J\x1b[1;1HFrame 1\r\n\x1b[2J\x1b[1;1HFrame 2"

	var outCGA bytes.Buffer
	cfgCGA := &ansimate.Config{
		BaudRate: 300, // 3 bytes per tick to ensure multiple content frames
		FPS:      10,
		EndDelay: 500 * time.Millisecond,
		CGA:      true,
	}
	err := ansimate.Animate(&outCGA, strings.NewReader(ansiData), cfgCGA)
	be.Err(t, err, nil)

	animCGA, err := animation.DecodeBytes(outCGA.Bytes())
	be.Err(t, err, nil)
	be.True(t, len(animCGA.Frames) >= 2)
	err = animCGA.DecodeFrames()
	be.Err(t, err, nil)

	// In CGA mode, the last frame must have its duration extended by EndDelay (500ms)
	// and its pixels must match CGA rendering, not VGA.
	lastFrame := animCGA.Frames[len(animCGA.Frames)-1]
	be.True(t, lastFrame.Duration >= 500*time.Millisecond)

	// Animate the same with CGA: false (VGA font).
	var outVGA bytes.Buffer
	cfgVGA := &ansimate.Config{
		BaudRate: 300,
		FPS:      10,
		EndDelay: 500 * time.Millisecond,
		CGA:      false,
	}
	err = ansimate.Animate(&outVGA, strings.NewReader(ansiData), cfgVGA)
	be.Err(t, err, nil)

	animVGA, err := animation.DecodeBytes(outVGA.Bytes())
	be.Err(t, err, nil)
	err = animVGA.DecodeFrames()
	be.Err(t, err, nil)
	lastFrameVGA := animVGA.Frames[len(animVGA.Frames)-1]

	// The CGA final frame and VGA final frame must have different pixels.
	be.True(t, !imagesEqual(lastFrame.Image, lastFrameVGA.Image))
}

func TestAnimateDelayPaddingAndSAUCE(t *testing.T) {
	t.Parallel()

	// Scene 1 followed by 300 bytes of \x1b[s delay padding before Scene 2.
	ansiBody := "\x1b[2J\x1b[1;1HScene 1" + strings.Repeat("\x1b[s", 100) + "\x1b[2J\x1b[1;1HScene 2"

	sauce := make([]byte, 128)
	copy(sauce[:5], "SAUCE")
	copy(sauce[5:7], "00")
	copy(sauce[7:42], "Test Title                         ")
	sauce[94] = 1
	sauce[95] = 2

	withSAUCE := append([]byte(ansiBody+"\x1a"), sauce...)

	cfg := &ansimate.Config{
		BaudRate: 19200,
		FPS:      20,
		EndDelay: 500 * time.Millisecond,
	}

	var outClean, outSAUCE bytes.Buffer
	be.Err(t, ansimate.Animate(&outClean, strings.NewReader(ansiBody), cfg), nil)
	be.Err(t, ansimate.Animate(&outSAUCE, bytes.NewReader(withSAUCE), cfg), nil)

	animClean, err := animation.DecodeBytes(outClean.Bytes())
	be.Err(t, err, nil)
	be.Err(t, animClean.DecodeFrames(), nil)

	animSAUCE, err := animation.DecodeBytes(outSAUCE.Bytes())
	be.Err(t, err, nil)
	be.Err(t, animSAUCE.DecodeFrames(), nil)

	// Scene 1 must hold for at least 3 ticks (150ms) across the \x1b[s padding.
	be.Equal(t, len(animSAUCE.Frames), 2)
	be.True(t, animSAUCE.Frames[0].Duration >= 150*time.Millisecond)

	// Stripping trailing SAUCE means the final rendered frame matches the clean input identically.
	be.True(t, imagesEqual(
		animSAUCE.Frames[len(animSAUCE.Frames)-1].Image,
		animClean.Frames[len(animClean.Frames)-1].Image,
	))
}

func TestAnimateFrameAndFrame(t *testing.T) {
	t.Parallel()

	ansiData := strings.Join([]string{
		"\x1b[2J\x1b[1;1HFrame 1",
		"\x1b[2J\x1b[1;1HFrame 2",
	}, "\r\n")

	var out bytes.Buffer
	cfg := &ansimate.Config{
		BaudRate: 9600,
		FPS:      10,
		EndDelay: 500 * time.Millisecond,
	}

	gotAnimFrame, err := ansimate.AnimateFrame(&out, strings.NewReader(ansiData), cfg)
	be.Err(t, err, nil)
	be.True(t, gotAnimFrame != nil)

	anim, err := animation.DecodeBytes(out.Bytes())
	be.Err(t, err, nil)
	be.Err(t, anim.DecodeFrames(), nil)
	be.True(t, imagesEqual(gotAnimFrame, anim.Frames[len(anim.Frames)-1].Image))

	gotFrame, err := ansimate.Frame(strings.NewReader(ansiData), cfg)
	be.Err(t, err, nil)
	be.True(t, imagesEqual(gotFrame, gotAnimFrame))
}

func TestAnimateFrameBlink(t *testing.T) {
	t.Parallel()

	ansiData := "\x1b[2J\x1b[1;1HHello \x1b[5mBlink\x1b[0m!"
	var out bytes.Buffer

	got, err := ansimate.AnimateFrame(&out, strings.NewReader(ansiData), nil)
	be.Err(t, err, nil)

	anim, err := animation.DecodeBytes(out.Bytes())
	be.Err(t, err, nil)
	be.Err(t, anim.DecodeFrames(), nil)
	be.True(t, imagesEqual(got, anim.Frames[0].Image))

	still, err := ansimate.Frame(strings.NewReader(ansiData), nil)
	be.Err(t, err, nil)
	be.True(t, imagesEqual(still, got))
}

func TestFrameErrors(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	_, err := ansimate.AnimateFrame(nil, strings.NewReader("hello"), nil)
	be.Err(t, err, ansimate.ErrNilWriter)

	_, err = ansimate.AnimateFrame(&buf, nil, nil)
	be.Err(t, err, ansimate.ErrNilReader)

	_, err = ansimate.Frame(nil, nil)
	be.Err(t, err, ansimate.ErrNilReader)

	_, err = ansimate.Frame(strings.NewReader(""), nil)
	be.Err(t, err, ansimate.ErrEmpty)

	_, err = ansimate.Frame(strings.NewReader("RIFF\x16\xd9\x00\x00WEBPVP8X"), nil)
	be.Err(t, err, ansimate.ErrBinary)
}

func imagesEqual(a, b image.Image) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Bounds() != b.Bounds() {
		return false
	}
	bnds := a.Bounds()
	for y := bnds.Min.Y; y < bnds.Max.Y; y++ {
		for x := bnds.Min.X; x < bnds.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}

	return true
}
