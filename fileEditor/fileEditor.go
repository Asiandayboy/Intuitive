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
	WindowResize
	EditorModeChange
	CursorPositionChange
	NewLineInserted
	NewLineInsertedAtLineEnd
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
	apparentCursorX    int      // cursor's X position
	apparentCursorY    int      // cursor's Y position
	bufferLine         int      // refers to current line of FileBuffer; used when editing FileBuffer
	bufferIndex        int      // refers to current index of current line of FileBuffer; used when editing FileBuffer
	TermWidth          int      // width of the terminal window
	TermHeight         int      // height of the terminal window
	EditorWidth        int      // refers to how many characters can fit on a single line
	StatusBarHeight    int      // height of the status bar

	ViewportOffsetX int // used for horizontal scrolling
	ViewportOffsetY int // used for vertical scrolling
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

	// initalize visual buffers to determine initial cursor position
	f.RefreshVisualBuffers()

	// set cursor's initial position to beginning of file
	f.apparentCursorX = EditorLeftMargin
	f.apparentCursorY = 1
	f.bufferLine = 0
	f.bufferIndex = 0

	return scanner.Err()
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

	ansi.MoveCursor(0, 0)

	var lastIdx int = -1
	for i, line := range f.VisualBuffer {
		ansi.EraseEntireLine()

		currIdx := CalcBufferLineFromACY(i+1, f.VisualBufferMapped)
		if lastIdx == currIdx {
			if f.bufferLine == currIdx {
				fmt.Printf("%s   %s", lineNumColor, currRowColor)
			} else {
				fmt.Printf("%s   %s", lineNumColor, wrappedColor)
			}

			nextIdx := CalcBufferLineFromACY(i+2, f.VisualBufferMapped)

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

		maxHeight--
		lastIdx = currIdx
	}

	// print the remaining empty spaces
	for range maxHeight {
		ansi.EraseEntireLine()
		fmt.Printf("%s   ~ %s%s%s\n", lineNumColor, borderColor, Vertical, Reset)
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
}

func (f *FileEditor) Render(flag byte) {
	ansi.HideCursor()

	// visual buffers are already refreshed when a new line is inserted at the end of a line
	if flag != NewLineInsertedAtLineEnd {
		f.RefreshVisualBuffers()
	}

	if flag == CursorPositionChange ||
		flag == KeyboardInput ||
		flag == NewLineInserted ||
		flag == NewLineInsertedAtLineEnd {

		f.UpdateBufferIndicies()
	}

	f.PrintBuffer()

	switch flag {
	case WindowResize:
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

func (editor *FileEditor) RenderLoop() {
updateLoop:
	for {
		termW, termH, _ := term.GetSize(int(os.Stdout.Fd()))

		if !(termW == editor.TermWidth && termH == editor.TermHeight) { // resize
			editor.inputChan <- WindowResize
		}

		select {
		case code := <-editor.inputChan:
			switch code {
			case Quit:
				break updateLoop
			case KeyboardInput, EditorModeChange, CursorPositionChange,
				NewLineInserted, NewLineInsertedAtLineEnd:
				editor.Render(code)
			case WindowResize:
				ansi.ClearEntireScreen()
				editor.TermHeight = termH
				editor.TermWidth = termW
				editor.EditorWidth = termW - EditorLeftMargin + 1
				editor.Render(WindowResize)
			}
		default:
		}
	}
}

func (editor *FileEditor) InputLoop() int {
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
		default:
			editor.inputChan <- ret
		}

	}

	return 0
}
