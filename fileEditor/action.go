package fileeditor

import (
	"github.com/Asiandayboy/CLITextEditor/util/ansi"
	"github.com/Asiandayboy/CLITextEditor/util/math"
)

type actionFunc func()

const (
	upDirection   uint8 = 1
	downDirection uint8 = 2
	noDirection   uint8 = 3
)

var savedACX int = 0
var savedCursorXFlag bool = false
var savedViewportXOffset int = 0

func setSavedCursorX(acX int, xOffset int, saveFlag bool) {
	savedCursorXFlag = saveFlag
	savedACX = acX
	savedViewportXOffset = xOffset
}

/*
This function determines if the current ACX is resting within a tab character
and returns the new ACX that is snapped to either the start of end of the tab character
and returns the TabInfo
*/
func (f *FileEditor) SnapACXToTabBoundary(currBufferLine int, currACX int, direction uint8) (int, TabInfo) {
	tabInfoArr, exists := f.TabMap[currBufferLine]
	if !exists {
		return currACX, TabInfo{}
	}

	var bufferIndex int
	if f.SoftWrap {
		var y int = f.apparentCursorY
		if direction == upDirection {
			y = f.apparentCursorY + 1
		} else if direction == downDirection {
			y = f.apparentCursorY - 1
		}

		bufferIndex = CalcBufferIndexFromACXY(
			currACX-EditorLeftMargin+1, y,
			f.bufferLine, f.VisualBuffer, f.VisualBufferMapped, f.ViewportOffsetY,
		)
	} else {
		bufferIndex = currACX - EditorLeftMargin + f.ViewportOffsetX
	}

	/*
		we set the flag to false bc the bufferIndex is aligned the the
		visualBuffer by default, so we're working with the visual index
	*/
	tabInfo, err := GetTabInfoByIndex(tabInfoArr, bufferIndex, false)
	if err != nil {
		return currACX, TabInfo{}
	}

	end := tabInfo.End
	start := tabInfo.Start

	dif1 := math.Abs(end - bufferIndex)

	var ret int

	if dif1 < 1 {
		/*
			if width == 1, and we return end + EditorLeftMargin + 1, then the cursor will ALWAYS
			clamp to the end of the tab even though the cursor is near the start of the tab.
			By subtracting 1, we allow it to clamp the start of the tab and let the next index
			handle making it look like the clamping to the end of the tab (which would be the start
			of the next tab or the end of the line)
		*/
		if tabInfo.TabWidth() == 1 {
			ret = end + EditorLeftMargin - f.ViewportOffsetX
		} else {
			ret = end + EditorLeftMargin + 1 - f.ViewportOffsetX
		}
	} else {
		ret = start + EditorLeftMargin - f.ViewportOffsetX
	}

	if f.SoftWrap && ret > (f.TermWidth) {
		return ret % (f.TermWidth - EditorLeftMargin), tabInfo
	}

	return ret, tabInfo
}

func constrainCursorX(f *FileEditor, direction uint8) {
	/*
		Constrain cursor when moving cursor up and down
		with keys, keeping the acX clamped from its initial position
		on the first up or down
	*/
	visualLineIdx := f.apparentCursorY - 1 + f.ViewportOffsetY
	currLine := f.VisualBuffer[visualLineIdx]

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
				prevLine := f.VisualBuffer[visualLineIdx+1]
				if len(currLine) > len(prevLine) {
					f.ViewportOffsetX = lineLength - f.TermWidth
				}
			} else if direction == downDirection && visualLineIdx > 0 {
				prevLine := f.VisualBuffer[visualLineIdx-1]
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

	if !f.SoftWrap {
		f.apparentCursorX, _ = f.SnapACXToTabBoundary(visualLineIdx, f.apparentCursorX, direction)
	} else {
		bufferLine := CalcBufferLineFromACY(f.apparentCursorY, f.VisualBufferMapped, f.ViewportOffsetY)

		upBufferLine := CalcBufferLineFromACY(f.apparentCursorY+1, f.VisualBufferMapped, f.ViewportOffsetY)
		downBufferLine := CalcBufferLineFromACY(f.apparentCursorY-1, f.VisualBufferMapped, f.ViewportOffsetY)

		if (direction == upDirection && bufferLine == upBufferLine) ||
			(direction == downDirection && bufferLine == downBufferLine) { // moving up/down on the same bufferLine

			x, tabInfo := f.SnapACXToTabBoundary(bufferLine, f.apparentCursorX, noDirection)
			empty := TabInfo{}

			bufIdx := CalcBufferIndexFromACXY(
				f.apparentCursorX, f.apparentCursorY,
				f.bufferLine, f.VisualBuffer, f.VisualBufferMapped, f.ViewportOffsetY,
			)

			if bufIdx <= f.TermWidth && tabInfo != empty {
				if f.apparentCursorX-EditorLeftMargin > 1 {
					if direction == downDirection {
						f.DecrementCursorY()
					}
				} else {
					f.DecrementCursorY()
				}
			}

			f.apparentCursorX = x
		} else {
			f.apparentCursorX, _ = f.SnapACXToTabBoundary(bufferLine, f.apparentCursorX, direction)
		}
	}
}

func (f *FileEditor) MoveToTabBoundary(tabInfoArr []TabInfo, currACX int, cursorDirection string) int {
	var bufferIndex int
	if f.SoftWrap {
		bufferIndex = CalcBufferIndexFromACXY(
			currACX-EditorLeftMargin+1, f.apparentCursorY,
			f.bufferLine, f.VisualBuffer, f.VisualBufferMapped, f.ViewportOffsetY,
		)
	} else {
		bufferIndex = currACX - EditorLeftMargin + f.ViewportOffsetX
	}

	if cursorDirection == "left" {
		tabInfo, err := GetTabInfoByIndex(tabInfoArr, bufferIndex, false)
		if err != nil {
			return math.Clamp(currACX, EditorLeftMargin, f.TermWidth)
		}

		tabWidth := tabInfo.TabWidth()
		return math.Clamp(currACX-tabWidth+1, EditorLeftMargin, f.TermWidth)
	} else {
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

	if !f.SoftWrap {
		x, _ = f.SnapACXToTabBoundary(currBufferLine, x, noDirection)
	} else {
		x, _ = f.SnapACXToTabBoundary(CalcBufferLineFromACY(m.Y, f.VisualBufferMapped, f.ViewportOffsetY), x, noDirection)
	}

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
				tabInfo, err := GetTabInfoByIndex(tabInfoArr, f.bufferIndex-1, false)
				if err != nil {
					f.apparentCursorX = f.MoveToTabBoundary(tabInfoArr, f.apparentCursorX-1, "left")
					setSavedCursorX(f.apparentCursorX, f.ViewportOffsetX, false)
					return
				}

				tabWidth := tabInfo.TabWidth()
				softWrapTabDif := f.apparentCursorX - tabWidth

				if softWrapTabDif < EditorLeftMargin { // deleting a tab that started at the end of the prev line
					f.DecrementCursorY()
					if f.apparentCursorY == 1 && f.ViewportOffsetY > 0 {
						f.actionScrollUp()
					}
					f.apparentCursorX = f.TermWidth - (EditorLeftMargin - softWrapTabDif)
				} else {
					f.apparentCursorX = f.MoveToTabBoundary(tabInfoArr, f.apparentCursorX-1, "left")
				}
			}
		}
	} else {
		if f.apparentCursorY > 1 { // move to end of previous line
			f.DecrementCursorY()
			line := f.VisualBuffer[f.apparentCursorY-1+f.ViewportOffsetY]
			f.apparentCursorX = len(line) + EditorLeftMargin
			if !f.SoftWrap && len(line) >= f.GetViewportWidth() { // scroll screen to end of line if line past screen
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
		tabInfoArr, exists := f.TabMap[f.bufferLine]
		if !exists {
			f.apparentCursorX++
			setSavedCursorX(f.apparentCursorX, f.ViewportOffsetX, false)
			return
		}

		tabInfo, err := GetTabInfoByIndex(tabInfoArr, f.bufferIndex, false)
		if err != nil {
			if f.apparentCursorX == f.TermWidth {
				f.ViewportOffsetX++
			} else {
				f.apparentCursorX = f.MoveToTabBoundary(tabInfoArr, f.apparentCursorX, "right")
			}
			setSavedCursorX(f.apparentCursorX, f.ViewportOffsetX, false)
			return
		}

		tabWidth := tabInfo.TabWidth()
		softWrapTabDif := f.apparentCursorX + tabWidth

		if f.apparentCursorX == f.TermWidth {
			f.ViewportOffsetX += tabWidth
		} else {
			if f.SoftWrap && softWrapTabDif > f.TermWidth { // adding a tab that continues to the next line
				f.IncrementCursorY()
				if f.apparentCursorY == f.GetViewportHeight() && f.apparentCursorY+f.ViewportOffsetY <= len(f.VisualBuffer) {
					f.actionScrollDown()
				}
				f.apparentCursorX = EditorLeftMargin + (softWrapTabDif - f.TermWidth)
			} else {
				if f.apparentCursorX+tabWidth > f.TermWidth {
					f.ViewportOffsetX += (f.apparentCursorX + tabWidth - f.TermWidth)
				}
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
	if f.TabIndentType == IndentWithSpace {
		index := f.apparentCursorX + f.ViewportOffsetX - EditorLeftMargin
		tabWidth := f.GetSpaceWidthOfTabChar(index)
		for range tabWidth {
			f.actionTyping(Space)
		}
	} else if f.TabIndentType == IndentWithTab {
		/*
			The following lines are taken from actionTyping() with additions to
			handle inserting tab characters. I didn't want to write them in the same
			func bc it looked ugly
		*/
		line := f.FileBuffer[f.bufferLine]

		actualBufferIndex := AlignBufferIndex(f.bufferIndex, f.bufferLine, f.TabMap)

		before := line[:actualBufferIndex]
		after := line[actualBufferIndex:]

		f.FileBuffer[f.bufferLine] = before + string(Tab) + after

		var tabVisualIndex int
		if f.SoftWrap {
			tabVisualIndex = CalcBufferIndexFromACXY(
				f.apparentCursorX-EditorLeftMargin+1, f.apparentCursorY,
				f.bufferLine, f.VisualBuffer, f.VisualBufferMapped, f.ViewportOffsetY,
			)
		} else {
			tabVisualIndex = f.apparentCursorX - EditorLeftMargin + f.ViewportOffsetX
		}
		tabWidth := f.GetSpaceWidthOfTabChar(tabVisualIndex)

		f.apparentCursorX += tabWidth

		if f.apparentCursorX > f.TermWidth {
			/*
				We need to know how much the tab exceeded the terminal width; we'll take that
				value and add it to the cursor or viewport offset, depending if soft wrap is on or not
			*/
			tabDif := f.apparentCursorX - f.TermWidth

			if f.SoftWrap {
				// BUG WHEN INSERTING TAB WITH SOFT WRAP ON
				// it seems it's always a difference of 1
				if tabWidth == 1 {
					f.apparentCursorX = EditorLeftMargin
				} else {
					f.apparentCursorX = EditorLeftMargin + tabDif
				}
				f.IncrementCursorY()

				if f.apparentCursorY == f.GetViewportHeight() && f.apparentCursorY+f.ViewportOffsetY <= len(f.VisualBuffer) {
					f.ViewportOffsetY++ // changing it directly bc the condition in actionScrollDown doesn't apply here, ig
				}
			} else {
				f.apparentCursorX = f.TermWidth
				f.ViewportOffsetX += tabDif
			}
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

		var isDeletingTabKey bool
		if actualBufferIndex == len(line)-1 {
			isDeletingTabKey = line[actualBufferIndex] == Tab
			actualBufferIndex++
		} else {
			isDeletingTabKey = line[actualBufferIndex-1] == Tab
		}

		before := line[:actualBufferIndex-1]
		after := line[actualBufferIndex:]

		f.FileBuffer[f.bufferLine] = before + after

		var softWrapTabDif int

		if f.ViewportOffsetX == 0 {
			if isDeletingTabKey { // move cursor to left according to tabwidth
				tabInfoArr := f.TabMap[f.bufferLine]
				tabInfo, _ := GetTabInfoByIndex(tabInfoArr, actualBufferIndex-1, true)
				tabWidth := tabInfo.TabWidth()

				softWrapTabDif = f.apparentCursorX - tabWidth

				f.apparentCursorX = math.Max(softWrapTabDif, EditorLeftMargin)
			} else {
				f.apparentCursorX--
			}
		} else {
			if isDeletingTabKey {
				tabInfoArr := f.TabMap[f.bufferLine]
				tabInfo, _ := GetTabInfoByIndex(tabInfoArr, actualBufferIndex-1, true)
				tabWidth := tabInfo.TabWidth()

				dif := f.ViewportOffsetX - tabWidth

				f.ViewportOffsetX = math.Max(dif, 0)
				if f.ViewportOffsetX == 0 {
					f.apparentCursorX -= math.Abs(dif)
				}
			} else {
				f.ViewportOffsetX = math.Max(f.ViewportOffsetX-1, 0)
			}
		}

		/*
			actualBufferIndex > 1 to avoid automatically moving to prev line when deleting from index <= 1;
			we want the conditional above this to handle that
		*/
		if actualBufferIndex > 1 && f.apparentCursorX <= EditorLeftMargin {
			f.DecrementCursorY()
			if f.apparentCursorY == 1 && f.ViewportOffsetY > 0 {
				f.actionScrollUp()
			}
			if isDeletingTabKey {
				f.apparentCursorX = f.TermWidth - (EditorLeftMargin - softWrapTabDif)
			} else {
				f.apparentCursorX = len(f.VisualBuffer[f.apparentCursorY-1+f.ViewportOffsetY]) + EditorLeftMargin
			}
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
