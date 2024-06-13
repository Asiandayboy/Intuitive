package fileeditor

import (
	"github.com/Asiandayboy/CLITextEditor/util/ansi"
	"github.com/Asiandayboy/CLITextEditor/util/math"
)

type actionFunc func()

var savedCursorX int = 0
var savedCursorXFlag bool = false

func setSavedCursorX(cursorX int, saveFlag bool) {
	savedCursorXFlag = saveFlag
	savedCursorX = cursorX
}

func constrainCursorX(f *FileEditor) {
	/*
		Constrain cursor when moving cursor up and down
		with keys, keeping the acX clamped from its initial position
		on the first up or down
	*/
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

func (f *FileEditor) SetCursorPositionOnClick(m MouseInput) byte {
	/*
		constrain cursor horizontally and vertically to not extend
		past visual buffer
	*/
	y := len(f.VisualBuffer)

	if m.Y > y {
		currLineLen := len(f.VisualBuffer[y-1])
		x := currLineLen + EditorLeftMargin

		f.apparentCursorX = x
		f.apparentCursorY = y
		setSavedCursorX(x, false)
		ansi.MoveCursor(y, x)
		return CursorPositionChange
	}

	currLineLen := len(f.VisualBuffer[m.Y-1])
	x := math.Clamp(m.X, EditorLeftMargin, currLineLen+EditorLeftMargin)

	f.apparentCursorX = x
	f.apparentCursorY = m.Y
	setSavedCursorX(x, false)
	ansi.MoveCursor(m.Y, x)

	return CursorPositionChange
}

func (f *FileEditor) actionCursorLeft() {
	if f.apparentCursorX > EditorLeftMargin {
		f.apparentCursorX--
	} else {
		if f.apparentCursorY > 1 { // move to end of previous line
			f.DecrementCursorY()
			line := f.VisualBuffer[f.apparentCursorY-1]
			f.apparentCursorX = len(line) + EditorLeftMargin
		}
	}

	setSavedCursorX(f.apparentCursorX, false)
}

func (f *FileEditor) actionCursorRight() {
	line := f.VisualBuffer[f.apparentCursorY-1]
	if f.apparentCursorX <= len(line)+EditorLeftMargin-1 {
		f.apparentCursorX++
	} else {
		if f.apparentCursorY < len(f.VisualBuffer) { // move to start of next line
			f.IncrementCursorY()
			f.apparentCursorX = EditorLeftMargin
		}
	}

	setSavedCursorX(f.apparentCursorX, false)
}

func (f *FileEditor) actionCursorUp() {
	if f.apparentCursorY > 1 {
		f.DecrementCursorY()
		constrainCursorX(f)
	}
}

func (f *FileEditor) actionCursorDown() {
	if f.apparentCursorY < len(f.VisualBuffer) {
		f.IncrementCursorY()
		constrainCursorX(f)
	}

}

func (f *FileEditor) actionScrollDown() {

}

func (f *FileEditor) actionScrollUp() {

}

/*
Adds a new line by mutating the FileBuffer
*/
func (f *FileEditor) actionNewLine() byte {
	line := f.FileBuffer[f.bufferLine]
	n := len(f.FileBuffer)

	// split the current line
	beforeSplit := line[:f.bufferIndex]
	afterSplit := line[f.bufferIndex:]

	// insert the new line (afterSplit) in the middle of buffer array
	result := make([]string, n+1)

	copy(result, f.FileBuffer[:f.bufferLine])

	result[f.bufferLine] = beforeSplit
	result[f.bufferLine+1] = afterSplit

	copy(result[f.bufferLine+2:], f.FileBuffer[f.bufferLine+1:])

	f.FileBuffer = result

	// update cursor position
	if f.bufferIndex == len(line) { // inserting new line at the end of a line
		f.RefreshVisualBuffers()
		f.actionCursorDown()
		setSavedCursorX(f.apparentCursorX, false)
		return NewLineInsertedAtLineEnd
	}

	f.IncrementCursorY()
	f.apparentCursorX = EditorLeftMargin
	setSavedCursorX(f.apparentCursorX, false)

	return NewLineInserted
}

func (f *FileEditor) actionTyping(key byte) {
	line := f.FileBuffer[f.bufferLine]

	before := line[:f.bufferIndex]
	after := line[f.bufferIndex:]

	f.FileBuffer[f.bufferLine] = before + string(key) + after
	f.apparentCursorX++
	if f.apparentCursorX > f.TermWidth {
		f.apparentCursorX = EditorLeftMargin + 1
		f.IncrementCursorY()
	}

}

func (f *FileEditor) actionDeleteText() {
	if f.bufferLine <= 0 && f.bufferIndex <= 0 {
		return
	}

	setSavedCursorX(f.apparentCursorX, false)

	line := f.FileBuffer[f.bufferLine]

	if len(line) <= 0 { // deleting an empty line
		f.DecrementCursorY()
		f.apparentCursorX = len(f.VisualBuffer[f.apparentCursorY-1]) + EditorLeftMargin
		f.FileBuffer = append(f.FileBuffer[:f.bufferLine], f.FileBuffer[f.bufferLine+1:]...)
		return
	}

	if f.bufferIndex <= 0 { // deleting at beginning of a line
		currLine := f.FileBuffer[f.bufferLine]

		/*
			append curr line to prev line, and remove the curr line
		*/
		f.FileBuffer[f.bufferLine-1] += currLine
		f.FileBuffer = append(f.FileBuffer[:f.bufferLine], f.FileBuffer[f.bufferLine+1:]...)
		f.DecrementCursorY()
		f.apparentCursorX = len(f.VisualBuffer[f.apparentCursorY-1]) + EditorLeftMargin

	} else { // deleting anywhere else
		before := line[:f.bufferIndex-1]
		after := line[f.bufferIndex:]

		f.FileBuffer[f.bufferLine] = before + after
		f.apparentCursorX--

		/*
			f.bufferIndex > 1 to avoid automatically moving to prev line when deleting from index <= 1;
			we want the conditional above this to handle that
		*/
		if f.bufferIndex > 1 && f.apparentCursorX <= EditorLeftMargin {
			f.DecrementCursorY()
			f.apparentCursorX = len(f.VisualBuffer[f.apparentCursorY-1]) + EditorLeftMargin
		}
	}
}
