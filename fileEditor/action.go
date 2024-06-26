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

	f.apparentCursorX = f.SnapACXToTabBoundary(currIdx, f.apparentCursorX)
}

/*
This function determines if the current ACX is resting within a tab character
and returns the new ACX that is snapped to either the start of end of the tab character
*/
func (f *FileEditor) SnapACXToTabBoundary(currBufferLine int, currACX int) int {
	tabInfoArr, exists := f.TabMap[currBufferLine]
	if !exists {
		return currACX
	}

	/*
		we set the flag to false bc the bufferIndex is aligned the the
		visualBuffer by default, so we're working with the visual index
	*/
	bufferIndex := currACX - EditorLeftMargin + f.ViewportOffsetX
	tabInfo, err := GetTabInfoByIndex(tabInfoArr, bufferIndex, false)
	if err != nil {
		return currACX
	}

	end := tabInfo.End
	start := tabInfo.Start

	dif1 := math.Abs(end - bufferIndex)

	if dif1 <= 1 {
		/*
			if width == 1, and we return end + EditorLeftMargin + 1, then the cursor will ALWAYS
			clamp to the end of the tab even though the cursor is near the start of the tab.
			By subtracting 1, we allow it to clamp the start of the tab and let the next index
			handle making it look like the clamping to the end of the tab (which would be the start
			of the next tab or the end of the line)
		*/
		if tabInfo.TabWidth() == 1 {
			return end + EditorLeftMargin - f.ViewportOffsetX
		}
		return end + EditorLeftMargin + 1 - f.ViewportOffsetX
	}

	return start + EditorLeftMargin - f.ViewportOffsetX
}

func (f FileEditor) MoveToTabBoundary(tabInfoArr []TabInfo, currACX int, cursorDirection string) int {
	if cursorDirection == "left" {
		bufferIndex := currACX - EditorLeftMargin + f.ViewportOffsetX
		tabInfo, err := GetTabInfoByIndex(tabInfoArr, bufferIndex, false)
		if err != nil {
			return math.Clamp(currACX, EditorLeftMargin, f.TermWidth)
		}

		tabWidth := tabInfo.TabWidth()
		return math.Clamp(currACX-tabWidth+1, EditorLeftMargin, f.TermWidth)
	} else {
		bufferIndex := currACX - EditorLeftMargin + f.ViewportOffsetX
		tabInfo, err := GetTabInfoByIndex(tabInfoArr, bufferIndex, false)
		if err != nil {
			return math.Clamp(currACX+1, EditorLeftMargin, f.TermWidth)
		}

		tabWidth := tabInfo.TabWidth()
		return math.Clamp(currACX+tabWidth, EditorLeftMargin, f.TermWidth)
	}
}

func (f *FileEditor) SetCursorPositionOnClick(m MouseInput) byte {
	ansi.HideCursor()
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

	currBufferLine := m.Y - 1 + f.ViewportOffsetY
	line := f.VisualBuffer[currBufferLine]
	currLineLen := len(line)

	isLineInViewport := len(line[math.Min(f.ViewportOffsetX, currLineLen):]) > 0
	if !f.SoftWrap && !isLineInViewport {
		f.ViewportOffsetX = math.Max(currLineLen, 0)
	}
	x := math.Clamp(m.X, EditorLeftMargin, currLineLen+EditorLeftMargin-f.ViewportOffsetX)

	x = f.SnapACXToTabBoundary(currBufferLine, x)

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
			tabInfoArr, exists := f.TabMap[f.bufferLine]
			if !exists {
				f.apparentCursorX--
			} else {
				f.apparentCursorX = f.MoveToTabBoundary(tabInfoArr, f.apparentCursorX-1, "left")
			}
		}
	} else {
		if f.apparentCursorY > 1 { // move to end of previous line
			f.DecrementCursorY()
			line := f.VisualBuffer[f.apparentCursorY-1+f.ViewportOffsetY]
			f.apparentCursorX = len(line) + EditorLeftMargin
			if len(line) >= f.GetViewportWidth() { // scroll screen to end of line if line past screen
				f.ViewportOffsetX = len(line) - f.GetViewportWidth() + 1
				f.apparentCursorX = f.TermWidth
			}
		} else if f.apparentCursorY == 1 && f.ViewportOffsetY > 0 { // begin scrolling up and moving to end of line
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
			tabInfoArr, exists := f.TabMap[f.bufferLine]
			if !exists {
				f.apparentCursorX++
			} else {
				f.apparentCursorX = f.MoveToTabBoundary(tabInfoArr, f.apparentCursorX, "right")
			}
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

	actualBufferIndex := AlignBufferIndex(f.bufferIndex, f.bufferLine, f.TabMap)

	// split the current line
	beforeSplit := line[:actualBufferIndex]
	afterSplit := line[actualBufferIndex:]

	// insert the new line (afterSplit) in the middle of buffer array
	result := make([]string, n+1)

	copy(result, f.FileBuffer[:f.bufferLine])

	result[f.bufferLine] = beforeSplit
	result[f.bufferLine+1] = afterSplit

	copy(result[f.bufferLine+2:], f.FileBuffer[f.bufferLine+1:])

	f.FileBuffer = result

	// update cursor position
	if actualBufferIndex == len(line) { // inserting new line at the end of a line
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

	actualBufferIndex := AlignBufferIndex(f.bufferIndex, f.bufferLine, f.TabMap)

	before := line[:actualBufferIndex]
	after := line[actualBufferIndex:]

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

func (f *FileEditor) actionInsertTab() {
	// f.actionTyping(Tab)

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
				/*
					some margin space added to the offset for lines with a length greater than
					the viewport width, so the user can see where the lines joined
				*/
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
		actualBufferIndex := AlignBufferIndex(f.bufferIndex, f.bufferLine, f.TabMap)
		actualBufferIndex = math.Clamp(actualBufferIndex, 0, len(line)-1)

		var tabKeyDelete bool
		if actualBufferIndex == len(line)-1 {
			tabKeyDelete = line[actualBufferIndex] == Tab
			actualBufferIndex++
		} else {
			tabKeyDelete = line[actualBufferIndex-1] == Tab
		}

		before := line[:actualBufferIndex-1]
		after := line[actualBufferIndex:]

		f.FileBuffer[f.bufferLine] = before + after

		if f.ViewportOffsetX == 0 {
			if tabKeyDelete { // move cursor to left according to tabwidth
				tabInfoArr := f.TabMap[f.bufferLine]
				tabInfo, _ := GetTabInfoByIndex(tabInfoArr, actualBufferIndex-1, true)
				tabWidth := tabInfo.TabWidth()

				f.apparentCursorX = math.Max(f.apparentCursorX-tabWidth, EditorLeftMargin)
			} else {
				f.apparentCursorX--
			}
		} else {
			if tabKeyDelete {
				tabInfoArr := f.TabMap[f.bufferLine]
				tabInfo, _ := GetTabInfoByIndex(tabInfoArr, actualBufferIndex-1, true)
				tabWidth := tabInfo.TabWidth()

				f.apparentCursorX = math.Max(f.apparentCursorX-tabWidth, EditorLeftMargin)
			} else {
				f.ViewportOffsetX = math.Max(f.ViewportOffsetX-1, 0)
			}
		}

		/*
			f.bufferIndex > 1 to avoid automatically moving to prev line when deleting from index <= 1;
			we want the conditional above this to handle that
		*/
		if actualBufferIndex > 1 && f.apparentCursorX <= EditorLeftMargin {
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
