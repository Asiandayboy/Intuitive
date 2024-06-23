package fileeditor

import (
	"fmt"

	"github.com/Asiandayboy/CLITextEditor/util/ansi"
	"github.com/Asiandayboy/CLITextEditor/util/math"
)

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
This function replaces every \t character (ASCII 9) in a line of a file
and replaces it with spaces defined by the tabsize, ensuring that the tab
is formatted such that a tab occurs at every interval defined by the tabsize.

During the process, an array containing the start and end index of each tab in
the line is constructed and returned with the new line.
*/
func (f *FileEditor) RenderTabCharWithSpaces(line string, lineNum int) (l string) {
	var tabPosArr []int
	/*
		We must keep track of the current tab index we are on to
		ensure tabs are occuring at the interval defined by tabsize.
		This is what the tabIntervalCount variable is for.

		While we're doing this, we must also build the tab map as well
		so that we can use it for when indent is using tabs
	*/
	var tabIntervalCount uint8 = 0
	for i, char := range line {
		if byte(char) == Tab {
			/*
				If a tab is the first character in the line (at index 0), then we need to make
				sure it's starting from 1 bc tabIntervalCount is 1-indexed
			*/
			if tabIntervalCount <= 0 && i == 0 {
				tabIntervalCount = 1
			}

			k := int(tabIntervalCount) - 1
			tabWidth := f.TabSize - (tabIntervalCount % f.TabSize)
			tabIntervalCount += tabWidth

			tabPosArr = append(tabPosArr, int(tabIntervalCount)-1)
			tabPosArr = append(tabPosArr, k)

			for range tabWidth { // render tab characters as tabsize x spaces
				l += " "
			}
		} else {
			l += string(char)
			tabIntervalCount++
		}
	}

	if len(tabPosArr) > 0 {
		f.TabMap[lineNum] = tabPosArr
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
	f.TabMap = make(map[int][]int)

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
	f.TabMap = make(map[int][]int)

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
