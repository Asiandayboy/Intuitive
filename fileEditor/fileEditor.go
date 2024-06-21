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
	EditorEditMode:    ansi.NewRGBColor(228, 37, 132), // orange
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
	StatusBarHeight    int      // height of the status bar
	ViewportOffsetX    int      // used for horizontal scrolling
	ViewportOffsetY    int      // used for vertical scrolling

	// Configs
	SoftWrap        bool
	PrintEmptyLines bool
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
		StatusBarHeight:    3,
		Saved:              false,
		EditorMode:         EditorCommandMode,
		Keybindings:        NewKeybind(),
		inputChan:          make(chan byte, 1),
		SoftWrap:           false,
		PrintEmptyLines:    false,
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
	if f.SoftWrap {
		f.RefreshSoftWrapVisualBuffers()
	} else {
		f.RefreshNoWrapVisualBuffers()
	}

	// set cursor's initial position to beginning of file
	f.apparentCursorX = EditorLeftMargin
	f.apparentCursorY = 1
	f.bufferLine = 0
	f.bufferIndex = 0

	return scanner.Err()
}

func (f FileEditor) PrintStatusBar() {
	width := f.TermWidth - EditorLeftMargin + 3
	height := f.StatusBarHeight

	xOffset := EditorLeftMargin - 3
	yOffset := f.TermHeight - height

	borderColor := ansi.NewRGBColor(60, 60, 60)

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

	// draw buffer indicies position + 1
	ansi.MoveCursor(yOffset+2, f.TermWidth-8)
	fmt.Printf(modeColors[f.EditorMode].ToFgColorANSI()+"%d:%d"+Reset, f.bufferLine+1, f.bufferIndex+1)

	// // draw word count
	// wordCountColor := ansi.NewRGBColor(120, 120, 120).ToFgColorANSI()
	// ansi.MoveCursor(yOffset+2, f.TermWidth-18)
	// fmt.Print(wordCountColor+"Char count: ", f.GetBufferCharCount(), Reset)

	// debugging purposes
	ansi.MoveCursor(yOffset+2, f.TermWidth-50)
	fmt.Print("Soft Wrap: ", f.SoftWrap)
	// ansi.MoveCursor(yOffset+2, f.TermWidth-65)
	// fmt.Print("OY:", f.ViewportOffsetY, " OX:", f.ViewportOffsetX)

	// draw the editor mode next to status bar
	render.DrawBox(render.Box{
		Width: 5, Height: height,
		X: 0, Y: yOffset,
		BorderColor: borderColor,
	}, true)

	ansi.MoveCursor(yOffset+2, 2)
	fmt.Printf(modeColors[f.EditorMode].ToFgColorANSI()+"[%c]"+Reset, f.EditorMode)
}

func (f *FileEditor) Render(flag byte) {
	ansi.HideCursor()

	// visual buffers are already refreshed when a new line is inserted at the end of a line
	if f.SoftWrap && flag != NewLineInsertedAtLineEnd {
		f.RefreshSoftWrapVisualBuffers()
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
		if f.SoftWrap {
			newACX, newACY := CalcNewACXY(
				f.VisualBufferMapped, f.bufferLine,
				f.bufferIndex, f.GetViewportWidth(), f.ViewportOffsetY,
			)

			indexCheck := CalcBufferIndexFromACXY(
				newACX, newACY,
				f.bufferLine, f.VisualBuffer, f.VisualBufferMapped,
				f.ViewportOffsetY,
			)

			if indexCheck != f.bufferIndex {
				newACX = f.GetViewportWidth()
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
			case CursorPositionChange, WindowResize:
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
