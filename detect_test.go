package ansimate_test

import (
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
