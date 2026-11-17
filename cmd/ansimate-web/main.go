//go:build js && wasm

// Package main provides the WebAssembly in-browser API for ansimate.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"syscall/js"
	"time"

	"github.com/bengarrett/ansimate"
)

var (
	errNoInput = errors.New("input data is required")
	errUnknown = errors.New("unknown palette (valid: vga, mono, amber, green)")
)

//nolint:gochecknoglobals
var (
	version = "devel"
	date    = ""
)

func main() {
	ans := js.Global().Get("Object").New()
	ans.Set("convert", js.FuncOf(convertPromise))
	ans.Set("convertSync", js.FuncOf(convertSync))
	ans.Set("detect", js.FuncOf(detect))
	ans.Set("version", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		if date != "" {
			return fmt.Sprintf("%s (%s)", version, date)
		}
		return version
	}))

	js.Global().Set("ansimate", ans)

	// Keep runtime active in browser/JS environment
	select {}
}

func convertPromise(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return promiseReject(errNoInput)
	}

	inputVal := args[0]
	var optsVal js.Value
	if len(args) > 1 {
		optsVal = args[1]
	}

	handler := js.FuncOf(func(_ js.Value, pArgs []js.Value) any {
		resolve := pArgs[0]
		reject := pArgs[1]

		go func() {
			input, err := getInputBytes(inputVal)
			if err != nil {
				reject.Invoke(js.Global().Get("Error").New(err.Error()))
				return
			}

			cfg, err := parseConfig(optsVal)
			if err != nil {
				reject.Invoke(js.Global().Get("Error").New(err.Error()))
				return
			}

			var out bytes.Buffer
			if err := ansimate.Animate(&out, bytes.NewReader(input), cfg); err != nil {
				reject.Invoke(js.Global().Get("Error").New(err.Error()))
				return
			}

			resolve.Invoke(toUint8Array(out.Bytes()))
		}()

		return nil
	})

	return js.Global().Get("Promise").New(handler)
}

func convertSync(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		panic(js.Global().Get("Error").New(errNoInput.Error()))
	}

	input, err := getInputBytes(args[0])
	if err != nil {
		panic(js.Global().Get("Error").New(err.Error()))
	}

	var cfg *ansimate.Config
	if len(args) > 1 {
		cfg, err = parseConfig(args[1])
		if err != nil {
			panic(js.Global().Get("Error").New(err.Error()))
		}
	}

	var out bytes.Buffer
	if err := ansimate.Animate(&out, bytes.NewReader(input), cfg); err != nil {
		panic(js.Global().Get("Error").New(err.Error()))
	}

	return toUint8Array(out.Bytes())
}

func detect(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		panic(js.Global().Get("Error").New(errNoInput.Error()))
	}

	input, err := getInputBytes(args[0])
	if err != nil {
		panic(js.Global().Get("Error").New(err.Error()))
	}

	isAnim, err := ansimate.Detect(bytes.NewReader(input))
	isBinary := false
	if err != nil && errors.Is(err, ansimate.ErrBinary) {
		isBinary = true
		err = nil
	}
	if err != nil {
		panic(js.Global().Get("Error").New(err.Error()))
	}

	res := js.Global().Get("Object").New()
	res.Set("isAnimation", isAnim)
	res.Set("isBinary", isBinary)
	return res
}

func getInputBytes(val js.Value) ([]byte, error) {
	if val.IsUndefined() || val.IsNull() {
		return nil, errNoInput
	}

	if val.Type() == js.TypeString {
		return []byte(val.String()), nil
	}

	if val.InstanceOf(js.Global().Get("ArrayBuffer")) {
		val = js.Global().Get("Uint8Array").New(val)
	}

	if val.InstanceOf(js.Global().Get("Uint8Array")) {
		length := val.Get("length").Int()
		dst := make([]byte, length)
		js.CopyBytesToGo(dst, val)
		return dst, nil
	}

	return nil, errors.New("input must be a Uint8Array, ArrayBuffer, or string")
}

func parseConfig(opts js.Value) (*ansimate.Config, error) {
	if opts.IsUndefined() || opts.IsNull() {
		return nil, nil
	}
	if opts.Type() != js.TypeObject {
		return nil, errors.New("options must be an object")
	}

	cfg := &ansimate.Config{}

	if v := opts.Get("baudRate"); !v.IsUndefined() && !v.IsNull() {
		cfg.BaudRate = v.Int()
	}
	if v := opts.Get("fps"); !v.IsUndefined() && !v.IsNull() {
		cfg.FPS = v.Int()
	}
	if v := opts.Get("delay"); !v.IsUndefined() && !v.IsNull() {
		// Milliseconds
		cfg.EndDelay = time.Duration(v.Float() * float64(time.Millisecond))
	}
	if v := opts.Get("maximum"); !v.IsUndefined() && !v.IsNull() {
		// Milliseconds
		cfg.MaxDuration = time.Duration(v.Float() * float64(time.Millisecond))
	}
	if v := opts.Get("lossy"); !v.IsUndefined() && !v.IsNull() {
		cfg.Lossy = v.Bool()
	}
	if v := opts.Get("cga"); !v.IsUndefined() && !v.IsNull() {
		cfg.CGA = v.Bool()
	}
	if v := opts.Get("columns"); !v.IsUndefined() && !v.IsNull() {
		cfg.Columns = v.Int()
	}
	if v := opts.Get("rows"); !v.IsUndefined() && !v.IsNull() {
		cfg.Rows = v.Int()
	}
	if v := opts.Get("palette"); !v.IsUndefined() && !v.IsNull() {
		palStr := strings.ToLower(v.String())
		switch palStr {
		case "", "vga":
			cfg.Palette = ansimate.PaletteVGA
		case "mono":
			cfg.Palette = ansimate.PaletteMono
		case "amber":
			cfg.Palette = ansimate.PaletteAmber
		case "green":
			cfg.Palette = ansimate.PaletteGreen
		default:
			return nil, fmt.Errorf("%w: %q", errUnknown, palStr)
		}
	}

	return cfg, nil
}

func toUint8Array(b []byte) js.Value {
	ua := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(ua, b)
	return ua
}

func promiseReject(err error) js.Value {
	handler := js.FuncOf(func(_ js.Value, pArgs []js.Value) any {
		reject := pArgs[1]
		reject.Invoke(js.Global().Get("Error").New(err.Error()))
		return nil
	})
	return js.Global().Get("Promise").New(handler)
}
