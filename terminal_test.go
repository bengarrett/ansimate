package ansimate_test

import (
	"testing"

	"github.com/bengarrett/ansimate"
	"github.com/nalgeon/be"
)

func TestNewTerminal(t *testing.T) {
	t.Parallel()
	term := ansimate.NewTerminal(80, 25, ansimate.BlinkOff)
	be.Equal(t, len(term.Grid()), 25)
	be.Equal(t, len(term.Grid()[0]), 80)
	be.True(t, term.IsDirty())

	term.ResetDirty()
	be.Equal(t, term.IsDirty(), false)

	// Defaults for zero or negative values
	termDef := ansimate.NewTerminal(0, -1, ansimate.BlinkOff)
	be.Equal(t, len(termDef.Grid()), 25)
	be.Equal(t, len(termDef.Grid()[0]), 80)
}

func TestTerminalPrintable(t *testing.T) {
	t.Parallel()
	term := ansimate.NewTerminal(10, 5, ansimate.BlinkOff)
	_, err := term.Write([]byte("Hello"))
	be.Err(t, err, nil)

	grid := term.Grid()
	be.Equal(t, grid[0][0].Char, byte('H'))
	be.Equal(t, grid[0][1].Char, byte('e'))
	be.Equal(t, grid[0][2].Char, byte('l'))
	be.Equal(t, grid[0][3].Char, byte('l'))
	be.Equal(t, grid[0][4].Char, byte('o'))
	be.Equal(t, grid[0][5].Char, byte(' '))
	be.True(t, term.IsDirty())
}

func TestTerminalControlCodes(t *testing.T) {
	t.Parallel()
	term := ansimate.NewTerminal(10, 5, ansimate.BlinkOff)

	// Backspace
	_, _ = term.Write([]byte("AB\bC"))
	grid := term.Grid()
	be.Equal(t, grid[0][0].Char, byte('A'))
	be.Equal(t, grid[0][1].Char, byte('C'))

	// Carriage return and line feed
	_, _ = term.Write([]byte("\r\nDE"))
	be.Equal(t, grid[1][0].Char, byte('D'))
	be.Equal(t, grid[1][1].Char, byte('E'))

	// Tab
	term.Clear()
	_, _ = term.Write([]byte("A\tB"))
	grid = term.Grid()
	be.Equal(t, grid[0][0].Char, byte('A'))
	be.Equal(t, grid[0][8].Char, byte('B'))

	// WriteByte
	err := term.WriteByte('!')
	be.Err(t, err, nil)
	be.Equal(t, grid[0][9].Char, byte('!'))
}

func TestTerminalCursorPosition(t *testing.T) {
	t.Parallel()
	term := ansimate.NewTerminal(80, 25, ansimate.BlinkOff)

	// Position row 3, col 5 (1-indexed -> y=2, x=4)
	_, _ = term.Write([]byte("\x1b[3;5HX"))
	grid := term.Grid()
	be.Equal(t, grid[2][4].Char, byte('X'))

	// Default position (1, 1) -> y=0, x=0
	_, _ = term.Write([]byte("\x1b[HY"))
	be.Equal(t, grid[0][0].Char, byte('Y'))

	// Alternative 'f' command
	_, _ = term.Write([]byte("\x1b[2;2fZ"))
	be.Equal(t, grid[1][1].Char, byte('Z'))

	// Out of bounds clamping (row clamps to 25 -> y=24)
	_, _ = term.Write([]byte("\x1b[999;50H#"))
	be.Equal(t, grid[24][49].Char, byte('#'))
}

func TestTerminalCursorMoves(t *testing.T) {
	t.Parallel()
	term := ansimate.NewTerminal(80, 25, ansimate.BlinkOff)

	// Start at row 10, col 10 (y=9, x=9)
	_, _ = term.Write([]byte("\x1b[10;10H"))

	// Up 2 rows -> y=7
	_, _ = term.Write([]byte("\x1b[2A"))
	// Back 3 cols -> x=6
	_, _ = term.Write([]byte("\x1b[3D"))
	_, _ = term.Write([]byte("A"))
	grid := term.Grid()
	be.Equal(t, grid[7][6].Char, byte('A'))

	// Down 4 rows -> y=11
	_, _ = term.Write([]byte("\x1b[4B"))
	// Forward 5 cols -> x=12
	_, _ = term.Write([]byte("\x1b[5C"))
	_, _ = term.Write([]byte("B"))
	be.Equal(t, grid[11][12].Char, byte('B'))
}

func TestTerminalSaveRestoreCursor(t *testing.T) {
	t.Parallel()
	term := ansimate.NewTerminal(80, 25, ansimate.BlinkOff)

	// Position at 5, 10 and save
	_, _ = term.Write([]byte("\x1b[5;10H\x1b[s"))
	// Move away and type
	_, _ = term.Write([]byte("\x1b[1;1HA"))
	// Restore and type
	_, _ = term.Write([]byte("\x1b[uB"))

	grid := term.Grid()
	be.Equal(t, grid[0][0].Char, byte('A'))
	be.Equal(t, grid[4][9].Char, byte('B'))
}

func TestTerminalErase(t *testing.T) {
	t.Parallel()
	term := ansimate.NewTerminal(10, 5, ansimate.BlinkOff)

	// Fill line 0
	_, _ = term.Write([]byte("0123456789"))
	// Move to col 6 (1-indexed) and erase to end of line
	_, _ = term.Write([]byte("\x1b[1;6H\x1b[K"))
	grid := term.Grid()
	be.Equal(t, grid[0][4].Char, byte('4'))
	be.Equal(t, grid[0][5].Char, byte(' '))
	be.Equal(t, grid[0][9].Char, byte(' '))

	// Erase from start of line to cursor
	_, _ = term.Write([]byte("\x1b[1;4H\x1b[1K"))
	be.Equal(t, grid[0][0].Char, byte(' '))
	be.Equal(t, grid[0][3].Char, byte(' '))
	be.Equal(t, grid[0][4].Char, byte('4'))

	// Erase entire line
	_, _ = term.Write([]byte("\x1b[2K"))
	be.Equal(t, grid[0][4].Char, byte(' '))

	// Erase in display modes 0, 1, 2
	term.Clear()
	_, _ = term.Write([]byte("Line0\r\nLine1\r\nLine2"))
	_, _ = term.Write([]byte("\x1b[2;3H\x1b[0J"))
	be.Equal(t, grid[0][0].Char, byte('L'))
	be.Equal(t, grid[1][0].Char, byte('L'))
	be.Equal(t, grid[1][3].Char, byte(' '))
	be.Equal(t, grid[2][0].Char, byte(' '))

	_, _ = term.Write([]byte("\x1b[2J"))
	be.Equal(t, grid[0][0].Char, byte(' '))
}

func TestTerminalSGR(t *testing.T) {
	t.Parallel()
	term := ansimate.NewTerminal(80, 25, ansimate.BlinkOff)

	// Red foreground (31), Green background (42)
	_, _ = term.Write([]byte("\x1b[31;42mX"))
	grid := term.Grid()
	be.Equal(t, grid[0][0].Char, byte('X'))
	be.Equal(t, grid[0][0].FG, byte(1)) // Red is 1 in ANSI
	be.Equal(t, grid[0][0].BG, byte(2)) // Green is 2 in ANSI

	// Bold adds bright offset (8) to foreground
	_, _ = term.Write([]byte("\x1b[1mY"))
	be.Equal(t, grid[0][1].Char, byte('Y'))
	be.Equal(t, grid[0][1].FG, byte(9)) // 1 + 8 = 9 (Bright Red)

	// High intensity colors
	_, _ = term.Write([]byte("\x1b[0;94;105mH"))
	be.Equal(t, grid[0][2].Char, byte('H'))
	be.Equal(t, grid[0][2].FG, byte(12)) // 94 - 90 + 8 = 12 (Bright Blue)
	be.Equal(t, grid[0][2].BG, byte(13)) // 105 - 100 + 8 = 13 (Bright Magenta)

	// Reset
	_, _ = term.Write([]byte("\x1b[0mZ"))
	be.Equal(t, grid[0][3].Char, byte('Z'))
	be.Equal(t, grid[0][3].FG, byte(7)) // default FG (7)
	be.Equal(t, grid[0][3].BG, byte(0)) // default BG (0)
}

func TestTerminalICEBlink(t *testing.T) {
	t.Parallel()
	termICE := ansimate.NewTerminal(80, 25, ansimate.BlinkICE)

	// Blue background (44) + blink (5) -> in iCE mode promotes to bright background (4 + 8 = 12)
	_, _ = termICE.Write([]byte("\x1b[44;5mI"))
	grid := termICE.Grid()
	be.Equal(t, grid[0][0].Char, byte('I'))
	be.Equal(t, grid[0][0].BG, byte(12))

	// In BlinkOff mode, background remains standard (4)
	termOff := ansimate.NewTerminal(80, 25, ansimate.BlinkOff)
	_, _ = termOff.Write([]byte("\x1b[44;5mO"))
	gridOff := termOff.Grid()
	be.Equal(t, gridOff[0][0].Char, byte('O'))
	be.Equal(t, gridOff[0][0].BG, byte(4))
}

func TestTerminalScrollingAndWrapping(t *testing.T) {
	t.Parallel()
	term := ansimate.NewTerminal(5, 3, ansimate.BlinkOff)

	// Fill line 0 and wrap to line 1
	_, _ = term.Write([]byte("12345AB"))
	grid := term.Grid()
	be.Equal(t, grid[0][0].Char, byte('1'))
	be.Equal(t, grid[0][4].Char, byte('5'))
	be.Equal(t, grid[1][0].Char, byte('A'))
	be.Equal(t, grid[1][1].Char, byte('B'))

	// Cause scrolling by writing more lines than terminal rows
	_, _ = term.Write([]byte("\r\nRow2\r\nRow3\r\nRow4"))
	be.Equal(t, grid[2][0].Char, byte('R'))
	be.Equal(t, grid[2][3].Char, byte('4'))

	// Disable line wrap: \x1b[?7l
	termWrap := ansimate.NewTerminal(5, 3, ansimate.BlinkOff)
	_, _ = termWrap.Write([]byte("\x1b[?7l12345678"))
	gridWrap := termWrap.Grid()
	// Line 0 col 4 should hold the last character written ('8')
	be.Equal(t, gridWrap[0][4].Char, byte('8'))
	// Line 1 should be untouched
	be.Equal(t, gridWrap[1][0].Char, byte(' '))
}
