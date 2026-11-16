// Package main provides the ansimate command-line interface.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bengarrett/ansimate"
)

const (
	bytesInKB = 1024
	bytesInMB = 1024 * 1024
	web       = "https://github.com/bengarrett/ansimate"
)

var (
	errNoInput        = errors.New("no input file specified")
	errUnknownPalette = errors.New("unknown palette (valid: vga, mono, amber, green)")
)

var version string // populated by GoReleaser ldflags

type flags struct {
	output   string
	palette  string
	baudRate int
	fps      int
	delay    time.Duration
	maximum  time.Duration
	lossy    bool
	check    bool
	ver      bool
}

func main() {
	err := run(os.Stdout, os.Args[1:]...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Oops, %v\n", err)
		os.Exit(1)
	}
}

func run(w io.Writer, args ...string) error {
	fs := flag.NewFlagSet("ansimate", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.Usage = func() {
		printUsage(w)
	}

	var fl flags
	setupFlags(fs, &fl)

	err := fs.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return fmt.Errorf("parse flags: %w", err)
	}

	if fl.ver {
		fmt.Fprintf(w, "ansimate %s\n© 2026 Ben Garrett. All rights reserved. MIT license.\n%s\n",
			buildVersion(), web)

		return nil
	}

	arg := fs.Args()
	if len(arg) == 0 {
		printUsage(w)

		return errNoInput
	}

	input := arg[0]
	if fl.check {
		return executeCheck(w, input)
	}

	output := determine(input, fl.output, arg...)

	return execute(w, input, output, &fl)
}

func buildVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && s.Value != "" {
			const shortSHA = 7
			if len(s.Value) > shortSHA {
				return s.Value[:shortSHA]
			}

			return s.Value
		}
	}

	return "devel"
}

func setupFlags(fs *flag.FlagSet, fl *flags) {
	fs.StringVar(&fl.output, "o", "",
		"Destination WebP filename (defaults to <input>.webp)")
	fs.StringVar(&fl.output, "output", "",
		"Destination WebP filename (defaults to <input>.webp)")
	fs.StringVar(&fl.palette, "palette", "vga",
		"Color palette: vga, mono, amber, green")
	fs.IntVar(&fl.baudRate, "baud", ansimate.BaudRate,
		"Simulated BBS modem baud rate")
	fs.IntVar(&fl.fps, "fps", ansimate.FPS,
		"Target animation frame rate (FPS)")
	fs.DurationVar(&fl.delay, "delay", ansimate.EndDelay,
		"Duration to hold final frame before looping")
	fs.DurationVar(&fl.maximum, "max", ansimate.MaxDuration,
		"Maximum animation duration cap")
	fs.BoolVar(&fl.lossy, "lossy", false,
		"Enable lossy WebP compression (not recommended)")
	fs.BoolVar(&fl.check, "check", false,
		"Only detect if file is an ANSImation without converting")
	fs.BoolVar(&fl.ver, "v", false,
		"Display version")
	fs.BoolVar(&fl.ver, "version", false,
		"Display version")
}

func determine(input, flagOut string, arg ...string) string {
	if len(arg) > 1 && arg[1] != "" {
		return arg[1]
	}

	if flagOut != "" {
		return flagOut
	}

	ext := filepath.Ext(input)
	if ext != "" {
		return strings.TrimSuffix(input, ext) + ".webp"
	}

	return input + ".webp"
}

func executeCheck(w io.Writer, input string) error {
	f, err := os.Open(filepath.Clean(input))
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	ok, err := ansimate.Detect(f)
	if err != nil {
		return fmt.Errorf("detect: %w", err)
	}

	if ok {
		fmt.Fprintf(w, "%s: ANSImation detected\n", filepath.Base(input))
	} else {
		fmt.Fprintf(w, "%s: Static ANSI art\n", filepath.Base(input))
	}

	return nil
}

func execute(w io.Writer, input, output string, fl *flags) error {
	cfg, err := buildConfig(fl)
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Clean(input))
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	reportDetection(w, f, input)

	fmt.Fprintf(w, "Converting %s -> %s (%d baud, %d fps)...\n",
		filepath.Base(input), filepath.Base(output), cfg.BaudRate, cfg.FPS)

	var buf bytes.Buffer
	start := time.Now()
	err = ansimate.Animate(&buf, f, cfg)
	if err != nil {
		return fmt.Errorf("animate: %w", err)
	}
	elapsed := time.Since(start)

	const filePerm = 0o600
	err = os.WriteFile(filepath.Clean(output), buf.Bytes(), filePerm)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	reportFinished(w, output, elapsed)

	return nil
}

func reportDetection(w io.Writer, f *os.File, input string) {
	ok, err := ansimate.Detect(f)
	if err == nil {
		if ok {
			fmt.Fprintf(w, "Found an ANSImation in %s\n", filepath.Base(input))
		} else {
			fmt.Fprintf(w, "File: %s appears to be text or a static ANSI (converting anyway)\n", filepath.Base(input))
		}
	}
	_, _ = f.Seek(0, io.SeekStart)
}

func buildConfig(fl *flags) (*ansimate.Config, error) {
	var pal [16]color.NRGBA
	switch strings.ToLower(fl.palette) {
	case "", "vga":
		pal = ansimate.PaletteVGA
	case "mono":
		pal = ansimate.PaletteMono
	case "amber":
		pal = ansimate.PaletteAmber
	case "green":
		pal = ansimate.PaletteGreen
	default:
		return nil, fmt.Errorf("%w: %q", errUnknownPalette, fl.palette)
	}

	return &ansimate.Config{
		BaudRate:    fl.baudRate,
		FPS:         fl.fps,
		MaxDuration: fl.maximum,
		EndDelay:    fl.delay,
		Lossy:       fl.lossy,
		Palette:     pal,
	}, nil
}

func reportFinished(w io.Writer, output string, elapsed time.Duration) {
	st, err := os.Stat(filepath.Clean(output))
	if err == nil {
		fmt.Fprintf(w, "Done! Saved %s (%s) in %v\n",
			output, humanSize(st.Size()), elapsed.Round(time.Millisecond))

		return
	}

	fmt.Fprintf(w, "Done in %v\n", elapsed.Round(time.Millisecond))
}

func humanSize(b int64) string {
	switch {
	case b >= bytesInMB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(bytesInMB))
	case b >= bytesInKB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(bytesInKB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "ANSImate: Converts ANSI BBS animations into animated WebP images.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  ansimate [flags] <input.ans> [output.webp]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -o, -output <file>   Destination WebP filename (default \"<input>.webp\")")
	fmt.Fprintln(w, "  -palette <name>      Color palette: vga, mono, amber, green (default \"vga\")")
	fmt.Fprintln(w, "  -baud <int>          Simulated BBS modem baud rate (default 19200)")
	fmt.Fprintln(w, "  -fps <int>           Target animation frame rate (default 20)")
	fmt.Fprintln(w, "  -delay <duration>    Duration to hold final frame before looping (default 2s)")
	fmt.Fprintln(w, "  -max <duration>      Maximum animation duration cap (default 30s)")
	fmt.Fprintln(w, "  -lossy               Enable lossy WebP compression (not recommended)")
	fmt.Fprintln(w, "  -check               Only detect if file is an ANSImation without converting")
	fmt.Fprintln(w, "  -v, -version         Display version")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  ansimate animation.ans")
	fmt.Fprintln(w, "  ansimate animation.ans output.webp")
	fmt.Fprintln(w, "  ansimate -check animation.ans")
	fmt.Fprintln(w, "  ansimate -baud 28800 animation.ans")
}
