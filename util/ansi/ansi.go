package ansi

import (
	"fmt"
	"os"
)

/*
Returns true if the given key is an alpha character
according to the ASCII table
*/
func IsAlphaChar(key byte) bool {
	return key >= 32 && key < 127
}

func MoveCursor(row, col int) {
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}

	fmt.Printf("\x1b[%d;%dH", row, col)
}

func MoveCursorRight(n int) {
	fmt.Printf("\x1b[%dC", n)
}

func MoveCursorLeft(n int) {
	fmt.Printf("\x1b[%dD", n)
}

func MoveCursorDown(n int) {
	fmt.Printf("\x1b[%dB", n)
}

func ClearEntireScreen() {
	fmt.Print("\x1b[2J")
}

func EnableMouseReporting() {
	fmt.Print("\x1b[?1003h")
	fmt.Print("\x1b[?1006h")
}

func DisableMouseReporting() {
	fmt.Print("\x1b[?1003l")
	fmt.Print("\x1b[?1006l")
}

func HideCursor() {
	fmt.Print("\x1b[?25l")
}

func ShowCursor() {
	fmt.Print("\x1b[?25h")
}

func EnableBlinkingLineCursor() {
	fmt.Print("\033[5 q")
}

func EraseEntireLine() {
	fmt.Print("\x1b[2K")
}

func EraseLineFromCursorToEnd() {
	fmt.Print("\x1b[0K")
}

func EraseLineFromCursorToStart() {
	fmt.Print("\x1b[1K")
}

func EraseFromCursorToEndOfLine() {
	fmt.Print("\x1b[0K")
}

func EnableAlternateScreenBuffer() {
	fmt.Print("\x1b[?1049h")
}

func DisableAlternateScreenBuffer() {
	fmt.Print("\x1b[?1049l")
}

func GetArrowKeyPress(buffer []byte) string {
	if !(buffer[0] == '\x1b' && buffer[1] == '[') {
		return ""
	}

	key := buffer[2]

	switch key {
	case 'A':
		return "UP"
	case 'B':
		return "DOWN"
	case 'C':
		return "RIGHT"
	case 'D':
		return "LEFT"
	}

	return ""
}

func SetTerminalWindowTitle(title string) {
	fmt.Print("\x1b]2;" + title + "\x07")
}

/*
Reads stdin up to 3 bytes and returns the key that was
pressed as a byte, as well as the buffer for optional use

# Returns an error if input could not be read from stdin

If the isArrowKey flag is true, key will either be 'A', 'B', 'C'
or 'D', which is up, down, right, and left, respectively
*/
func ListenForKeyPress() (key byte, isArrowKey bool, buffer []byte, err error) {
	buffer = make([]byte, 4)
	_, err = os.Stdin.Read(buffer)
	if err != nil {
		return 0, false, buffer, err
	}

	if buffer[0] == 0x1b && buffer[1] == '[' {
		return buffer[2], true, buffer, nil
	}

	return buffer[0], false, buffer, nil

}
