package ansimate

import (
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
)

const (
	DefaultBaudRate    = 19200            // DefaultBaudRate is the simulated BBS transmission speed, higher is faster.
	DefaultFPS         = 20               // DefaultFPS is the default target frame rate.
	DefaultMaxDuration = 30 * time.Second // DefaultMaxDuration is the maximum playback time before stopping.
	DefaultEndDelay    = 2 * time.Second  // DefaultEndDelay is the pause time on the final frame before looping.

	bitsPerByte = 10
	maxFrames   = 500
)

// Config defines options for the animated WebP conversion.
type Config struct {
	BaudRate    int             // BaudRate is the simulated transmission speed in bits per second (default 19200).
	FPS         int             // FPS is the target frames per second (default 20).
	MaxDuration time.Duration   // MaxDuration is the maximum playback duration before capping (default 30s).
	EndDelay    time.Duration   // EndDelay is the duration held on the final frame before looping (default 2s).
	Columns     int             // Columns is the terminal width in character cells (default 80).
	Rows        int             // Rows is the terminal height in character cells (default 25).
	Blink       BlinkMode       // Blink sets the interpretation of ANSI SGR 5 (BlinkOff, BlinkText, BlinkICE).
	Lossy       bool            // Lossy enables lossy WebP compression (default false).
	Palette     [16]color.NRGBA // Palette is the 16-color palette to use. If unset, DefaultPalette() is used.
}

func (c *Config) sanitize() {
	if c.BaudRate <= 0 {
		c.BaudRate = DefaultBaudRate
	}
	if c.FPS <= 0 {
		c.FPS = DefaultFPS
	}
	if c.MaxDuration <= 0 {
		c.MaxDuration = DefaultMaxDuration
	}
	if c.EndDelay <= 0 {
		c.EndDelay = DefaultEndDelay
	}
	if c.Columns <= 0 {
		c.Columns = Cols
	}
	if c.Rows <= 0 {
		c.Rows = Rows
	}
	// Check if palette is uninitialized (alpha channel is 0)
	if c.Palette[15].A == 0 {
		c.Palette = Palette
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

	animation := c.animate(w)

	return animation.run(r)
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

	canvasW := c.Columns * GlyphW
	canvasH := c.Rows * GlyphH
	encoder := animation.NewEncoder(w, canvasW, canvasH, opts)

	bytesPerSec := c.BaudRate / bitsPerByte
	bytesPerTick := max(1, bytesPerSec/c.FPS)
	tick := time.Second / time.Duration(c.FPS)

	term := newTerminal(c.Columns, c.Rows, c.Blink)
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

func (a *animator) run(r io.Reader) error {
	buf := make([]byte, a.bytesPerTick)
	var duration time.Duration
	var bytesRead int

	for {
		n, err := r.Read(buf)
		if n > 0 {
			bytesRead += n
			_, _ = a.term.Write(buf[:n])
			if a.term.IsDirty() {
				errAdd := a.addFrame(a.tick)
				if errAdd != nil {
					return errAdd
				}
				duration += a.tick
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("read: %w", err)
		}

		if duration >= a.cfg.MaxDuration || a.frames >= maxFrames {
			break
		}
	}

	if bytesRead == 0 {
		return ErrEmpty
	}

	return a.finalize()
}

func (a *animator) addFrame(dur time.Duration) error {
	img := drawScreen(a.term.Grid(), a.cfg.Palette)
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
		img := drawScreen(a.term.Grid(), a.cfg.Palette)
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
