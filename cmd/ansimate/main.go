// Package main provides the ansimate command-line interface.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bengarrett/ansimate"
)

const (
	defaultBaudRate = 19200
	defaultFPS      = 20
	defaultEndDelay = 2.0
	defaultMaxSec   = 30.0
	bytesInKB       = 1024
	bytesInMB       = 1024 * 1024
	version         = "0.1.0"
)

var errNoInput = errors.New("no input file specified")

type flags struct {
	output   string
	palette  string
	baudRate int
	fps      int
	delay    float64
	maxSec   float64
	lossy    bool
	check    bool
	ver      bool
}

func main() {
	err := run(os.Stdout, os.Args[1:]...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(w io.Writer, args ...string) error {
	fs := flag.NewFlagSet("ansimate", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.Usage = func() {
		printUsage(w, fs)
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
		fmt.Fprintf(w, "ansimate v%s\n", version)

		return nil
	}

	parsedArgs := fs.Args()
	if len(parsedArgs) == 0 {
		printUsage(w, fs)

		return errNoInput
	}

	inputFile := parsedArgs[0]
	if fl.check {
		return executeCheck(w, inputFile)
	}

	outputFile := determineOutput(inputFile, fl.output, parsedArgs)

	return executeConvert(w, inputFile, outputFile, &fl)
}

func setupFlags(fs *flag.FlagSet, fl *flags) {
	fs.StringVar(&fl.output, "o", "", "Destination WebP filename (defaults to <input>.webp)")
	fs.StringVar(&fl.output, "output", "", "Destination WebP filename (defaults to <input>.webp)")
	fs.StringVar(&fl.palette, "palette", "vga", "Color palette: vga, mono, amber, green")
	fs.IntVar(&fl.baudRate, "baud", defaultBaudRate, "Simulated BBS modem baud rate")
	fs.IntVar(&fl.fps, "fps", defaultFPS, "Target animation frame rate (FPS)")
	fs.Float64Var(&fl.delay, "delay", defaultEndDelay, "Seconds to hold final frame before looping")
	fs.Float64Var(&fl.maxSec, "max", defaultMaxSec, "Maximum animation duration cap in seconds")
	fs.BoolVar(&fl.lossy, "lossy", false, "Enable lossy WebP compression (default is lossless)")
	fs.BoolVar(&fl.check, "check", false, "Only detect if file is an ANSImation without converting")
	fs.BoolVar(&fl.ver, "v", false, "Display version")
	fs.BoolVar(&fl.ver, "version", false, "Display version")
}

func determineOutput(input, flagOut string, args []string) string {
	if len(args) > 1 && args[1] != "" {
		return args[1]
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

	isAnim, err := ansimate.Detect(f)
	if err != nil {
		return fmt.Errorf("detect: %w", err)
	}

	if isAnim {
		fmt.Fprintf(w, "%s: ANSImation detected\n", filepath.Base(input))
	} else {
		fmt.Fprintf(w, "%s: Static ANSI art\n", filepath.Base(input))
	}

	return nil
}

func executeConvert(w io.Writer, input, output string, fl *flags) error {
	f, err := os.Open(filepath.Clean(input))
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	reportDetection(w, f, input)

	out, err := os.Create(filepath.Clean(output))
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	cfg := buildConfig(fl)
	fmt.Fprintf(w, "Converting %s -> %s (%d baud, %d fps)...\n",
		filepath.Base(input), filepath.Base(output), cfg.BaudRate, cfg.FPS)

	start := time.Now()
	err = ansimate.Animate(out, f, cfg)
	if err != nil {
		return fmt.Errorf("animate: %w", err)
	}
	elapsed := time.Since(start)

	reportFinished(w, output, elapsed)

	return nil
}

func reportDetection(w io.Writer, f *os.File, input string) {
	isAnim, err := ansimate.Detect(f)
	if err == nil {
		if isAnim {
			fmt.Fprintf(w, "Detected: ANSImation in %s\n", filepath.Base(input))
		} else {
			fmt.Fprintf(w, "Notice: %s appears to be static ANSI (converting anyway)\n", filepath.Base(input))
		}
	}
	_, _ = f.Seek(0, io.SeekStart)
}

func buildConfig(fl *flags) *ansimate.Config {
	pal := ansimate.Palette
	switch strings.ToLower(fl.palette) {
	case "mono":
		pal = ansimate.PaletteMono
	case "amber":
		pal = ansimate.PaletteAmber
	case "green":
		pal = ansimate.PaletteGreen
	}

	return &ansimate.Config{
		BaudRate:    fl.baudRate,
		FPS:         fl.fps,
		MaxDuration: time.Duration(fl.maxSec * float64(time.Second)),
		EndDelay:    time.Duration(fl.delay * float64(time.Second)),
		Lossy:       fl.lossy,
		Palette:     pal,
	}
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

func printUsage(w io.Writer, fs *flag.FlagSet) {
	fmt.Fprintln(w, "ANSI-mate: Converts ANSI BBS animations into animated WebP images.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  ansimate [flags] <input.ans> [output.webp]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fs.PrintDefaults()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  ansimate animation.ans")
	fmt.Fprintln(w, "  ansimate animation.ans output.webp")
	fmt.Fprintln(w, "  ansimate -check animation.ans")
	fmt.Fprintln(w, "  ansimate -baud 28800 animation.ans")
}
