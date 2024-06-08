package fileeditor

import (
	"fmt"

	"github.com/Asiandayboy/CLITextEditor/util/ansi"
	"github.com/Asiandayboy/CLITextEditor/util/math"
)

type actionFunc func(key byte)

var savedCursorX int = 0
var savedCursorXFlag bool = false

func setSavedCursorX(cursorX int, saveFlag bool) {
	savedCursorXFlag = saveFlag
	savedCursorX = cursorX
}

func verticallyConstrainCursor(f *FileEditor) {
	line := f.VisualBuffer[f.apparentCursorY-1]

	if !savedCursorXFlag {
		setSavedCursorX(f.apparentCursorX, true)
	}

	if len(line)+EditorLeftMargin < savedCursorX {
		f.apparentCursorX = len(line) + EditorLeftMargin
	} else {
		f.apparentCursorX = math.Clamp(
			f.apparentCursorX,
			savedCursorX,
			len(line)+EditorLeftMargin,
		)
	}
}

func (f *FileEditor) actionDeleteText(key byte) {
	fmt.Println("delete text")
}

func (f *FileEditor) actionCursorLeft(key byte) {
	if f.apparentCursorX > EditorLeftMargin {
		f.apparentCursorX--
	} else {
		if f.apparentCursorY > 1 { // move to end of previous line
			f.apparentCursorY--
			line := f.VisualBuffer[f.apparentCursorY-1]
			f.apparentCursorX = len(line) + EditorLeftMargin
		}
	}

	setSavedCursorX(f.apparentCursorX, false)
}

func (f *FileEditor) actionCursorRight(key byte) {
	line := f.VisualBuffer[f.apparentCursorY-1]
	if f.apparentCursorX <= len(line)+EditorLeftMargin-1 {
		f.apparentCursorX++
	} else {
		if f.apparentCursorY < len(f.VisualBuffer) { // move to start of next line
			f.apparentCursorY++
			f.apparentCursorX = EditorLeftMargin
		}
	}

	setSavedCursorX(f.apparentCursorX, false)
}

func (f *FileEditor) actionCursorUp(key byte) {
	if f.apparentCursorY > 1 {
		f.apparentCursorY--
		verticallyConstrainCursor(f)
	}
}

func (f *FileEditor) actionCursorDown(key byte) {
	if f.apparentCursorY < len(f.VisualBuffer) {
		f.apparentCursorY++
		verticallyConstrainCursor(f)
	}

}

func (f *FileEditor) actionNewLine(key byte) {
	fmt.Println("new line")
}

func (f *FileEditor) actionTyping(key byte) {
	if ansi.IsAlphaChar(key) {
		fmt.Println(string(key))
	} else { // typing a char that is not printable
		fmt.Println("typing")
	}
}
