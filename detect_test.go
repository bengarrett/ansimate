package ansimate_test

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/bengarrett/ansimate"
	"github.com/nalgeon/be"
)

func TestDetectNil(t *testing.T) {
	t.Parallel()
	isAnim, err := ansimate.Detect(nil)
	be.Err(t, err, ansimate.ErrNilReader)
	be.Equal(t, isAnim, false)
}

func TestDetectEmpty(t *testing.T) {
	t.Parallel()
	isAnim, err := ansimate.Detect(strings.NewReader(""))
	be.Err(t, err, nil)
	be.Equal(t, isAnim, false)

	be.Equal(t, ansimate.DetectB(nil), false)
	be.Equal(t, ansimate.DetectB([]byte{}), false)
}

func TestDetectStaticANSI(t *testing.T) {
	t.Parallel()
	// Static ANSI: 1 clear screen at start, draws downwards with no rewinds or overwrites
	staticArt := strings.Join([]string{
		"\x1b[2J\x1b[1;1H",
		"  ____  ____ ____  ",
		" | __ )| __ ) ___| ",
		" |  _ \\|  _ \\___ \\ ",
		" | |_) | |_) |__) |",
		" |____/|____/____/ ",
	}, "\r\n")

	isAnim, err := ansimate.Detect(strings.NewReader(staticArt))
	be.Err(t, err, nil)
	be.Equal(t, isAnim, false)
}

func TestDetectStaticTheDrawWrapped(t *testing.T) {
	t.Parallel()
	// Static ANSI saved by TheDraw using \x1b[s\r\n\x1b[u line-wraps and \x1b[nC forward moves
	var sb strings.Builder
	sb.WriteString("\x1b[40m\x1b[2J")
	for row := 1; row <= 20; row++ {
		sb.WriteString("\x1b[")
		sb.WriteString(strconv.Itoa(row))
		sb.WriteString("H\x1b[2C##############################\x1b[s\r\n\x1b[u##############################")
	}

	isAnim, err := ansimate.Detect(strings.NewReader(sb.String()))
	be.Err(t, err, nil)
	be.Equal(t, isAnim, false)
}

func TestDetectStaticLongScroll(t *testing.T) {
	t.Parallel()
	// Static text/ANSI stream taller than 25 rows scrolling downward
	var sb strings.Builder
	sb.WriteString("\x1b[2J")
	for range 60 {
		sb.WriteString("This is a normal static line of text that scrolls down the terminal.\r\n")
	}

	isAnim, err := ansimate.Detect(strings.NewReader(sb.String()))
	be.Err(t, err, nil)
	be.Equal(t, isAnim, false)
}

func TestDetectMultiClearAnimation(t *testing.T) {
	t.Parallel()
	// Multiple clears: \x1b[2J used for scene transitions
	animData := strings.Join([]string{
		"\x1b[2J\x1b[1;1HScene 1",
		"\x1b[2J\x1b[1;1HScene 2",
		"\x1b[2J\x1b[1;1HScene 3",
	}, "\r\n")

	isAnim, err := ansimate.Detect(strings.NewReader(animData))
	be.Err(t, err, nil)
	be.Equal(t, isAnim, true)
}

func TestDetectCursorChurnAnimation(t *testing.T) {
	t.Parallel()
	// Rewinds cursor and repeatedly overwrites characters
	var sb strings.Builder
	for range 15 {
		sb.WriteString("\x1b[5;10HOverwriting text in frame!")
	}

	isAnim, err := ansimate.Detect(strings.NewReader(sb.String()))
	be.Err(t, err, nil)
	be.Equal(t, isAnim, true)
}

func TestDetectSAUCE(t *testing.T) {
	t.Parallel()
	// Create dummy SAUCE record (128 bytes)
	createRecord := func(dataType, fileType byte) []byte {
		rec := make([]byte, 128)
		copy(rec[:5], "SAUCE")
		copy(rec[5:7], "00")
		rec[94] = dataType
		rec[95] = fileType

		return rec
	}

	// SAUCE DataType 1 (Character), FileType 2 (AnsiMation)
	sauceAnim := append([]byte("ANSI content"), createRecord(1, 2)...)
	be.Equal(t, ansimate.DetectB(sauceAnim), true)

	// SAUCE DataType 1 (Character), FileType 1 (Static ANSI)
	sauceStatic := append([]byte("ANSI content"), createRecord(1, 1)...)
	be.Equal(t, ansimate.DetectB(sauceStatic), false)
}

func TestDetectBinary(t *testing.T) {
	t.Parallel()

	// Data with internal NUL byte
	binaryData := []byte("RIFF\x16\xd9\x00\x00WEBPVP8X\x00\x00\x00\x00")
	isAnim, err := ansimate.Detect(bytes.NewReader(binaryData))
	be.Err(t, err, ansimate.ErrBinary)
	be.Equal(t, isAnim, false)
	be.Equal(t, ansimate.DetectB(binaryData), false)
}

func TestDetectPlainText(t *testing.T) {
	t.Parallel()

	// Plain ASCII text without ANSI escape sequences (\x1b[)
	plainText := "Hello world, this is a plain text file without any escape codes.\nLine 2\nLine 3\n"
	isAnim, err := ansimate.Detect(strings.NewReader(plainText))
	be.Err(t, err, nil)
	be.Equal(t, isAnim, false)
	be.Equal(t, ansimate.DetectB([]byte(plainText)), false)
}

func TestDetectSectorPadding(t *testing.T) {
	t.Parallel()

	// Valid static ANSI art with trailing DOS sector padding (0x1A and 0x00 bytes)
	paddedANSI := append([]byte("\x1b[2J\x1b[1;1HHello ANSI Art\r\n\x1a"), make([]byte, 64)...)
	isAnim, err := ansimate.Detect(bytes.NewReader(paddedANSI))
	be.Err(t, err, nil)
	be.Equal(t, isAnim, false)
	be.Equal(t, ansimate.DetectB(paddedANSI), false)
}

func TestDetectBlink(t *testing.T) {
	t.Parallel()

	// Static ANSI screen with blinking text
	blinkArt := "\x1b[2J\x1b[1;1HCall BBS \x1b[5m123-4567\x1b[0m"
	hasBlink, err := ansimate.DetectBlink(strings.NewReader(blinkArt))
	be.Err(t, err, nil)
	be.Equal(t, hasBlink, true)
	be.Equal(t, ansimate.DetectBlinkB([]byte(blinkArt)), true)

	// Static ANSI screen without blinking text
	staticArt := "\x1b[2J\x1b[1;1HCall BBS 123-4567"
	hasBlink, err = ansimate.DetectBlink(strings.NewReader(staticArt))
	be.Err(t, err, nil)
	be.Equal(t, hasBlink, false)
	be.Equal(t, ansimate.DetectBlinkB([]byte(staticArt)), false)

	// ANSImation with blinking text (should return false because it is an ANSImation, not static)
	animBlink := "\x1b[2J\x1b[1;1HScene 1 \x1b[5mBlink\x1b[0m\x1b[2J\x1b[1;1HScene 2"
	hasBlink, err = ansimate.DetectBlink(strings.NewReader(animBlink))
	be.Err(t, err, nil)
	be.Equal(t, hasBlink, false)
	be.Equal(t, ansimate.DetectBlinkB([]byte(animBlink)), false)
}
