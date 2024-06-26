package fileeditor

import (
	"fmt"

	"github.com/Asiandayboy/CLITextEditor/util/ansi"
	"github.com/Asiandayboy/CLITextEditor/util/math"
)

type TabInfo struct {
	BufferIndex int
	Start       int
	End         int
}

func (t TabInfo) TabWidth() int {
	return t.End - t.Start + 1
}

const TabInfoErrMsg string = "A tab character does not exist at the visual index"

type TabInfoErr struct {
	msg string
}

func (t TabInfoErr) Error() string {
	return t.msg
}

type TabMapType map[int][]TabInfo

/*
Returns the TabInfo of the tab that occupies the visual index.
If a TabInfo cannot be found, an error is returned

If the bufferIdxFlag is set to true, the index will be used to search
for the TabInfo based on the bufferIndex and not the end and start visual indicies
*/
func GetTabInfoByIndex(tabInfoArr []TabInfo, index int, bufferIdxFlag bool) (TabInfo, error) {
	left := 0
	right := len(tabInfoArr) - 1

	for left <= right {
		mid := (left + right) / 2

		tab := tabInfoArr[mid]

		if !bufferIdxFlag {
			if index > tab.End {
				left = mid + 1
			} else if index < tab.Start {
				right = mid - 1
			} else {
				return tab, nil
			}
		} else {
			if index > tab.BufferIndex {
				left = mid + 1
			} else if index < tab.BufferIndex {
				right = mid - 1
			} else {
				return tab, nil
			}
		}
	}

	return TabInfo{}, TabInfoErr{msg: TabInfoErrMsg}
}

var lineNumColor string = ansi.NewRGBColor(80, 80, 80).ToFgColorANSI()
var borderColor string = ansi.NewRGBColor(60, 60, 60).ToFgColorANSI()
var wrappedColor string = ansi.NewRGBColor(60, 60, 60).ToFgColorANSI()

func (f FileEditor) GetBufferCharCount() int {
	var count int = 0
	for _, line := range f.FileBuffer {
		count += len(line)
	}

	return count
}

/*
Returns the number of spaces a tab character is rendered with
according to its visual index in the line
*/
func (f FileEditor) GetSpaceWidthOfTabChar(visualIndex int) int {
	return int(f.TabSize) - (visualIndex % int(f.TabSize))
}

/*
This function replaces every \t character (ASCII 9) in a line of a file
and replaces it with spaces defined by the tabsize, ensuring that the tab
is formatted such that a tab occurs at every interval defined by the tabsize.

During the process, an array containing the start and end index of each tab in
the line is constructed and returned with the new line.
*/
func (f *FileEditor) RenderTabCharWithSpaces(line string, lineNum int) (l string) {
	lineTabArr := make([]TabInfo, 0)
	/*
		We must keep track of the current tab index we are on to
		ensure tabs are occuring at the interval defined by tabsize.
		This is what the tabIntervalCount variable is for.

		While we're doing this, we must also build the tab map as well
		so that we can use it for when indent is using tabs
	*/
	var tabIntervalCount int = 0
	for i, char := range line {
		if byte(char) == Tab {
			start := int(tabIntervalCount)

			tabWidth := f.GetSpaceWidthOfTabChar(tabIntervalCount)
			tabIntervalCount += tabWidth

			end := int(tabIntervalCount) - 1

			tabInfo := TabInfo{
				BufferIndex: i,
				Start:       start,
				End:         end,
			}

			lineTabArr = append(lineTabArr, tabInfo)

			for range tabWidth { // render tab characters as tabsize x spaces
				l += " "
			}
		} else {
			l += string(char)
			tabIntervalCount++
		}
	}

	if len(lineTabArr) > 0 {
		f.TabMap[lineNum] = lineTabArr
	}

	return l
}

func (f *FileEditor) GetWordWrappedLines(line string, maxWidth int) (lines []string) {
	length := len(line)

	for length >= maxWidth {
		dif := maxWidth - length - 1
		cutoffIndex := length + dif
		lines = append(lines, line[:cutoffIndex])
		line = line[cutoffIndex:]
		length = len(line)
	}

	/*
		when the loop breaks, we need to add the last remaining line,
		which will be less than maxWidth
	*/
	lines = append(lines, line)

	return lines
}

func (f *FileEditor) RefreshSoftWrapVisualBuffers() {
	f.VisualBuffer = []string{}
	f.VisualBufferMapped = []int{}
	f.TabMap = make(TabMapType)

	var end int = 1

	viewportWidth := f.GetViewportWidth()

	for i, line := range f.FileBuffer {
		line = f.RenderTabCharWithSpaces(line, i)
		if len(line) >= viewportWidth {
			wordWrappedLines := f.GetWordWrappedLines(line, viewportWidth)

			end += len(wordWrappedLines) - 1

			f.VisualBuffer = append(f.VisualBuffer, wordWrappedLines...)
			f.VisualBufferMapped = append(f.VisualBufferMapped, end)
		} else {
			f.VisualBuffer = append(f.VisualBuffer, line)
			f.VisualBufferMapped = append(f.VisualBufferMapped, end)
		}
		end++
	}
}

func (f *FileEditor) RefreshNoWrapVisualBuffers() {
	f.VisualBufferMapped = nil
	f.VisualBuffer = make([]string, len(f.FileBuffer))
	f.TabMap = make(TabMapType)

	for i, line := range f.FileBuffer {
		line = f.RenderTabCharWithSpaces(line, i)
		f.VisualBuffer[i] = line
	}
}

func (f *FileEditor) PrintBuffer() {
	currRowColor := modeColors[f.EditorMode].ToFgColorANSI()

	/*
		The status bar height is subtracted from the terminal height to avoid the unnecessary scrolling,
		which would truncate the beginning of the buffer; The maxHeight value will be decreased for each
		word-wrapped line there is.
	*/
	var maxHeight int = f.TermHeight - f.StatusBarHeight

	ansi.MoveCursor(0, 0)

	var lastIdx int = -1 // only used for soft-wrap
	var linesPrinted = 0
	for i := f.ViewportOffsetY; i < len(f.VisualBuffer); i++ {
		/*
			only print the number of lines that can fit within the viewport, or else the screen
			will scroll down and we won't be able to see what we're supposed to be seeing

			linesPrinted will count how many lines of the visual buffer we've printed so far
		*/
		if linesPrinted == f.GetViewportHeight() {
			break
		}
		var line string
		if f.SoftWrap {
			line = f.VisualBuffer[i]
		} else {
			line = f.VisualBuffer[i][math.Min(f.ViewportOffsetX, len(f.VisualBuffer[i])):math.Min(
				f.ViewportOffsetX+f.GetViewportWidth()-1, len(f.VisualBuffer[i]))]
		}
		ansi.EraseEntireLine()

		if f.SoftWrap {
			currIdx := CalcBufferLineFromACY(i+1, f.VisualBufferMapped, 0)
			if lastIdx == currIdx {
				if f.bufferLine == currIdx {
					fmt.Printf("%s   %s", lineNumColor, currRowColor)
				} else {
					fmt.Printf("%s   %s", lineNumColor, wrappedColor)
				}

				nextIdx := CalcBufferLineFromACY(i+2, f.VisualBufferMapped, 0)

				if currIdx != nextIdx || i+1 == f.VisualBufferMapped[len(f.VisualBufferMapped)-1] {
					fmt.Printf("%s %s%s%s %s\n", BotLCorner, borderColor, Vertical, Reset, line)
				} else {
					fmt.Printf("%s %s%s%s %s\n", Vertical, borderColor, Vertical, Reset, line)
				}

			} else {
				if f.bufferLine == currIdx {
					fmt.Printf("%s%s%4d %s%s%s %s\n", lineNumColor, currRowColor, currIdx+1, borderColor, Vertical, Reset, line)
				} else {
					fmt.Printf("%s%4d %s%s%s %s\n", lineNumColor, currIdx+1, borderColor, Vertical, Reset, line)
				}
			}
			lastIdx = currIdx
		} else {
			if f.bufferLine == i {
				fmt.Printf("%s%s%4d %s%s%s %s\n", lineNumColor, currRowColor, i+1, borderColor, Vertical, Reset, line)
			} else {
				fmt.Printf("%s%4d %s%s%s %s\n", lineNumColor, i+1, borderColor, Vertical, Reset, line)
			}
		}

		linesPrinted++
		maxHeight--
	}

	// print the remaining empty spaces (if there is any in the viewport space avaiable)
	for range maxHeight {
		ansi.EraseEntireLine()
		if f.PrintEmptyLines {
			fmt.Printf("%s   ~ %s%s%s\n", lineNumColor, borderColor, Vertical, Reset)
		} else {
			ansi.MoveCursorDown(1)
		}
	}
}
