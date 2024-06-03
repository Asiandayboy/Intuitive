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
)

// the three editor modes a user can be in
const (
	EditorCommandMode uint8 = 'C'
	EditorEditMode    uint8 = 'E'
	EditorViewMode    uint8 = 'V'
)

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
	Filename string
	Buffer   []string
	file     *os.File

	CursorX         int
	CursorY         int
	ViewportOffsetX int
	ViewportOffsetY int
	TermWidth       int
	TermHeight      int
	MaxTermHeight   int
	StatusBarHeight int
	Saved           bool
	EditorMode      byte

	Keybindings Keybind

	inputChan chan byte
}

func NewFileEditor(filename string) FileEditor {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}

	return FileEditor{
		Filename:        filename,
		Buffer:          make([]string, 0),
		CursorX:         0,
		CursorY:         0,
		ViewportOffsetX: 7,
		ViewportOffsetY: 0,
		TermWidth:       width,
		TermHeight:      height,
		MaxTermHeight:   height,
		StatusBarHeight: 3,
		Saved:           false,
		EditorMode:      EditorCommandMode,

		Keybindings: NewKeybind(),
		inputChan:   make(chan byte, 1),
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
	for _, line := range f.Buffer {
		count += len(line)
	}

	return count
}

func (f *FileEditor) ReadFileToBuffer() error {
	scanner := bufio.NewScanner(f.file)
	var rowCount int = 0
	for scanner.Scan() {
		f.Buffer = append(f.Buffer, scanner.Text())
		rowCount++
	}

	if rowCount == 0 {
		f.Buffer = append(f.Buffer, "")
	}

	// // set cursor's initial position to end of file
	// if rowCount > 0 {
	// 	f.CursorX = len(f.Buffer[rowCount-1])
	// 	f.CursorY = rowCount - 1
	// }

	return scanner.Err()
}

func (f *FileEditor) GetWordWrappedLines(
	line string,
	maxWidth int,
	maxHeight *int,
) (lines []string) {

	length := len(line)

	for length >= maxWidth {
		dif := maxWidth - length - 1
		cutoffIndex := length + dif
		lines = append(lines, line[:cutoffIndex])
		line = line[cutoffIndex:]
		length = len(line)
		*maxHeight--
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

	/*
		The status bar height is subtracted from the terminal height to avoid the unnecessary scrolling,
		which would truncate the beginning of the buffer; And it also refers to the height
		that the status bar will takeup. The maxHeight value will also be decreased the more
		word-wrapped lines there are.
	*/
	var maxHeight int = f.TermHeight - f.StatusBarHeight

	var maxWidth int = f.TermWidth - EditorLeftMargin
	var cursorX, cursorY int

	ansi.MoveCursor(0, 0)

	for i := 0; i < maxHeight; i++ {
		if i >= len(f.Buffer) {
			fmt.Printf("%s   ~ %s%s%s\n", lineNumColor, borderColor, Vertical, Reset)
		} else {
			line := f.Buffer[i]

			if len(line) >= maxWidth {
				wordWrappedLines := f.GetWordWrappedLines(line, maxWidth, &maxHeight)

				for j, l := range wordWrappedLines {
					if j == 0 {
						fmt.Printf("%s%4d %s%s%s %s\n", lineNumColor, i+1, borderColor, Vertical, Reset, l)
						cursorX = len(line) + EditorLeftMargin
						cursorY += 1
					} else {
						fmt.Printf("%s     %s%s%s %s\n", lineNumColor, borderColor, Vertical, Reset, l)
						cursorX = len(line) + EditorLeftMargin
						cursorY += 1
					}
				}
			} else {
				fmt.Printf("%s%4d %s%s%s %s\n", lineNumColor, i+1, borderColor, Vertical, Reset, line)
				cursorX = len(line) + EditorLeftMargin
				cursorY += 1
			}

		}
	}

	ansi.MoveCursor(cursorY, cursorX)
}

func (f FileEditor) PrintStatusBar() {
	width := f.TermWidth - EditorLeftMargin + 3
	height := f.StatusBarHeight

	xOffset := EditorLeftMargin - 3
	yOffset := f.TermHeight - height

	borderColor := ansi.NewRGBColor(60, 60, 60)

	modeColors := map[byte]ansi.RGBColor{
		EditorCommandMode: ansi.NewRGBColor(75, 176, 255), // blue
		EditorEditMode:    ansi.NewRGBColor(216, 148, 53), // orange
		EditorViewMode:    ansi.NewRGBColor(158, 75, 253), // purple
	}

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

	wordCountColor := ansi.NewRGBColor(100, 100, 100).ToFgColorANSI()

	// draw word count
	ansi.MoveCursor(yOffset+2, f.TermWidth-18)
	fmt.Print(wordCountColor+"Char count: ", f.GetBufferCharCount(), Reset)
}

func (f FileEditor) Render() {
	f.PrintBuffer()
	f.PrintStatusBar()
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
				editor.Render()
			case EditorModeChange:
				editor.Render()
			case Resize:
				ansi.ClearEntireScreen()
				editor.TermHeight = termH
				editor.TermWidth = termW

				editor.Render()
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
			HandleMouseInput(editor, mouseEvent)
		} else {
			ret := HandleEscapeInput(editor, buf[:], n)

			if ret == EditorModeChange {
				editor.inputChan <- EditorModeChange
			}
		}
	} else {
		ret := HandleKeyboardInput(editor, buf[0])

		switch ret {
		case Quit:
			editor.inputChan <- ret
			return 1
		case EditorModeChange:
			editor.inputChan <- ret
		case 0:
			editor.inputChan <- ret
		}

	}

	return 0
}

func drawBox(width uint16, height uint16, words [][]string) {
	xOffset := "     "

	// draw top border
	fmt.Print(xOffset)
	fmt.Print(Grey + TopLCorner + Reset)
	var i uint16
	for i = 0; i < width; i++ {
		fmt.Print(Grey + Horizontal + Reset)
		if i+1 == width {
			fmt.Println(Grey + TopRCorner + Reset)
		}
	}

	// draw side borders and words
	for i = 0; i < height; i++ {
		fmt.Print(xOffset)
		fmt.Print(Grey + Vertical + Reset)
		var totalFormatedWidth uint16 = width

		firstWords := words[i]
		for _, word := range firstWords {
			wordLen := uint16(len(word))
			if wordLen >= 50 {
				panic(fmt.Sprintf(`"%s" is bigger than the specified box width.`, word))
			}

			if i == 3 { // row containing filename
				fmt.Printf(" Filename: %s%s %s%s(Unsaved)%s", Green, word, Italic, Cyan, Reset)
				wordLen += 21
			} else {
				fmt.Print(word)
			}
			totalFormatedWidth -= wordLen
		}

		formatted := fmt.Sprintf("%%%ds", totalFormatedWidth)
		fmt.Printf("%s%s\n", fmt.Sprintf(formatted, " "), Grey+Vertical+Reset)
	}

	// draw bottom border
	fmt.Print(xOffset)
	fmt.Print(Grey + BotLCorner + Reset)
	for i = 0; i < width; i++ {
		fmt.Print(Grey + Horizontal + Reset)
		if i+1 == width {
			fmt.Println(Grey + BotRCorner + Reset)
		}
	}
}

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
