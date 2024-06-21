package fileeditor

import (
	"github.com/Asiandayboy/CLITextEditor/util/ansi"
	"github.com/Asiandayboy/CLITextEditor/util/math"
)

type actionFunc func()

const (
	upDirection   uint8 = 1
	downDirection uint8 = 2
)

var savedACX int = 0
var savedCursorXFlag bool = false
var savedViewportXOffset int = 0

func setSavedCursorX(acX int, xOffset int, saveFlag bool) {
	savedCursorXFlag = saveFlag
	savedACX = acX
	savedViewportXOffset = xOffset
}

func constrainCursorX(f *FileEditor, direction uint8) {
	/*
		Constrain cursor when moving cursor up and down
		with keys, keeping the acX clamped from its initial position
		on the first up or down
	*/
	currIdx := f.apparentCursorY - 1 + f.ViewportOffsetY
	currLine := f.VisualBuffer[currIdx]

	if !savedCursorXFlag {
		setSavedCursorX(f.apparentCursorX, f.ViewportOffsetX, true)
	}

	lineLength := len(currLine) + EditorLeftMargin

	if lineLength < savedACX+savedViewportXOffset {
		if len(currLine) <= f.GetViewportWidth() {
			/*
				Bring offset back to 0 if the entire line can fit on the viewport
				without horizontal scrolling
			*/
			f.ViewportOffsetX = 0
		} else if !f.SoftWrap {
			if direction == upDirection {
				prevLine := f.VisualBuffer[currIdx+1]
				if len(currLine) > len(prevLine) {
					f.ViewportOffsetX = lineLength - f.TermWidth
				}
			} else if direction == downDirection && currIdx > 0 {
				prevLine := f.VisualBuffer[currIdx-1]
				if len(currLine) > len(prevLine) {
					f.ViewportOffsetX = lineLength - f.TermWidth
				}
			}
		}
		f.apparentCursorX = lineLength - f.ViewportOffsetX
	} else {
		f.ViewportOffsetX = savedViewportXOffset
		f.apparentCursorX = math.Clamp(
			f.apparentCursorX,
			savedACX,
			len(currLine)+EditorLeftMargin,
		)
	}
}

func (f *FileEditor) SetCursorPositionOnClick(m MouseInput) byte {
	/*
		constrain cursor horizontally and vertically to not extend
		past visual buffer
	*/
	y := len(f.VisualBuffer)

	/*
		checking if absolute cursor Y pos exceeds the range of the visual buffer
		so that we can clamp the cursor position
		- absolute cursor Y pos = apparentCursorY + ViewportOffsetY
	*/
	if m.Y+f.ViewportOffsetY > y {
		line := f.VisualBuffer[y-1]
		currLineLen := len(line)

		if !f.SoftWrap && currLineLen > f.GetViewportWidth() {
			f.ViewportOffsetX = currLineLen - f.GetViewportWidth() + 1 // maybe + 2 if we want extra space?
		}
		x := currLineLen + EditorLeftMargin - f.ViewportOffsetX

		f.apparentCursorX = x
		f.apparentCursorY = y - f.ViewportOffsetY
		setSavedCursorX(x, f.ViewportOffsetX, false)
		ansi.MoveCursor(f.apparentCursorY, f.apparentCursorX)
		return CursorPositionChange
	}

	line := f.VisualBuffer[m.Y-1+f.ViewportOffsetY]
	currLineLen := len(line)

	isLineInViewport := len(line[math.Min(f.ViewportOffsetX, currLineLen):]) > 0
	if !f.SoftWrap && !isLineInViewport {
		f.ViewportOffsetX = math.Max(currLineLen, 0)
	}
	x := math.Clamp(m.X, EditorLeftMargin, currLineLen+EditorLeftMargin-f.ViewportOffsetX)

	f.apparentCursorX = x
	f.apparentCursorY = m.Y
	setSavedCursorX(x, f.ViewportOffsetX, false)
	ansi.MoveCursor(m.Y, x)

	return CursorPositionChange
}

func (f *FileEditor) actionCursorLeft() {
	if f.apparentCursorX+f.ViewportOffsetX > EditorLeftMargin {
		if f.apparentCursorX == EditorLeftMargin && f.ViewportOffsetX > 0 {
			f.ViewportOffsetX--
		} else {
			f.apparentCursorX--
		}
	} else {
		if f.apparentCursorY > 1 { // move to end of previous line
			f.DecrementCursorY()
			line := f.VisualBuffer[f.apparentCursorY-1+f.ViewportOffsetY]
			f.apparentCursorX = len(line) + EditorLeftMargin
			if len(line) >= f.GetViewportWidth() {
				f.ViewportOffsetX = len(line) - f.GetViewportWidth() + 1
				f.apparentCursorX = f.TermWidth
			}
		} else if f.apparentCursorY == 1 && f.ViewportOffsetY > 0 {
			f.actionScrollUp()
			line := f.VisualBuffer[f.apparentCursorY-1+f.ViewportOffsetY]
			f.apparentCursorX = len(line) + EditorLeftMargin
			if len(line) > f.GetViewportWidth() {
				f.ViewportOffsetX = len(line) - f.GetViewportWidth() + 1
				f.apparentCursorX = f.TermWidth
			}
		}
	}

	setSavedCursorX(f.apparentCursorX, f.ViewportOffsetX, false)
}

func (f *FileEditor) actionCursorRight() {
	line := f.VisualBuffer[f.apparentCursorY-1+f.ViewportOffsetY]

	if f.apparentCursorX+f.ViewportOffsetX <= len(line)+EditorLeftMargin-1 {
		if f.apparentCursorX == f.TermWidth {
			f.ViewportOffsetX++
		} else {
			f.apparentCursorX++
		}
	} else {
		if f.apparentCursorY+f.ViewportOffsetY < len(f.VisualBuffer) { // move to start of next line
			f.ViewportOffsetX = 0
			f.IncrementCursorY()
			f.apparentCursorX = EditorLeftMargin

			if f.apparentCursorY == f.GetViewportHeight() {
				f.actionScrollDown()
			}
		}
	}

	setSavedCursorX(f.apparentCursorX, f.ViewportOffsetX, false)
}

func (f *FileEditor) actionCursorUp() {
	if f.apparentCursorY > 1 {
		f.DecrementCursorY()
		constrainCursorX(f, upDirection)
	} else if f.apparentCursorY == 1 && f.ViewportOffsetY > 0 {
		f.actionScrollUp()
		constrainCursorX(f, upDirection)
	}
}

func (f *FileEditor) actionCursorDown() {
	if f.apparentCursorY+f.ViewportOffsetY < len(f.VisualBuffer) {
		if f.apparentCursorY == f.GetViewportHeight() {
			f.actionScrollDown()
		}
		f.IncrementCursorY()
		constrainCursorX(f, downDirection)
	}

}

func (f *FileEditor) actionScrollDown() {
	if f.apparentCursorY+f.ViewportOffsetY < len(f.VisualBuffer) {
		f.ViewportOffsetY++
	}
}

func (f *FileEditor) actionScrollUp() {
	if f.ViewportOffsetY > 0 {
		f.ViewportOffsetY--
	}

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
		if f.SoftWrap {
			f.RefreshSoftWrapVisualBuffers()
		} else {
			f.RefreshNoWrapVisualBuffers()
		}
		f.actionCursorDown()
		setSavedCursorX(f.apparentCursorX, f.ViewportOffsetX, false)
		return NewLineInsertedAtLineEnd
	}

	f.IncrementCursorY()
	f.ViewportOffsetX = 0
	f.apparentCursorX = EditorLeftMargin
	setSavedCursorX(f.apparentCursorX, f.ViewportOffsetX, false)

	return NewLineInserted
}

func (f *FileEditor) actionTyping(key byte) {
	line := f.FileBuffer[f.bufferLine]

	before := line[:f.bufferIndex]
	after := line[f.bufferIndex:]

	f.FileBuffer[f.bufferLine] = before + string(key) + after
	f.apparentCursorX++

	if f.apparentCursorX > f.TermWidth {
		if f.SoftWrap {
			f.apparentCursorX = EditorLeftMargin + 1
			f.IncrementCursorY()

			if f.apparentCursorY == f.GetViewportHeight() {
				f.ViewportOffsetY++ // changing it directly bc the condition in actionScrollDown doesn't apply here, ig
			}
		} else {
			f.apparentCursorX = f.TermWidth
			f.ViewportOffsetX++
		}
	}

}

func (f *FileEditor) actionDeleteText() {
	if f.bufferLine <= 0 && f.bufferIndex <= 0 {
		return
	}

	setSavedCursorX(f.apparentCursorX, f.ViewportOffsetX, false)

	line := f.FileBuffer[f.bufferLine]

	if len(line) <= 0 { // deleting an empty line
		if f.apparentCursorY == 1 && f.ViewportOffsetY > 0 {
			f.actionScrollUp()
		}
		f.DecrementCursorY()
		prevLine := f.VisualBuffer[f.apparentCursorY-1+f.ViewportOffsetY]
		if f.SoftWrap {
			f.apparentCursorX = len(prevLine) + EditorLeftMargin
		} else {
			f.apparentCursorX = math.Clamp(
				len(prevLine)+EditorLeftMargin,
				EditorLeftMargin,
				f.TermWidth,
			)
			f.ViewportOffsetX = math.Clamp(
				len(prevLine)+EditorLeftMargin-f.TermWidth,
				0,
				len(prevLine)+EditorLeftMargin-f.TermWidth,
			)
		}
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
		if f.apparentCursorY == 1 && f.ViewportOffsetY > 0 {
			f.actionScrollUp()
		}
		f.DecrementCursorY()
		prevLine := f.VisualBuffer[f.apparentCursorY-1+f.ViewportOffsetY]
		if f.SoftWrap {
			f.apparentCursorX = len(prevLine) + EditorLeftMargin
		} else {
			if len(prevLine) >= f.GetViewportWidth()-1 {
				// some margin space added to the offset so the user can see where the lines joined
				const extraSpace int = 1
				f.apparentCursorX = math.Clamp(
					len(prevLine)+EditorLeftMargin-extraSpace,
					EditorLeftMargin,
					f.TermWidth-extraSpace,
				)
				f.ViewportOffsetX = math.Clamp(
					len(prevLine)+EditorLeftMargin-f.TermWidth+extraSpace,
					0,
					len(prevLine)+EditorLeftMargin-f.TermWidth+extraSpace,
				)
			} else {
				f.apparentCursorX = math.Clamp(
					len(prevLine)+EditorLeftMargin,
					EditorLeftMargin,
					f.TermWidth,
				)
				f.ViewportOffsetX = math.Clamp(
					len(prevLine)+EditorLeftMargin-f.TermWidth,
					0,
					len(prevLine)+EditorLeftMargin-f.TermWidth,
				)
			}
		}

	} else { // deleting anywhere else
		before := line[:f.bufferIndex-1]
		after := line[f.bufferIndex:]

		f.FileBuffer[f.bufferLine] = before + after

		if f.ViewportOffsetX == 0 {
			f.apparentCursorX--
		} else {
			f.ViewportOffsetX = math.Max(f.ViewportOffsetX-1, 0)
		}

		/*
			f.bufferIndex > 1 to avoid automatically moving to prev line when deleting from index <= 1;
			we want the conditional above this to handle that
		*/
		if f.bufferIndex > 1 && f.apparentCursorX <= EditorLeftMargin {
			f.DecrementCursorY()
			if f.apparentCursorY == 1 && f.ViewportOffsetY > 0 {
				f.actionScrollUp()
			}
			f.apparentCursorX = len(f.VisualBuffer[f.apparentCursorY-1+f.ViewportOffsetY]) + EditorLeftMargin
		}
	}
}

func (f *FileEditor) ToggleSoftWrap(softWrapEnabled bool) byte {
	if softWrapEnabled {
		if f.SoftWrap {
			return 0
		}
		f.SoftWrap = true
		lenBefore := len(f.VisualBuffer)
		f.RefreshSoftWrapVisualBuffers()
		lenAfter := len(f.VisualBuffer)

		if f.ViewportOffsetY > 0 {
			f.ViewportOffsetY += lenAfter - lenBefore
		}
		f.ViewportOffsetX = 0

		return SoftWrapEnabled
	} else {
		if !f.SoftWrap {
			return 0
		}
		f.SoftWrap = false
		lenBefore := len(f.VisualBuffer)
		f.RefreshNoWrapVisualBuffers()
		lenAfter := len(f.VisualBuffer)

		if f.bufferIndex+1 > f.GetViewportWidth() {
			f.ViewportOffsetX = f.bufferIndex - f.GetViewportWidth() + 1
			f.apparentCursorX = f.bufferIndex - f.ViewportOffsetX + EditorLeftMargin
		}

		if f.ViewportOffsetY > 0 {
			f.ViewportOffsetY -= lenBefore - lenAfter
		}

		f.apparentCursorY = f.bufferLine + 1 - f.ViewportOffsetY

		return SoftWrapDisabled
	}
}
