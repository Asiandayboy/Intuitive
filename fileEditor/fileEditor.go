package fileeditor

import (
	"bufio"
	"os"

	"fmt"

	// "github.com/Asiandayboy/CLITextEditor/render"

	"github.com/Asiandayboy/CLITextEditor/render"
	"github.com/Asiandayboy/CLITextEditor/util/ansi"
	"golang.org/x/term"
)

const (
	Reset      string = "\033[0m"
	Black      string = "\033[30m"
	Red        string = "\033[31m"
	Green      string = "\033[32m"
	Yellow     string = "\033[33m"
	Blue       string = "\033[34m"
	Magenta    string = "\033[35m"
	Cyan       string = "\033[36m"
	White      string = "\033[37m"
	Grey       string = "\033[90m"
	Italic     string = "\033[3m"
	TopLCorner string = "\u250C"
	TopRCorner string = "\u2510"
	BotLCorner string = "\u2514"
	BotRCorner string = "\u2518"
	Horizontal string = "\u2500"
	Vertical   string = "\u2502"
)

const (
	Quit byte = iota + 1
	KeyboardInput
	Resize
	EditorModeChange
	CursorPositionChange
)

// the three editor modes a user can be in
const (
	EditorCommandMode uint8 = 'C'
	EditorEditMode    uint8 = 'E'
	EditorViewMode    uint8 = 'V'
)

var modeColors = map[byte]ansi.RGBColor{
	EditorCommandMode: ansi.NewRGBColor(75, 176, 255), // blue
	EditorEditMode:    ansi.NewRGBColor(216, 148, 53), // orange
	EditorViewMode:    ansi.NewRGBColor(158, 75, 253), // purple
}

/*
left margin accounts for the line numbers, and the vertical border

4 for the digits (I doubt a single file will exceed 9999 lines)
1 for the space between the line numbers and the vertical border
1 for the vertical border
1 for the space between the vertical border and the start of the line
1 for the start of the line
*/
const EditorLeftMargin int = 8

type FileEditor struct {
	Saved       bool
	EditorMode  byte
	Keybindings Keybind
	file        *os.File
	Filename    string
	inputChan   chan byte

	FileBuffer         []string // contains each line of the actual file
	VisualBuffer       []string // contains word wrapped lines; this is what gets rendered to the screen
	VisualBufferMapped []int    // contains the ending index (1-indexed) of word-wrapped lines
	apparentCursorX    int
	apparentCursorY    int
	bufferLine         int // refers to current line of FileBuffer; used when editing FileBuffer
	bufferIndex        int // refers to current index of current line of FileBuffer; used when editing FileBuffer
	TermWidth          int
	TermHeight         int
	EditorWidth        int // refers to how many characters can fit on a single line
	MaxTermHeight      int
	StatusBarHeight    int

	ViewportOffsetX int
	ViewportOffsetY int

	maxCursorX int // used to constrain and snap cursor
	maxCursorY int // used to constrain and snap cursor

}

func NewFileEditor(filename string) FileEditor {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}

	return FileEditor{
		Filename:           filename,
		FileBuffer:         make([]string, 0),
		VisualBuffer:       make([]string, 0),
		VisualBufferMapped: make([]int, 0),
		TermWidth:          width,
		TermHeight:         height,
		EditorWidth:        width - EditorLeftMargin + 1,
		MaxTermHeight:      height,
		StatusBarHeight:    3,
		Saved:              false,
		EditorMode:         EditorCommandMode,
		Keybindings:        NewKeybind(),
		inputChan:          make(chan byte, 1),
	}
}

/*
Opens the file or creates a new one if it cannot be found,
and reads its content into the buffer
*/
func (f *FileEditor) OpenFile() {
	file, err := os.Open(f.Filename)
	if err != nil {
		file, _ = os.Create(f.Filename)
	}

	f.file = file
}

func (f *FileEditor) CloseFile() error {
	err := f.file.Close()
	return err
}

func (f FileEditor) GetBufferCharCount() int {
	var count int = 0
	for _, line := range f.FileBuffer {
		count += len(line)
	}

	return count
}

func (f *FileEditor) ReadFileToBuffer() error {
	scanner := bufio.NewScanner(f.file)
	var rowCount int = 0
	for scanner.Scan() {
		f.FileBuffer = append(f.FileBuffer, scanner.Text())
		rowCount++
	}

	if rowCount == 0 {
		f.FileBuffer = append(f.FileBuffer, "")
		rowCount++
	}

	var maxHeight int = f.TermHeight - f.StatusBarHeight
	var maxWidth int = f.TermWidth - EditorLeftMargin

	// initialize bufferLine and bufferIndex
	// initialize mapped visual buffer
	// initalize visual buffer to determine initial cursor position
	var end int = 1
	for i := 0; i < maxHeight; i++ {
		if i >= rowCount {
			break
		}

		line := f.FileBuffer[i]
		n := len(line)

		if n >= maxWidth {
			wrappedLines := f.GetWordWrappedLines(line, maxWidth)

			end += len(wrappedLines) - 1
			f.VisualBuffer = append(f.VisualBuffer, wrappedLines...)
			f.VisualBufferMapped = append(f.VisualBufferMapped, end)
		} else {
			f.VisualBuffer = append(f.VisualBuffer, line)
			f.VisualBufferMapped = append(f.VisualBufferMapped, end)
		}
		end++
	}

	// set cursor's initial position to beginning of file
	f.apparentCursorX = EditorLeftMargin
	f.apparentCursorY = 1

	//// the commented code sets the cursor to end of file
	// n := len(f.VisualBuffer)
	// f.apparentCursorX = len(f.VisualBuffer[n-1]) + EditorLeftMargin
	// f.apparentCursorY = len(f.VisualBuffer)

	f.bufferLine = 0
	f.bufferIndex = 0

	return scanner.Err()
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

func (f *FileEditor) PrintBuffer() {
	lineNumColor := ansi.NewRGBColor(80, 80, 80).ToFgColorANSI()
	borderColor := ansi.NewRGBColor(60, 60, 60).ToFgColorANSI()
	wrappedColor := ansi.NewRGBColor(60, 60, 60).ToFgColorANSI()
	currRowColor := modeColors[f.EditorMode].ToFgColorANSI()

	/*
		The status bar height is subtracted from the terminal height to avoid the unnecessary scrolling,
		which would truncate the beginning of the buffer; And it also refers to the height
		that the status bar will takeup. The maxHeight value will also be decreased the more
		word-wrapped lines there are.
	*/
	var maxHeight int = f.TermHeight - f.StatusBarHeight

	f.maxCursorX = 0
	f.maxCursorY = 0

	// refresh visual buffer and mapped buffer
	f.VisualBuffer = []string{}
	f.VisualBufferMapped = []int{}

	ansi.MoveCursor(0, 0)

	var end int = 1
	for i := 0; i < maxHeight; i++ {
		if i >= len(f.FileBuffer) {
			ansi.EraseEntireLine()
			fmt.Printf("%s   ~ %s%s%s\n", lineNumColor, borderColor, Vertical, Reset)
		} else {
			line := f.FileBuffer[i]

			if len(line) >= f.EditorWidth {
				wordWrappedLines := f.GetWordWrappedLines(line, f.EditorWidth)

				end += len(wordWrappedLines) - 1

				for j, l := range wordWrappedLines {
					ansi.EraseEntireLine()
					if j == 0 {
						if f.bufferLine == i { //  highlighting line number
							fmt.Printf("%s%s%4d %s%s%s %s\n", lineNumColor, currRowColor, i+1, borderColor, Vertical, Reset, l)
						} else {
							fmt.Printf("%s%4d %s%s%s %s\n", lineNumColor, i+1, borderColor, Vertical, Reset, l)
						}
					} else {
						if f.bufferLine == i {
							fmt.Printf("%s   %s", lineNumColor, currRowColor)
						} else {
							fmt.Printf("%s   %s", lineNumColor, wrappedColor)
						}

						if j == len(wordWrappedLines)-1 {
							fmt.Printf("%s %s%s%s %s\n", BotLCorner, borderColor, Vertical, Reset, l)
						} else {
							fmt.Printf("%s %s%s%s %s\n", Vertical, borderColor, Vertical, Reset, l)
						}
					}

					maxHeight--
					f.maxCursorX = len(l) + EditorLeftMargin
					f.maxCursorY += 1
				}
				f.VisualBuffer = append(f.VisualBuffer, wordWrappedLines...)
				f.VisualBufferMapped = append(f.VisualBufferMapped, end)
			} else {
				ansi.EraseEntireLine()
				if f.bufferLine == i {
					fmt.Printf("%s%s%4d %s%s%s %s\n", lineNumColor, currRowColor, i+1, borderColor, Vertical, Reset, line)
				} else {
					fmt.Printf("%s%4d %s%s%s %s\n", lineNumColor, i+1, borderColor, Vertical, Reset, line)
				}
				f.VisualBuffer = append(f.VisualBuffer, line)
				f.VisualBufferMapped = append(f.VisualBufferMapped, end)
				f.maxCursorX = len(line) + EditorLeftMargin
				f.maxCursorY += 1
			}
			end++
		}
	}

}

func (f FileEditor) PrintStatusBar() {
	width := f.TermWidth - EditorLeftMargin + 3
	height := f.StatusBarHeight

	xOffset := EditorLeftMargin - 3
	yOffset := f.TermHeight - height

	borderColor := ansi.NewRGBColor(60, 60, 60)

	// draw the editor mode next to status bar
	render.DrawBox(render.Box{
		Width: 5, Height: height,
		X: 0, Y: yOffset,
		BorderColor: borderColor,
	}, true)

	ansi.MoveCursor(yOffset+2, 2)
	fmt.Printf(modeColors[f.EditorMode].ToFgColorANSI()+"[%c]"+Reset, f.EditorMode)

	// draw the main part of the status bar
	render.DrawBox(render.Box{
		Width: width, Height: height,
		X: xOffset, Y: yOffset,
		BorderColor: borderColor,
	}, true)

	// draw file name
	ansi.MoveCursor(yOffset+2, EditorLeftMargin)
	if !f.Saved {
		fmt.Print(Green + f.Filename + Cyan + Italic + " (Unsaved)" + Reset)
	} else {
		fmt.Print(Green + f.Filename + Reset)
	}

	// draw cursor position
	ansi.MoveCursor(yOffset+2, f.TermWidth-25)
	fmt.Printf(modeColors[f.EditorMode].ToFgColorANSI()+"%d:%d"+Reset, f.apparentCursorY, f.apparentCursorX-EditorLeftMargin+1)

	// draw word count
	wordCountColor := ansi.NewRGBColor(120, 120, 120).ToFgColorANSI()

	ansi.MoveCursor(yOffset+2, f.TermWidth-18)
	fmt.Print(wordCountColor+"Char count: ", f.GetBufferCharCount(), Reset)

	// debugging pu rposes
	ansi.MoveCursor(yOffset+2, f.TermWidth-50)
	fmt.Print("l:", f.bufferLine, " i:", f.bufferIndex)

	ansi.MoveCursorRight(3)
	fmt.Print("Width:", f.EditorWidth)
}

func (f *FileEditor) Render(flag byte) {
	ansi.HideCursor()

	if flag == CursorPositionChange {
		f.UpdateBufferIndicies()
	}

	f.PrintBuffer()

	switch flag {
	case Resize:
		{
			newACX, newACY := CalcNewACXY(
				f.VisualBufferMapped, f.bufferLine,
				f.bufferIndex, f.EditorWidth,
			)

			indexCheck := CalcBufferIndexFromACXY(
				newACX, newACY,
				f.bufferLine, f.VisualBuffer, f.VisualBufferMapped,
			)

			if indexCheck != f.bufferIndex {
				newACX = f.EditorWidth
			}

			f.apparentCursorX = newACX + EditorLeftMargin - 1
			f.apparentCursorY = newACY

		}
	}

	f.PrintStatusBar()
	ansi.MoveCursor(f.apparentCursorY, f.apparentCursorX)

	ansi.ShowCursor()
}

func (editor *FileEditor) UpdateLoop() {
updateLoop:
	for {
		termW, termH, _ := term.GetSize(int(os.Stdout.Fd()))

		if !(termW == editor.TermWidth && termH == editor.TermHeight) { // resize
			editor.inputChan <- Resize
		}

		select {
		case code := <-editor.inputChan:
			switch code {
			case Quit:
				break updateLoop
			case KeyboardInput:
				editor.Render(KeyboardInput)
			case EditorModeChange:
				editor.Render(EditorModeChange)
			case Resize:
				ansi.ClearEntireScreen()
				editor.TermHeight = termH
				editor.TermWidth = termW
				editor.EditorWidth = termW - EditorLeftMargin + 1
				editor.Render(Resize)
			case CursorPositionChange:
				editor.Render(CursorPositionChange)
			}
		default:
		}
	}
}

func (editor *FileEditor) OnInput() int {
	var buf [16]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		fmt.Println("Error reading from stdin:", err)
		return 1
	}

	if buf[0] == Escape {
		isMouseInput, mouseEvent := ReadEscSequence(buf[:], n)

		if isMouseInput {
			ret := HandleMouseInput(editor, mouseEvent)

			switch ret {
			case CursorPositionChange:
				editor.inputChan <- ret
			}
		} else {
			ret := HandleEscapeInput(editor, buf[:], n)

			switch ret {
			case CursorPositionChange:
				editor.inputChan <- CursorPositionChange
			case EditorModeChange:
				editor.inputChan <- EditorModeChange
			}
		}
	} else {
		ret := HandleKeyboardInput(editor, buf[0])

		switch ret {
		case Quit:
			editor.inputChan <- Quit
			return 1
		case EditorModeChange:
			editor.inputChan <- EditorModeChange
		case KeyboardInput:
			editor.inputChan <- KeyboardInput
		}

	}

	return 0
}

// func drawBox(width uint16, height uint16, words [][]string) {
// 	xOffset := "     "

// 	// draw top border
// 	fmt.Print(xOffset)
// 	fmt.Print(Grey + TopLCorner + Reset)
// 	var i uint16
// 	for i = 0; i < width; i++ {
// 		fmt.Print(Grey + Horizontal + Reset)
// 		if i+1 == width {
// 			fmt.Println(Grey + TopRCorner + Reset)
// 		}
// 	}

// 	// draw side borders and words
// 	for i = 0; i < height; i++ {
// 		fmt.Print(xOffset)
// 		fmt.Print(Grey + Vertical + Reset)
// 		var totalFormatedWidth uint16 = width

// 		firstWords := words[i]
// 		for _, word := range firstWords {
// 			wordLen := uint16(len(word))
// 			if wordLen >= 50 {
// 				panic(fmt.Sprintf(`"%s" is bigger than the specified box width.`, word))
// 			}

// 			if i == 3 { // row containing filename
// 				fmt.Printf(" Filename: %s%s %s%s(Unsaved)%s", Green, word, Italic, Cyan, Reset)
// 				wordLen += 21
// 			} else {
// 				fmt.Print(word)
// 			}
// 			totalFormatedWidth -= wordLen
// 		}

// 		formatted := fmt.Sprintf("%%%ds", totalFormatedWidth)
// 		fmt.Printf("%s%s\n", fmt.Sprintf(formatted, " "), Grey+Vertical+Reset)
// 	}

// 	// draw bottom border
// 	fmt.Print(xOffset)
// 	fmt.Print(Grey + BotLCorner + Reset)
// 	for i = 0; i < width; i++ {
// 		fmt.Print(Grey + Horizontal + Reset)
// 		if i+1 == width {
// 			fmt.Println(Grey + BotRCorner + Reset)
// 		}
// 	}
// }

/*
	n, _ := os.Stdin.Read(buffer)

	if n == 1 {
		switch buffer[0] {
		case 'q':
			fmt.Print("\x1b[H")
			fmt.Print("\x1b[2J") //clear the screen
			return 0
		case 127: // backspace
			if editor.CursorX > 0 {
				currIndex := editor.CursorY + editor.ViewportOffsetY
				line := editor.Buffer[currIndex]
				editor.Buffer[currIndex] = line[:editor.CursorX-1] + line[editor.CursorX:]
				editor.CursorX--
			} else { // move cursor to end of previous line
				if editor.CursorY > 0 {
					currIndex := editor.CursorY + editor.ViewportOffsetY

					// delete row from buffer
					editor.Buffer = append(editor.Buffer[:currIndex], editor.Buffer[currIndex+1:]...)

					editor.CursorY--
					editor.CursorX = len(editor.Buffer[editor.CursorY+editor.ViewportOffsetY])

					if editor.ViewportOffsetY > 0 {
						editor.ViewportOffsetY--
						editor.CursorY++
					}
				}
			}
		case 13: // enter
			editor.Buffer = append(editor.Buffer, "")
			editor.CursorX = 0
			if editor.CursorY < editor.MaxTermHeight-1 {
				editor.CursorY++
			} else { // cursor extends passed viewport height; move virtul screen up
				editor.ViewportOffsetY++
			}

		default:
			if len(editor.Buffer) == 0 {
				editor.Buffer = append(editor.Buffer, string(buffer[0]))
				editor.CursorX++
			} else {
				currIndex := editor.CursorY + editor.ViewportOffsetY
				line := editor.Buffer[currIndex]
				editor.Buffer[currIndex] = line[:editor.CursorX] + string(buffer[0]) + line[editor.CursorX:]
				editor.CursorX++

				// word wrap
				if editor.CursorX >= editor.TermWidth-editor.ViewportOffsetX {
					editor.Buffer = append(editor.Buffer, "")
					editor.CursorY++
					editor.CursorX = 0

					if editor.CursorY >= editor.MaxTermHeight-1 {
						editor.ViewportOffsetY++
						editor.CursorY--
					}
				}
			}

		}
	} else if n == 3 {
		if buffer[0] == '\x1b' && buffer[1] == '[' {
			key := buffer[2]

			switch key {
			case 'A': // up
			case 'B': // down
			case 'C': // right
				editor.CursorX++
			case 'D': // left
				editor.CursorX--
			}
		}

	}
*/
