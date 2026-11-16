package ansimate

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"time"

	// Register pure Go VP8 and VP8L encoders for WebP animation.
	_ "github.com/deepteams/webp"
	"github.com/deepteams/webp/animation"
)

var (
	// ErrNilWriter is returned when the target writer is nil.
	ErrNilWriter = errors.New("writer is nil")
	// ErrNilReader is returned when the source reader is nil.
	ErrNilReader = errors.New("reader is nil")
	// ErrEmpty is returned when the input stream produces no frames.
	ErrEmpty = errors.New("input stream is empty")
)

const (
	// DefaultBaudRate is the standard simulated BBS transmission speed.
	DefaultBaudRate = 19200
	// DefaultFPS is the default target frame rate.
	DefaultFPS = 20
	// DefaultMaxDuration is the maximum playback time before stopping.
	DefaultMaxDuration = 15 * time.Second
	// DefaultEndDelay is the pause time on the final frame before looping.
	DefaultEndDelay = 2 * time.Second

	bitsPerByte = 10
	maxFrames   = 500
)

// Config defines options for ANSImation to Animated WebP conversion.
type Config struct {
	// BaudRate is the simulated transmission speed in bits per second (default 19200).
	BaudRate int
	// FPS is the target frames per second (default 20).
	FPS int
	// MaxDuration is the maximum playback duration before capping (default 15s).
	MaxDuration time.Duration
	// EndDelay is the duration held on the final frame before looping (default 2s).
	EndDelay time.Duration
	// Columns is the terminal width in character cells (default 80).
	Columns int
	// Rows is the terminal height in character cells (default 25).
	Rows int
	// Blink sets the interpretation of ANSI SGR 5 (BlinkOff, BlinkText, BlinkICE).
	Blink BlinkMode
	// Lossy enables lossy WebP compression (default false, meaning 100% pixel-perfect lossless).
	Lossy bool
	// Palette is the 16-color palette to use. If unset, DefaultPalette() is used.
	Palette [16]color.NRGBA
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
		c.Columns = DefaultColumns
	}
	if c.Rows <= 0 {
		c.Rows = DefaultRows
	}
	// Check if palette is uninitialized (alpha channel is 0)
	if c.Palette[15].A == 0 {
		c.Palette = DefaultPalette()
	}
}

// Animate reads an ANSI animation from r and writes an animated WebP to w.
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

	anim := newAnimator(w, &c)

	return anim.run(r)
}

type animator struct {
	cfg          *Config
	enc          *animation.AnimEncoder
	term         *Terminal
	tickDuration time.Duration
	bytesPerTick int
	frameCount   int
}

func newAnimator(w io.Writer, c *Config) *animator {
	opts := &animation.EncodeOptions{
		LoopCount:       0,
		Lossless:        !c.Lossy,
		BackgroundColor: c.Palette[0],
	}

	canvasW := c.Columns * GlyphWidth
	canvasH := c.Rows * GlyphHeight
	enc := animation.NewEncoder(w, canvasW, canvasH, opts)

	bytesPerSec := c.BaudRate / bitsPerByte
	bytesPerTick := max(1, bytesPerSec/c.FPS)
	tickDuration := time.Second / time.Duration(c.FPS)

	term := NewTerminal(c.Columns, c.Rows, c.Blink)
	term.ResetDirty()

	return &animator{
		cfg:          c,
		enc:          enc,
		term:         term,
		tickDuration: tickDuration,
		bytesPerTick: bytesPerTick,
		frameCount:   0,
	}
}

func (a *animator) run(r io.Reader) error {
	buf := make([]byte, a.bytesPerTick)
	var totalDuration time.Duration
	var bytesRead int

	for {
		n, err := r.Read(buf)
		if n > 0 {
			bytesRead += n
			_, _ = a.term.Write(buf[:n])
			if a.term.IsDirty() {
				errAdd := a.addFrame(a.tickDuration)
				if errAdd != nil {
					return errAdd
				}
				totalDuration += a.tickDuration
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("read: %w", err)
		}

		if totalDuration >= a.cfg.MaxDuration || a.frameCount >= maxFrames {
			break
		}
	}

	if bytesRead == 0 {
		return ErrEmpty
	}

	return a.finalize()
}

func (a *animator) addFrame(dur time.Duration) error {
	img := DrawScreen(a.term.Grid(), a.cfg.Palette)
	err := a.enc.AddFrame(img, dur)
	if err != nil {
		return fmt.Errorf("add frame: %w", err)
	}
	a.frameCount++
	a.term.ResetDirty()

	return nil
}

func (a *animator) finalize() error {
	if a.frameCount == 0 {
		return a.finalizeEmpty()
	}

	if a.cfg.EndDelay > 0 {
		img := DrawScreen(a.term.Grid(), a.cfg.Palette)
		err := a.enc.AddFrame(img, a.cfg.EndDelay)
		if err != nil {
			return fmt.Errorf("add final delay: %w", err)
		}
	}

	err := a.enc.Close()
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

	err = a.enc.Close()
	if err != nil {
		return fmt.Errorf("close encoder: %w", err)
	}

	return nil
}
