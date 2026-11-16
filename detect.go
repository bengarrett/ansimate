package ansimate

import (
	"bytes"
	"fmt"
	"io"
	"slices"
)

const (
	minClearsForAnim     = 2
	minBacktracksForAnim = 5
	minOverwritesForAnim = 25

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

	return DetectB(data), nil
}

// DetectB inspects the bytes and returns true if it represents an ANSImation.
func DetectB(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	if checkSAUCE(data) {
		return true
	}

	return checkHeuristics(data)
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
	grid       [Rows][Cols]byte
	params     []int
	x          int
	y          int
	clears     int
	backtracks int
	overwrites int
	state      parserState
	curParam   int
	hasParam   bool
}

func checkHeuristics(data []byte) bool {
	d := &detector{
		params: make([]int, 0, brightColorOffset),
	}

	return slices.ContainsFunc(data, d.processByte)
}

func (d *detector) processByte(b byte) bool {
	switch d.state {
	case stateNormal:
		d.processNormal(b)
	case stateEscape:
		d.processEscape(b)
	case stateCSI:
		d.processCSI(b)
	case stateCSIPrivate:
		if (b < '0' || b > '9') && b != ';' {
			d.state = stateNormal
		}
	}

	return d.isAnimation()
}

func (d *detector) isAnimation() bool {
	if d.clears >= minClearsForAnim {
		return true
	}

	return d.backtracks >= minBacktracksForAnim && d.overwrites >= minOverwritesForAnim
}

func (d *detector) processNormal(b byte) {
	if d.processControl(b) {
		return
	}
	d.writeChar(b)
}

func (d *detector) processControl(b byte) bool {
	switch b {
	case asciiNUL, asciiEOF:
	case asciiESC:
		d.state = stateEscape
	case asciiCR:
		if d.x > 0 {
			d.backtracks++
		}
		d.x = 0
	case asciiLF:
		d.lineFeed()
	case asciiBS:
		d.backspace()
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
		d.y = Rows - 1
	}
}

func (d *detector) backspace() {
	if d.x > 0 {
		d.x--
		d.backtracks++
	}
}

func (d *detector) tab() {
	d.x = (d.x + tabStop) & ^(tabStop - 1)
	if d.x >= Cols {
		d.x = Cols - 1
	}
}

func (d *detector) processEscape(b byte) {
	if b == '[' {
		d.state = stateCSI
		d.params = d.params[:0]
		d.curParam = 0
		d.hasParam = false

		return
	}

	d.state = stateNormal
}

func (d *detector) processCSI(b byte) {
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
	case 'A':
		d.cursorUp()
	case 'D':
		d.cursorBack()
	case 'J':
		d.eraseDisplay()
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

	newY := max(0, min(Rows-1, row-1))
	newX := max(0, min(Cols-1, col-1))

	if newY < d.y || (newY == d.y && newX < d.x) {
		d.backtracks++
	}

	d.y = newY
	d.x = newX
}

func (d *detector) cursorUp() {
	n := 1
	if len(d.params) > 0 && d.params[0] > 0 {
		n = d.params[0]
	}

	d.backtracks += n
	d.y = max(0, d.y-n)
}

func (d *detector) cursorBack() {
	n := 1
	if len(d.params) > 0 && d.params[0] > 0 {
		n = d.params[0]
	}

	d.backtracks += n
	d.x = max(0, d.x-n)
}

func (d *detector) eraseDisplay() {
	mode := eraseToEnd
	if len(d.params) > 0 {
		mode = d.params[0]
	}

	if mode == eraseAll {
		d.clears++
		d.x = 0
		d.y = 0
	}
}

func (d *detector) writeChar(b byte) {
	if d.y < Rows && d.x < Cols {
		existing := d.grid[d.y][d.x]
		if existing != 0 && existing != asciiSpace {
			d.overwrites++
		}
		d.grid[d.y][d.x] = b
	}

	d.x++
	if d.x >= Cols {
		d.x = 0
		d.y++
		if d.y >= Rows {
			d.y = Rows - 1
		}
	}
}
