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
	EnumQuit byte = iota + 1
	EnumKeyboardInput
	EnumWindowResize
	EnumEditorModeChange
	EnumCursorPositionChange
	EnumNewLineInserted
	EnumNewLineInsertedAtLineEnd
	EnumSoftWrapEnabled
	EnumSoftWrapDisabled
	EnumToggleCommandBar
)

const (
	IndentWithSpace uint8 = 1
	IndentWithTab   uint8 = 2
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
	Saved           bool // refers to whether the file has been saved since the last modification
	EditorMode      byte
	Keybindings     Keybind
	file            *os.File
	Filename        string
	inputChan       chan byte
	QuitProgramFlag bool

	CommandBarBuffer  string
	CommandBarCursorX int

	FileBuffer         []string   // contains each line of the actual file
	VisualBuffer       []string   // contains word wrapped lines; this is what gets rendered to the screen
	VisualBufferMapped []int      // contains the ending index (1-indexed) of word-wrapped lines
	apparentCursorX    int        // cursor's X position
	apparentCursorY    int        // cursor's Y position
	bufferLine         int        // refers to current line of FileBuffer; used when editing FileBuffer
	bufferIndex        int        // refers to current index of current line of FileBuffer; used when editing FileBuffer
	TermWidth          int        // width of the terminal window
	TermHeight         int        // height of the terminal window
	StatusBarHeight    int        // height of the status bar
	ViewportOffsetX    int        // used for horizontal scrolling
	ViewportOffsetY    int        // used for vertical scrolling
	TabMap             TabMapType // stores the start and end indicies of each tab character; used only when TabIndentType = IndentWithTab
	CommandBarToggled  bool

	// Configs
	SoftWrapEnabled bool
	PrintEmptyLines bool  // print tildes for empty lines
	TabIndentType   uint8 // determines how tabs are stored in the FileBuffer (either as ASCII 9 or ASCII 32)
	TabSize         uint8

	// debugging
	actualBufferIndex int
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
		Saved:              true,
		EditorMode:         EditorCommandMode,
		Keybindings:        NewKeybind(),
		inputChan:          make(chan byte, 1),
		CommandBarToggled:  false,
		QuitProgramFlag:    false,

		SoftWrapEnabled: true,
		PrintEmptyLines: false,
		TabIndentType:   IndentWithTab,
		TabSize:         4,
	}
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
	if f.SoftWrapEnabled {
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
	var savedText string = " (Saved)"
	if !f.Saved {
		savedText = " (Unsaved)"
	}

	if !f.Saved {
		fmt.Print(Green + f.Filename + Red + Italic + savedText + Reset)
	} else {
		fmt.Print(Green + f.Filename + Blue + Italic + savedText + Reset)
	}

	// draw buffer indicies position + 1
	ansi.MoveCursor(yOffset+2, f.TermWidth-8)
	fmt.Printf(modeColors[f.EditorMode].ToFgColorANSI()+"%d:%d"+Reset, f.bufferLine, f.bufferIndex)

	// // draw word count
	// wordCountColor := ansi.NewRGBColor(120, 120, 120).ToFgColorANSI()
	// ansi.MoveCursor(yOffset+2, f.TermWidth-18)
	// fmt.Print(wordCountColor+"Char count: ", f.GetBufferCharCount(), Reset)

	// debugging purposes
	ansi.MoveCursor(yOffset+2, f.TermWidth-50)
	fmt.Print("EditorWidth: ", f.GetViewportWidth())
	ansi.MoveCursor(yOffset+2, f.TermWidth-30)
	fmt.Print(Grey+"abi: ", f.GetViewportHeight(), Reset)
	fmt.Print(Grey+"te: ", f.apparentCursorY, Reset)
	// fmt.Print(Grey+"Tab Size: ", f.TabSize, Reset)
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
	if f.SoftWrapEnabled && flag != EnumNewLineInsertedAtLineEnd {
		f.RefreshSoftWrapVisualBuffers()
	} else if !f.SoftWrapEnabled && flag != EnumNewLineInsertedAtLineEnd {
		f.RefreshNoWrapVisualBuffers()
	}

	if flag == EnumCursorPositionChange ||
		flag == EnumKeyboardInput ||
		flag == EnumNewLineInserted ||
		flag == EnumNewLineInsertedAtLineEnd {

		f.UpdateBufferIndicies()
	}

	if f.EditorMode == EditorEditMode && (flag == EnumKeyboardInput ||
		flag == EnumNewLineInserted ||
		flag == EnumNewLineInsertedAtLineEnd) {

		f.Saved = false
	}

	switch flag {
	case EnumWindowResize, EnumSoftWrapEnabled:
		if f.SoftWrapEnabled {
			f.CalculateNewCursorPos()
		}
	case EnumToggleCommandBar:
		f.ToggleCommandBar(!f.CommandBarToggled)
	}

	f.PrintBuffer()
	f.PrintStatusBar()

	ansi.MoveCursor(f.apparentCursorY, f.apparentCursorX)

	if f.CommandBarToggled {
		f.UpdateCommandBarState()
	}

	// FIXED: 9/25 -> hide cursor when cursorY exceeds viewport height
	if f.apparentCursorY > f.GetViewportHeight() {
		ansi.HideCursor()
	} else {
		ansi.ShowCursor()
	}
}

func (editor *FileEditor) RenderLoop() int {
	termW, termH, _ := term.GetSize(int(os.Stdout.Fd()))

	if !(termW == editor.TermWidth && termH == editor.TermHeight) { // resize
		editor.inputChan <- EnumWindowResize
	}

	select {
	case inputCode := <-editor.inputChan:
		switch inputCode {
		case EnumQuit:
			return 1
		case EnumKeyboardInput, EnumEditorModeChange, EnumCursorPositionChange, EnumToggleCommandBar,
			EnumNewLineInserted, EnumNewLineInsertedAtLineEnd, EnumSoftWrapDisabled, EnumSoftWrapEnabled:
			editor.Render(inputCode)
		case EnumWindowResize:
			ansi.ClearEntireScreen()
			editor.TermHeight = termH
			editor.TermWidth = termW
			editor.Render(EnumWindowResize)
		}
	default:
	}

	return 0
}

func (editor *FileEditor) InputLoop() int {
	var buf [16]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		fmt.Println("Error reading from stdin:", err)
		return 1
	}

	if buf[0] == Escape { // mouse input and arrow keys, etc.
		isMouseInput, mouseEvent := ReadEscSequence(buf[:], n)

		if isMouseInput && !editor.CommandBarToggled {
			ret := HandleMouseInput(editor, mouseEvent)

			switch ret {
			case EnumCursorPositionChange, EnumWindowResize, EnumSoftWrapDisabled, EnumSoftWrapEnabled:
				editor.inputChan <- ret
			}
		} else {
			ret := HandleEscapeInput(editor, buf[:], n)

			switch ret {
			case EnumCursorPositionChange:
				editor.inputChan <- EnumCursorPositionChange
			case EnumEditorModeChange:
				editor.inputChan <- EnumEditorModeChange
			}
		}
	} else {
		ret := HandleKeyboardInput(editor, buf[0])

		switch ret {
		case EnumQuit:
			editor.inputChan <- EnumQuit
			return 1
		default:
			editor.inputChan <- ret
		}
	}

	return 0
}
