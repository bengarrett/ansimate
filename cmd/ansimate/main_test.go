package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nalgeon/be"
)

func TestCLINoArgs(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := run(&buf)
	be.Err(t, err, errNoInput)
	be.True(t, strings.Contains(buf.String(), "Usage:"))
}

func TestCLIVersion(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := run(&buf, "-version")
	be.Err(t, err, nil)
	be.True(t, strings.Contains(buf.String(), "ansimate "))

	var shortBuf bytes.Buffer
	err = run(&shortBuf, "-v")
	be.Err(t, err, nil)
	be.True(t, strings.Contains(shortBuf.String(), "ansimate "))
}

func TestCLINonExistent(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := run(&buf, "/path/does/not/exist/test.ans")
	be.True(t, err != nil)
}

func TestCLICheck(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	ansFile := filepath.Join(tmpDir, "test.ans")
	err := os.WriteFile(ansFile, []byte("\x1b[2J\x1b[1;1HHello\x1b[2J\x1b[1;1HWorld"), 0o600)
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = run(&buf, "-check", ansFile)
	be.Err(t, err, nil)
	be.True(t, strings.Contains(buf.String(), "ANSImation detected"))

	// Test static ANSI art with blinking text
	blinkFile := filepath.Join(tmpDir, "blink.ans")
	err = os.WriteFile(blinkFile, []byte("\x1b[2J\x1b[1;1HCall BBS \x1b[5m555-1234\x1b[0m"), 0o600)
	be.Err(t, err, nil)

	var buf2 bytes.Buffer
	err = run(&buf2, "-check", blinkFile)
	be.Err(t, err, nil)
	be.True(t, strings.Contains(buf2.String(), "blinking text detected"))
}

func TestCLIConvert(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	ansFile := filepath.Join(tmpDir, "anim.ans")
	err := os.WriteFile(ansFile, []byte("\x1b[2J\x1b[1;1HFrame1\x1b[2J\x1b[1;1HFrame2"), 0o600)
	be.Err(t, err, nil)

	// Test with auto-generated output path (<input>.webp)
	var buf bytes.Buffer
	err = run(&buf, ansFile)
	be.Err(t, err, nil)
	expectedOut := filepath.Join(tmpDir, "anim.webp")
	st, err := os.Stat(expectedOut)
	be.Err(t, err, nil)
	be.True(t, st.Size() > 0)
	be.True(t, strings.Contains(buf.String(), "Done!"))

	// Test with explicit output path
	explicitOut := filepath.Join(tmpDir, "custom.webp")
	var buf2 bytes.Buffer
	err = run(&buf2, ansFile, explicitOut)
	be.Err(t, err, nil)
	st2, err := os.Stat(explicitOut)
	be.Err(t, err, nil)
	be.True(t, st2.Size() > 0)
}

func TestCLIFlagsAndErrors(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	ansFile := filepath.Join(tmpDir, "anim.ans")
	err := os.WriteFile(ansFile, []byte("\x1b[2J\x1b[1;1HFrame1\x1b[2J\x1b[1;1HFrame2"), 0o600)
	be.Err(t, err, nil)

	// Valid duration flags and monochrome palette
	monoOut := filepath.Join(tmpDir, "mono.webp")
	var buf bytes.Buffer
	err = run(&buf, "-delay", "3s", "-max", "15s", "-palette", "mono", ansFile, monoOut)
	be.Err(t, err, nil)
	st, err := os.Stat(monoOut)
	be.Err(t, err, nil)
	be.True(t, st.Size() > 0)

	// Invalid palette returns errUnknownPalette and does not create output file
	badOut := filepath.Join(tmpDir, "bad.webp")
	var badBuf bytes.Buffer
	err = run(&badBuf, "-palette", "invalid", ansFile, badOut)
	be.True(t, errors.Is(err, errUnknownPalette))
	_, statErr := os.Stat(badOut)
	be.True(t, os.IsNotExist(statErr))

	// Empty input fails and does not leave a 0-byte output file
	emptyFile := filepath.Join(tmpDir, "empty.ans")
	err = os.WriteFile(emptyFile, nil, 0o600)
	be.Err(t, err, nil)
	emptyOut := filepath.Join(tmpDir, "empty.webp")
	var emptyBuf bytes.Buffer
	err = run(&emptyBuf, emptyFile, emptyOut)
	be.True(t, err != nil)
	_, statErr = os.Stat(emptyOut)
	be.True(t, os.IsNotExist(statErr))

	// Binary file fails and does not produce an output file
	binFile := filepath.Join(tmpDir, "sample.bin")
	err = os.WriteFile(binFile, []byte("RIFF\x00\x00\x00\x00WEBP"), 0o600)
	be.Err(t, err, nil)
	binOut := filepath.Join(tmpDir, "sample.webp")
	var binBuf bytes.Buffer
	err = run(&binBuf, binFile, binOut)
	be.True(t, err != nil)
	_, statErr = os.Stat(binOut)
	be.True(t, os.IsNotExist(statErr))

	// Check on binary file reports binary
	var checkBuf bytes.Buffer
	err = run(&checkBuf, "-check", binFile)
	be.Err(t, err, nil)
	be.True(t, strings.Contains(checkBuf.String(), "Binary file (not ANSI art)"))
}
