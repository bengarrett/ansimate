package ansimate

const (
	asciiNUL   = 0x00
	asciiBS    = 0x08
	asciiTAB   = 0x09
	asciiLF    = 0x0a
	asciiCR    = 0x0d
	asciiEOF   = 0x1a
	asciiESC   = 0x1b
	asciiSpace = 0x20

	defaultFGColor    = 7
	defaultBGColor    = 0
	brightColorOffset = 8
	tabStop           = 8

	sgrReset     = 0
	sgrBold      = 1
	sgrBlink     = 5
	sgrBoldOff   = 22
	sgrBlinkOff  = 25
	sgrFGStart   = 30
	sgrFGEnd     = 37
	sgrFGDefault = 39
	sgrBGStart   = 40
	sgrBGEnd     = 47
	sgrBGDefault = 49
	sgrHiFGStart = 90
	sgrHiFGEnd   = 97
	sgrHiBGStart = 100
	sgrHiBGEnd   = 107

	eraseToEnd   = 0
	eraseToStart = 1
	eraseAll     = 2
)

type parserState int

const (
	stateNormal parserState = iota
	stateEscape
	stateCSI
	stateCSIPrivate
)

// Terminal simulates a fixed-dimension character grid with ANSI escape code processing.
type Terminal struct {
	grid      [][]Cell
	params    []int
	cols      int
	rows      int
	x         int
	y         int
	savedX    int
	savedY    int
	curFG     byte
	curBG     byte
	bold      bool
	blink     bool
	lineWrap  bool
	dirty     bool
	state     parserState
	curParam  int
	hasParam  bool
	blinkMode BlinkMode
}

// NewTerminal creates a new terminal instance with specified columns, rows, and blink mode.
func NewTerminal(cols, rows int, blink BlinkMode) *Terminal {
	if cols <= 0 {
		cols = DefaultColumns
	}
	if rows <= 0 {
		rows = DefaultRows
	}

	t := &Terminal{
		cols:      cols,
		rows:      rows,
		curFG:     defaultFGColor,
		curBG:     defaultBGColor,
		lineWrap:  true,
		blinkMode: blink,
		grid:      make([][]Cell, rows),
		params:    make([]int, 0, brightColorOffset),
	}

	t.Clear()

	return t
}

// Clear clears the terminal grid and resets the cursor to (0, 0).
func (t *Terminal) Clear() {
	blank := Cell{Char: ' ', FG: defaultFGColor, BG: defaultBGColor}
	for r := range t.rows {
		t.grid[r] = make([]Cell, t.cols)
		for c := range t.cols {
			t.grid[r][c] = blank
		}
	}
	t.x = 0
	t.y = 0
	t.dirty = true
}

// Grid returns the current 2D cell grid of the terminal.
func (t *Terminal) Grid() [][]Cell {
	return t.grid
}

// IsDirty reports whether any cell on the screen has changed since the last ResetDirty.
func (t *Terminal) IsDirty() bool {
	return t.dirty
}

// ResetDirty clears the dirty flag.
func (t *Terminal) ResetDirty() {
	t.dirty = false
}

// Write processes a byte slice of ANSI data.
func (t *Terminal) Write(p []byte) (int, error) {
	for _, b := range p {
		t.processByte(b)
	}

	return len(p), nil
}

// WriteByte processes a single byte.
func (t *Terminal) WriteByte(b byte) error {
	t.processByte(b)

	return nil
}

func (t *Terminal) processByte(b byte) {
	switch t.state {
	case stateNormal:
		t.processNormal(b)
	case stateEscape:
		t.processEscape(b)
	case stateCSI:
		t.processCSI(b)
	case stateCSIPrivate:
		t.processCSIPrivate(b)
	}
}

func (t *Terminal) processNormal(b byte) {
	switch b {
	case asciiNUL, asciiEOF:
		// ignore
	case asciiESC:
		t.state = stateEscape
	case asciiCR:
		t.x = 0
	case asciiLF:
		t.lineFeed()
	case asciiBS:
		if t.x > 0 {
			t.x--
		}
	case asciiTAB:
		t.tab()
	default:
		t.writeChar(b)
	}
}

func (t *Terminal) processEscape(b byte) {
	if b == '[' {
		t.state = stateCSI
		t.params = t.params[:0]
		t.curParam = 0
		t.hasParam = false

		return
	}

	// Unrecognized escape character, return to normal state
	t.state = stateNormal
}

func (t *Terminal) processCSI(b byte) {
	switch {
	case b >= '0' && b <= '9':
		const base10 = 10
		t.curParam = t.curParam*base10 + int(b-'0')
		t.hasParam = true
	case b == ';':
		t.params = append(t.params, t.curParam)
		t.curParam = 0
		t.hasParam = false
	case b == '?':
		t.state = stateCSIPrivate
	default:
		if t.hasParam || len(t.params) > 0 {
			t.params = append(t.params, t.curParam)
		}
		t.dispatchCSI(b)
		t.state = stateNormal
	}
}

func (t *Terminal) processCSIPrivate(b byte) {
	switch b {
	case 'h':
		// \x1b[?7h: enable line wrap
		t.lineWrap = true
		t.state = stateNormal
	case 'l':
		// \x1b[?7l: disable line wrap
		t.lineWrap = false
		t.state = stateNormal
	default:
		if (b < '0' || b > '9') && b != ';' {
			t.state = stateNormal
		}
	}
}

func (t *Terminal) dispatchCSI(cmd byte) {
	switch cmd {
	case 'H', 'f':
		t.cursorPosition()
	case 'A', 'B', 'C', 'D':
		t.cursorMove(cmd)
	case 'J':
		t.eraseDisplay()
	case 'K':
		t.eraseLine()
	case 's':
		t.savedX = t.x
		t.savedY = t.y
	case 'u':
		t.x = max(0, min(t.cols-1, t.savedX))
		t.y = max(0, min(t.rows-1, t.savedY))
	case 'm':
		t.sgr()
	}
}

func (t *Terminal) cursorPosition() {
	row := 1
	col := 1
	if len(t.params) > 0 && t.params[0] > 0 {
		row = t.params[0]
	}
	if len(t.params) > 1 && t.params[1] > 0 {
		col = t.params[1]
	}

	t.y = max(0, min(t.rows-1, row-1))
	t.x = max(0, min(t.cols-1, col-1))
}

func (t *Terminal) cursorMove(cmd byte) {
	n := 1
	if len(t.params) > 0 && t.params[0] > 0 {
		n = t.params[0]
	}

	switch cmd {
	case 'A': // Up
		t.y = max(0, t.y-n)
	case 'B': // Down
		t.y = min(t.rows-1, t.y+n)
	case 'C': // Forward
		t.x = min(t.cols-1, t.x+n)
	case 'D': // Back
		t.x = max(0, t.x-n)
	}
}

func (t *Terminal) eraseDisplay() {
	mode := eraseToEnd
	if len(t.params) > 0 {
		mode = t.params[0]
	}

	blank := Cell{Char: ' ', FG: defaultFGColor, BG: t.curBG}

	switch mode {
	case eraseToEnd:
		t.eraseDisplayToEnd(blank)
	case eraseToStart:
		t.eraseDisplayToStart(blank)
	case eraseAll:
		t.Clear()
	}
	t.dirty = true
}

func (t *Terminal) eraseDisplayToEnd(blank Cell) {
	for x := t.x; x < t.cols; x++ {
		t.grid[t.y][x] = blank
	}
	for y := t.y + 1; y < t.rows; y++ {
		for x := range t.cols {
			t.grid[y][x] = blank
		}
	}
}

func (t *Terminal) eraseDisplayToStart(blank Cell) {
	for y := range t.y {
		for x := range t.cols {
			t.grid[y][x] = blank
		}
	}
	for x := 0; x <= t.x && x < t.cols; x++ {
		t.grid[t.y][x] = blank
	}
}

func (t *Terminal) eraseLine() {
	mode := eraseToEnd
	if len(t.params) > 0 {
		mode = t.params[0]
	}

	blank := Cell{Char: ' ', FG: defaultFGColor, BG: t.curBG}

	switch mode {
	case eraseToEnd:
		for x := t.x; x < t.cols; x++ {
			t.grid[t.y][x] = blank
		}
	case eraseToStart:
		for x := 0; x <= t.x && x < t.cols; x++ {
			t.grid[t.y][x] = blank
		}
	case eraseAll:
		for x := range t.cols {
			t.grid[t.y][x] = blank
		}
	}
	t.dirty = true
}

func (t *Terminal) sgr() {
	if len(t.params) == 0 {
		t.resetAttrs()

		return
	}

	for _, param := range t.params {
		t.applySGRParam(param)
	}
}

func (t *Terminal) resetAttrs() {
	t.bold = false
	t.blink = false
	t.curFG = defaultFGColor
	t.curBG = defaultBGColor
}

func (t *Terminal) applySGRParam(param int) {
	if t.applySGRStyle(param) {
		return
	}
	t.applySGRColor(param)
}

func (t *Terminal) applySGRStyle(param int) bool {
	switch param {
	case sgrReset:
		t.resetAttrs()
	case sgrBold:
		t.bold = true
	case sgrBlink:
		t.blink = true
	case sgrBoldOff:
		t.bold = false
	case sgrBlinkOff:
		t.blink = false
	default:
		return false
	}

	return true
}

func (t *Terminal) applySGRColor(param int) {
	if t.applySGRFG(param) {
		return
	}
	t.applySGRBG(param)
}

func (t *Terminal) applySGRFG(param int) bool {
	switch {
	case param >= sgrFGStart && param <= sgrFGEnd:
		t.curFG = byte(param - sgrFGStart)
	case param == sgrFGDefault:
		t.curFG = defaultFGColor
	case param >= sgrHiFGStart && param <= sgrHiFGEnd:
		t.curFG = byte(param - sgrHiFGStart + brightColorOffset)
	default:
		return false
	}

	return true
}

func (t *Terminal) applySGRBG(param int) {
	switch {
	case param >= sgrBGStart && param <= sgrBGEnd:
		t.curBG = byte(param - sgrBGStart)
	case param == sgrBGDefault:
		t.curBG = defaultBGColor
	case param >= sgrHiBGStart && param <= sgrHiBGEnd:
		t.curBG = byte(param - sgrHiBGStart + brightColorOffset)
	}
}

func (t *Terminal) effectiveFG() byte {
	if t.bold && t.curFG < brightColorOffset {
		return t.curFG + brightColorOffset
	}

	return t.curFG
}

func (t *Terminal) effectiveBG() byte {
	if t.blink && t.blinkMode == BlinkICE && t.curBG < brightColorOffset {
		return t.curBG + brightColorOffset
	}

	return t.curBG
}

func (t *Terminal) writeChar(b byte) {
	cell := Cell{
		Char: b,
		FG:   t.effectiveFG(),
		BG:   t.effectiveBG(),
	}

	t.grid[t.y][t.x] = cell
	t.dirty = true
	t.x++

	if t.x >= t.cols {
		if t.lineWrap {
			t.x = 0
			t.lineFeed()
		} else {
			t.x = t.cols - 1
		}
	}
}

func (t *Terminal) lineFeed() {
	t.y++
	if t.y >= t.rows {
		t.scrollUp()
		t.y = t.rows - 1
	}
}

func (t *Terminal) scrollUp() {
	blank := Cell{Char: ' ', FG: defaultFGColor, BG: defaultBGColor}
	for y := range t.rows - 1 {
		copy(t.grid[y], t.grid[y+1])
	}
	for x := range t.cols {
		t.grid[t.rows-1][x] = blank
	}
	t.dirty = true
}

func (t *Terminal) tab() {
	t.x = (t.x + tabStop) & ^(tabStop - 1)
	if t.x >= t.cols {
		t.x = t.cols - 1
	}
}
