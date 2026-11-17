package ansimate

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"io"
	"time"

	// Register VP8 and VP8L encoders for WebP animation.
	_ "github.com/deepteams/webp"
	"github.com/deepteams/webp/animation"
)

var (
	ErrNilWriter = errors.New("writer is nil")
	ErrNilReader = errors.New("reader is nil")
	ErrEmpty     = errors.New("input stream is empty")
	ErrBinary    = errors.New("input is a binary file")
)

const (
	BaudRate    = 19200                  // BaudRate is the default simulated BBS transmission speed in bits per second.
	FPS         = 20                     // FPS is the default target frame rate.
	MaxDuration = 0                      // MaxDuration is 0 (unlimited duration by default, plays until EOF).
	EndDelay    = 2 * time.Second        // EndDelay is the default pause time on the final frame before looping.
	BlinkDelay  = 533 * time.Millisecond // BlinkDelay is the ~1.88 Hz IBM PC VGA hardware text blink rate.
	Cols        = 80                     // Cols is the default terminal width in character cells.
	Rows        = 25                     // Rows is the default terminal height in character cells.

	bitsPerByte = 10
	minBaudRate = 300              // Bell 103 300-baud modem floor.
	maxBaudRate = 115200           // PC 16550A UART hardware ceiling.
	maxFPS      = 70               // VGA 80x25 text-mode 70 Hz vertical refresh rate ceiling.
	maxEndDelay = 10 * time.Second // Prevents looping animations from appearing frozen.
	maxCols     = 160              // Bounds per-frame RGBA memory allocation (SVGA text mode limit).
	maxRows     = 100              // Bounds per-frame RGBA memory allocation (extended ANSI limit).
)

// Config defines options for the animated WebP conversion.
type Config struct {
	BaudRate    int             // BaudRate is the simulated transmission speed in bps (default 19200, range 300–115200).
	FPS         int             // FPS is the target frames per second (default 20, max 70).
	MaxDuration time.Duration   // MaxDuration is the maximum playback duration (default 0 for unlimited).
	EndDelay    time.Duration   // EndDelay is the pause held on the final frame (default 2s, max 10s).
	Columns     int             // Columns is the terminal width in character cells (default 80, max 160).
	Rows        int             // Rows is the terminal height in character cells (default 25, max 100).
	Lossy       bool            // Lossy enables lossy WebP compression (default false).
	CGA         bool            // CGA uses the 8x8 IBM CGA font (scanline-doubled to 4:3 aspect ratio).
	Palette     [16]color.NRGBA // Palette is optional (default PaletteVGA).
}

func (c *Config) sanitize() {
	if c.BaudRate <= 0 {
		c.BaudRate = BaudRate
	}
	c.BaudRate = min(max(c.BaudRate, minBaudRate), maxBaudRate)

	if c.FPS <= 0 {
		c.FPS = FPS
	}
	c.FPS = min(c.FPS, maxFPS)

	if c.MaxDuration < 0 {
		c.MaxDuration = 0
	}

	if c.EndDelay <= 0 {
		c.EndDelay = EndDelay
	}
	c.EndDelay = min(c.EndDelay, maxEndDelay)

	if c.Columns <= 0 {
		c.Columns = Cols
	}
	c.Columns = min(c.Columns, maxCols)

	if c.Rows <= 0 {
		c.Rows = Rows
	}
	c.Rows = min(c.Rows, maxRows)

	// Check if palette is uninitialized (alpha channel is 0)
	if c.Palette[15].A == 0 {
		c.Palette = PaletteVGA
	}
}

// Animate reads an ANSI animation from the reader and writes it as a WebP animation.
// The writer and readers are required, while the config is optional and can be left as nil.
func Animate(w io.Writer, r io.Reader, cfg *Config) error {
	if w == nil {
		return ErrNilWriter
	}
	if r == nil {
		return ErrNilReader
	}

	c := Config{}
	if cfg != nil {
		c = *cfg
	}
	c.sanitize()

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if len(data) == 0 {
		return ErrEmpty
	}
	if binary(data) {
		return ErrBinary
	}

	if !DetectB(data) {
		term := newTerminal(c.Columns, c.Rows)
		_, _ = term.Write(data)
		if term.HasBlink() {
			return c.animateBlink(w, term)
		}
	}

	animation := c.animate(w)

	return animation.run(bytes.NewReader(data))
}

type animator struct {
	cfg          *Config
	encoder      *animation.AnimEncoder
	term         *terminal
	tick         time.Duration
	bytesPerTick int
	frames       int
}

func (c *Config) animate(w io.Writer) *animator {
	opts := &animation.EncodeOptions{
		LoopCount:       0,
		Lossless:        !c.Lossy,
		BackgroundColor: c.Palette[0],
	}

	canvasW := c.Columns * glyphW
	canvasH := c.Rows * glyphH
	encoder := animation.NewEncoder(w, canvasW, canvasH, opts)

	bytesPerSec := c.BaudRate / bitsPerByte
	bytesPerTick := max(1, bytesPerSec/c.FPS)
	tick := time.Second / time.Duration(c.FPS)

	term := newTerminal(c.Columns, c.Rows)
	term.ResetDirty()

	return &animator{
		cfg:          c,
		encoder:      encoder,
		term:         term,
		tick:         tick,
		bytesPerTick: bytesPerTick,
		frames:       0,
	}
}

func (c *Config) animateBlink(w io.Writer, term *terminal) error {
	opts := &animation.EncodeOptions{
		LoopCount:       0,
		Lossless:        !c.Lossy,
		BackgroundColor: c.Palette[0],
	}

	canvasW := c.Columns * glyphW
	canvasH := c.Rows * glyphH
	encoder := animation.NewEncoder(w, canvasW, canvasH, opts)

	frameOn := drawScreenFont(term.Grid(), c.Palette, true, c.CGA)
	err := encoder.AddFrame(frameOn, BlinkDelay)
	if err != nil {
		return fmt.Errorf("add blink frame on: %w", err)
	}

	frameOff := drawScreenFont(term.Grid(), c.Palette, false, c.CGA)
	err = encoder.AddFrame(frameOff, BlinkDelay)
	if err != nil {
		return fmt.Errorf("add blink frame off: %w", err)
	}

	err = encoder.Close()
	if err != nil {
		return fmt.Errorf("close blink encoder: %w", err)
	}

	return nil
}

func (a *animator) run(r io.Reader) error {
	buf := make([]byte, a.bytesPerTick)
	var duration time.Duration
	var bytesRead int

	for {
		n, err := r.Read(buf)
		if n > 0 {
			bytesRead += n
			added, errStep := a.step(buf[:n], bytesRead)
			if errStep != nil {
				return errStep
			}
			duration += added
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("read: %w", err)
		}

		if a.cfg.MaxDuration > 0 && duration >= a.cfg.MaxDuration {
			break
		}
	}

	if bytesRead == 0 {
		return ErrEmpty
	}

	return a.finalize()
}

func (a *animator) step(chunk []byte, bytesRead int) (time.Duration, error) {
	err := a.processChunk(chunk, bytesRead)
	if err != nil {
		return 0, err
	}
	if a.term.IsDirty() {
		errAdd := a.addFrame(a.tick)
		if errAdd != nil {
			return 0, errAdd
		}

		return a.tick, nil
	}

	return 0, nil
}

func (a *animator) processChunk(chunk []byte, bytesRead int) error {
	if bytesRead <= sauceRecordLen {
		trimmed := bytes.TrimRight(chunk, "\x00\x1a\x20")
		if bytes.IndexByte(trimmed, asciiNUL) != -1 {
			return ErrBinary
		}
	}
	_, _ = a.term.Write(chunk)

	return nil
}

func (a *animator) addFrame(dur time.Duration) error {
	img := drawScreenFont(a.term.Grid(), a.cfg.Palette, true, a.cfg.CGA)
	err := a.encoder.AddFrame(img, dur)
	if err != nil {
		return fmt.Errorf("add frame: %w", err)
	}
	a.frames++
	a.term.ResetDirty()

	return nil
}

func (a *animator) finalize() error {
	if a.frames == 0 {
		return a.finalizeEmpty()
	}

	if a.cfg.EndDelay > 0 {
		img := drawScreenFont(a.term.Grid(), a.cfg.Palette, true, a.cfg.CGA)
		err := a.encoder.AddFrame(img, a.cfg.EndDelay)
		if err != nil {
			return fmt.Errorf("add final delay: %w", err)
		}
	}

	err := a.encoder.Close()
	if err != nil {
		return fmt.Errorf("close encoder: %w", err)
	}

	return nil
}

func (a *animator) finalizeEmpty() error {
	if !a.term.IsDirty() {
		return ErrEmpty
	}

	err := a.addFrame(a.cfg.EndDelay)
	if err != nil {
		return err
	}

	err = a.encoder.Close()
	if err != nil {
		return fmt.Errorf("close encoder: %w", err)
	}

	return nil
}
