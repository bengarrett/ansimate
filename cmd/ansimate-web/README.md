# ANSImate Web

The `ansimate-web` directory is a web browser port of ANSImate, allowing the converting of ANSI art into animated WebP images entirely client-side without a server.

#### Building

```bash
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o ansimate-web.wasm ./cmd/ansimate-web
```

Or using Task:

```bash
task build-web
```

#### JS Usage

Once loaded, the module exposes a global a object:

- `window.ansimate`
- `globalThis.ansimate`

##### Asynchronously converts ANSI bytes into WebP animation bytes.

`ansimate.convert(input, [options])` -> `Promise<Uint8Array>`

- **`input`**: `Uint8Array`, `ArrayBuffer`, or raw ANSI `string`.
- **`options`** _(optional object)_:
  - `baudRate` _(number)_: Simulated transmission speed in bps (e.g. `19200`, `57600`, `115200`).
  - `fps` _(number)_: Target frames per second (default: `20`, max: `70`).
  - `delay` _(number)_: Final frame hold duration in milliseconds before loop (default: `2000`).
  - `maximum` _(number)_: Maximum animation cap in milliseconds (default: `0` for unlimited).
  - `lossy` _(boolean)_: Enable lossy WebP compression (default: `false`).
  - `cga` _(boolean)_: Use 8x8 CGA font instead of 8x16 VGA font (default: `false`).
  - `palette` _(string)_: Color palette (`"vga"`, `"mono"`, `"amber"`, `"green"`).
  - `columns` _(number)_: Terminal character width (default: `80`).
  - `rows` _(number)_: Terminal character height (default: `25`).
- **Returns**: `Promise<Uint8Array>` containing the encoded WebP file.

##### Synchronous version of `convert`. Throws an `Error` on invalid input or conversion failure.

`ansimate.convertSync(input, [options])` -> `Uint8Array`

##### Detect whether the provided input is an ANSImation and whether it is binary data.

`ansimate.detect(input) -> { isAnimation: boolean, isBinary: boolean }`

##### Return the ansimate version string.

`ansimate.version() -> string`

---

#### Example

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>ANSImate</title>
    <script src="wasm_exec.js"></script>
  </head>
  <body>
    <input type="file" id="fileInput" accept=".ans,.txt" />
    <img id="preview" alt="Animation Preview" />

    <script>
      const go = new Go();
      WebAssembly.instantiateStreaming(
        fetch('ansimate-web.wasm'),
        go.importObject
      ).then((result) => {
        go.run(result.instance);
        console.log('ANSImate loaded:', ansimate.version());
      });

      document
        .getElementById('fileInput')
        .addEventListener('change', async (e) => {
          const file = e.target.files[0];
          if (!file) return;

          const buffer = new Uint8Array(await file.arrayBuffer());
          const webpBytes = await ansimate.convert(buffer, {
            palette: 'amber',
            baudRate: 57600,
          });

          const blob = new Blob([webpBytes], { type: 'image/webp' });
          document.getElementById('preview').src = URL.createObjectURL(blob);
        });
    </script>
  </body>
</html>
```
