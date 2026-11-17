package ansimate

import (
	"bytes"
	"fmt"
	"io"
	"slices"
)

const (
	minBacktracks = 5
	minOverwrites = 25
	minClears     = 2
	minChurn      = 100

	sauceRecordLen  = 128
	sauceTagLen     = 5
	sauceDataType   = 94
	sauceFileType   = 95
	sauceCharType   = 1
	sauceAnsimation = 2

	maxDetectBytes = 2 * 1024 * 1024 // 2 MB limit for stream detection
)

// Detect analyzes the reader and reports whether it is an ANSI stream ANSImation.
func Detect(r io.Reader) (bool, error) {
	if r == nil {
		return false, ErrNilReader
	}

	lr := io.LimitReader(r, maxDetectBytes)
	data, err := io.ReadAll(lr)
	if err != nil {
		return false, fmt.Errorf("read: %w", err)
	}

	if binary(data) {
		return false, ErrBinary
	}

	return DetectB(data), nil
}

// DetectB inspects the bytes and returns true if it represents an ANSImation.
func DetectB(data []byte) bool {
	// Detection runs in three stages:
	//
	// 1. Binary rejection (isBinary):
	//    Rejects binary data containing null bytes (0x00) in the body (excluding
	//    trailing SAUCE records or DOS sector padding).
	//
	// 2. SAUCE metadata (checkSAUCE):
	//    Inspects the trailing 128-byte SAUCE record (ignoring a trailing 0x1A EOF
	//    byte) for DataType 1 (Character) and FileType 2 (ANSImation).
	//
	// 3. Terminal stream heuristics (checkHeuristics):
	//
	//    Requires at least one ANSI CSI sequence (\x1b[) and simulates an 80x25
	//    ANSI.SYS grid—including cursor save/restore (\x1b[s and \x1b[u used by
	//    TheDraw to wrap file lines via \x1b[s\r\n\x1b[u), relative moves, line
	//    wrapping, erasing, and bottom-row scrolling—so static left-to-right,
	//    top-to-bottom streams produce zero backtracks or overwrites.
	//
	//    A stream is classified as an ANSImation as soon as any threshold is met:
	//    - Multi-clear transitions (clears >= 2): multiple \x1b[2J or 0x0C (Form
	//      Feed) screen clears.
	//    - Cell overwrites (backtracks >= 5 && overwrites >= 25): cursor rewinds to
	//      an earlier screen coordinate and overwrites existing non-space cells.
	//    - Choreographed construction (backtracks >= 100): non-linear drawing order
	//      (e.g. TheDraw animation mode) where characters are painted out of
	//      reading order even without overwriting previously drawn cells.
	if len(data) == 0 || binary(data) {
		return false
	}

	if checkSAUCE(data) {
		return true
	}

	return checkHeuristics(data)
}

// DetectBlink reports whether the reader is a static ANSI file containing visible blinking text.
func DetectBlink(r io.Reader) (bool, error) {
	if r == nil {
		return false, ErrNilReader
	}

	lr := io.LimitReader(r, maxDetectBytes)
	data, err := io.ReadAll(lr)
	if err != nil {
		return false, fmt.Errorf("read: %w", err)
	}

	return DetectBlinkB(data), nil
}

// DetectBlinkB reports whether data is a static ANSI file containing visible blinking text.
func DetectBlinkB(data []byte) bool {
	if len(data) == 0 || binary(data) || DetectB(data) {
		return false
	}

	term := newTerminal(Cols, Rows)
	_, _ = term.Write(data)

	return term.HasBlink()
}

// binary reports whether data appears to be binary rather than text or ANSI art.
// ANSI streams never contain null bytes (0x00) in the body, while binary formats do.
func binary(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Strip trailing SAUCE record (128 bytes, optionally preceded by 0x1A EOF)
	// because SAUCE binary fields may legitimately contain 0x00 bytes.
	body := data
	if len(body) >= sauceRecordLen {
		if bytes.HasPrefix(body[len(body)-sauceRecordLen:], []byte("SAUCE")) {
			body = body[:len(body)-sauceRecordLen]
		} else if len(body) > sauceRecordLen && body[len(body)-sauceRecordLen-1] == asciiEOF &&
			bytes.HasPrefix(body[len(body)-sauceRecordLen:], []byte("SAUCE")) {
			body = body[:len(body)-sauceRecordLen-1]
		}
	}

	// Strip trailing DOS EOF (0x1A) and sector padding (0x00, 0x20).
	body = bytes.TrimRight(body, "\x00\x1a\x20")

	return bytes.IndexByte(body, asciiNUL) != -1
}

func checkSAUCE(data []byte) bool {
	// Strip trailing DOS EOF character if present at the end
	if len(data) > 0 && data[len(data)-1] == asciiEOF {
		data = data[:len(data)-1]
	}

	if len(data) < sauceRecordLen {
		return false
	}

	record := data[len(data)-sauceRecordLen:]
	if !bytes.Equal(record[:sauceTagLen], []byte("SAUCE")) {
		return false
	}

	dataType := record[sauceDataType]
	fileType := record[sauceFileType]

	return dataType == sauceCharType && fileType == sauceAnsimation
}

type detector struct {
	grid         [Rows][Cols]byte
	params       []int
	x            int
	y            int
	savedX       int
	savedY       int
	lastWriteX   int
	lastWriteY   int
	clears       int
	backtracks   int
	overwrites   int
	pendingClear bool
	lineWrap     bool
	state        parserState
	curParam     int
	hasParam     bool
}

func checkHeuristics(data []byte) bool {
	if !bytes.Contains(data, []byte("\x1b[")) {
		return false
	}

	d := &detector{
		params:     make([]int, 0, brightOffset),
		lastWriteX: -1,
		lastWriteY: -1,
		lineWrap:   true,
	}

	return slices.ContainsFunc(data, d.process)
}

func (d *detector) process(b byte) bool {
	switch d.state {
	case stateNormal:
		d.normal(b)
	case stateEscape:
		d.escape(b)
	case stateCSI:
		d.csi(b)
	case stateCSIPrivate:
		d.private(b)
	}

	return d.isAnimation()
}

func (d *detector) private(b byte) {
	switch b {
	case 'h':
		d.lineWrap = true
		d.state = stateNormal
	case 'l':
		d.lineWrap = false
		d.state = stateNormal
	default:
		if (b < '0' || b > '9') && b != ';' {
			d.state = stateNormal
		}
	}
}

func (d *detector) isAnimation() bool {
	if d.clears >= minClears || d.backtracks >= minChurn {
		return true
	}

	return d.backtracks >= minBacktracks && d.overwrites >= minOverwrites
}

func (d *detector) normal(b byte) {
	if d.control(b) {
		return
	}
	d.writeChar(b)
}

func (d *detector) control(b byte) bool {
	switch b {
	case asciiNUL, asciiEOF:
	case asciiESC:
		d.state = stateEscape
	case asciiFF:
		d.pendingClear = true
		d.x = 0
		d.y = 0
	case asciiCR:
		d.x = 0
	case asciiLF:
		d.lineFeed()
	case asciiBS:
		if d.x > 0 {
			d.x--
		}
	case asciiTAB:
		d.tab()
	default:
		return false
	}

	return true
}

func (d *detector) lineFeed() {
	d.y++
	if d.y >= Rows {
		d.scrollUp()
		d.y = Rows - 1
	}
}

func (d *detector) scrollUp() {
	for y := range Rows - 1 {
		d.grid[y] = d.grid[y+1]
	}
	d.grid[Rows-1] = [Cols]byte{}
	if d.lastWriteY >= 0 {
		d.lastWriteY--
	}
}

func (d *detector) tab() {
	d.x = (d.x + tabStop) & ^(tabStop - 1)
	if d.x >= Cols {
		d.x = Cols - 1
	}
}

func (d *detector) escape(b byte) {
	if b == '[' {
		d.state = stateCSI
		d.params = d.params[:0]
		d.curParam = 0
		d.hasParam = false

		return
	}

	d.state = stateNormal
}

func (d *detector) csi(b byte) {
	switch {
	case b >= '0' && b <= '9':
		const base10 = 10
		d.curParam = d.curParam*base10 + int(b-'0')
		d.hasParam = true
	case b == ';':
		d.params = append(d.params, d.curParam)
		d.curParam = 0
		d.hasParam = false
	case b == '?':
		d.state = stateCSIPrivate
	default:
		if d.hasParam || len(d.params) > 0 {
			d.params = append(d.params, d.curParam)
		}
		d.dispatchCSI(b)
		d.state = stateNormal
	}
}

func (d *detector) dispatchCSI(cmd byte) {
	switch cmd {
	case 'H', 'f':
		d.cursorPosition()
	case 'A', 'B', 'C', 'D':
		d.cursorMove(cmd)
	case 'J':
		d.eraseDisplay()
	case 'K':
		d.eraseLine()
	case 's':
		d.savedX = d.x
		d.savedY = d.y
	case 'u':
		d.x = max(0, min(Cols-1, d.savedX))
		d.y = max(0, min(Rows-1, d.savedY))
	}
}

func (d *detector) cursorPosition() {
	row := 1
	col := 1
	if len(d.params) > 0 && d.params[0] > 0 {
		row = d.params[0]
	}
	if len(d.params) > 1 && d.params[1] > 0 {
		col = d.params[1]
	}

	d.y = max(0, min(Rows-1, row-1))
	d.x = max(0, min(Cols-1, col-1))
}

func (d *detector) cursorMove(cmd byte) {
	n := 1
	if len(d.params) > 0 && d.params[0] > 0 {
		n = d.params[0]
	}

	switch cmd {
	case 'A':
		d.y = max(0, d.y-n)
	case 'B':
		d.y = min(Rows-1, d.y+n)
	case 'C':
		d.x = min(Cols-1, d.x+n)
	case 'D':
		d.x = max(0, d.x-n)
	}
}

func (d *detector) eraseDisplay() {
	mode := eraseToEnd
	if len(d.params) > 0 {
		mode = d.params[0]
	}

	switch mode {
	case eraseToEnd:
		d.eraseDisplayToEnd()
	case eraseToStart:
		d.eraseDisplayToStart()
	case eraseAll:
		d.clearScreen()
	}
}

func (d *detector) clearScreen() {
	d.clears++
	d.grid = [Rows][Cols]byte{}
	d.x = 0
	d.y = 0
	d.lastWriteX = -1
	d.lastWriteY = -1
	d.pendingClear = false
}

func (d *detector) eraseDisplayToEnd() {
	for x := d.x; x < Cols; x++ {
		d.grid[d.y][x] = 0
	}
	for y := d.y + 1; y < Rows; y++ {
		d.grid[y] = [Cols]byte{}
	}
}

func (d *detector) eraseDisplayToStart() {
	for y := range d.y {
		d.grid[y] = [Cols]byte{}
	}
	for x := 0; x <= d.x && x < Cols; x++ {
		d.grid[d.y][x] = 0
	}
}

func (d *detector) eraseLine() {
	mode := eraseToEnd
	if len(d.params) > 0 {
		mode = d.params[0]
	}

	switch mode {
	case eraseToEnd:
		for x := d.x; x < Cols; x++ {
			d.grid[d.y][x] = 0
		}
	case eraseToStart:
		for x := 0; x <= d.x && x < Cols; x++ {
			d.grid[d.y][x] = 0
		}
	case eraseAll:
		d.grid[d.y] = [Cols]byte{}
	}
}

func (d *detector) writeChar(b byte) {
	if d.y < Rows && d.x < Cols {
		d.recordWrite(b)
	}

	d.x++
	if d.x >= Cols {
		if d.lineWrap {
			d.x = 0
			d.lineFeed()
		} else {
			d.x = Cols - 1
		}
	}
}

func (d *detector) recordWrite(b byte) {
	if d.pendingClear {
		d.clears++
		d.grid = [Rows][Cols]byte{}
		d.lastWriteX = -1
		d.lastWriteY = -1
		d.pendingClear = false
	}

	if d.lastWriteY >= 0 {
		if d.y < d.lastWriteY || (d.y == d.lastWriteY && d.x <= d.lastWriteX) {
			d.backtracks++
		}
	}
	d.lastWriteX = d.x
	d.lastWriteY = d.y

	existing := d.grid[d.y][d.x]
	if existing != 0 && existing != asciiSpace {
		d.overwrites++
	}
	d.grid[d.y][d.x] = b
}
