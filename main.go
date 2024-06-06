package main

import (
	"fmt"
	"os"

	fileeditor "github.com/Asiandayboy/CLITextEditor/fileEditor"
	"github.com/Asiandayboy/CLITextEditor/table"
	"github.com/Asiandayboy/CLITextEditor/util/ansi"
	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

func enableVirtualTerminalInput() error {
	handle := windows.Handle(os.Stdin.Fd())

	var mode uint32
	err := windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return err
	}

	mode |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT
	err = windows.SetConsoleMode(handle, mode)
	if err != nil {
		return err
	}

	return nil
}

func newTable() table.Table {
	t := table.NewTable()
	// t.InnerBordersEnabled = false
	t.AddTitle("My ADC Champ Pool")
	t.AddRow([]string{"Champion Name", "Difficulty"})
	t.AddRow([]string{"Nilah", "Medium"})
	t.AddRow([]string{"Aphelios", "Hard"})
	t.AddRow([]string{"Varus", "Easy"})
	t.AddRow([]string{"Lucian", "Medium"})
	t.AddRow([]string{"Tristana", "Easy"})

	t.SetHeaderColorsRGB([]ansi.RGBColor{
		ansi.NewRGBColor(39, 183, 133),
		ansi.NewRGBColor(166, 39, 183),
	})

	t.SetSelectedColor(ansi.NewRGBColor(255, 255, 0))
	t.SetHeaderFontStyle("bold")
	t.SetHorizontalCellPadding(0, 0)
	t.SetHorizontalTextPadding(1, 1)
	t.Selected = true

	// buffer := make([]byte, 3)
	// _, err := os.Stdin.Read(buffer)
	// if err != nil {
	// 	panic(err)
	// }

	// if buffer[0] == 'q' {
	// 	return
	// } else if buffer[0] == 127 {
	// 	t.DeleteRow()
	// }

	// if buffer[0] == '\x1b' && buffer[1] == '[' {

	// 	key := buffer[2]

	// 	switch key {
	// 	case 'A':
	// 		t.HighlightPrevRow()
	// 	case 'B':
	// 		t.HighlightNextRow()
	// 	case 'C':
	// 		t.EditRowInput()
	// 	case 'D':
	// 		t.AcceptRowInput()
	// 	}
	// }

	// t.PrintTable(0, 0)

	return t
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("\033[31m[Error]: A filename must be provided as an argument\n\033[0m")
		return
	}

	var filename string = os.Args[1]

	editor := fileeditor.NewFileEditor(filename)
	editor.OpenFile()
	if err := editor.ReadFileToBuffer(); err != nil {
		fmt.Println(err)
		return
	}

	// set terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	ansi.ClearEntireScreen()
	ansi.SetTerminalWindowTitle(editor.Filename)

	enableVirtualTerminalInput()

	ansi.EnableMouseReporting()
	ansi.EnableAlternateScreenBuffer()
	ansi.EnableBlinkingLineCursor()

	editor.Render()

	// render loop
	go func() {
		editor.UpdateLoop()
	}()

	// input loop
	for {
		quit := editor.OnInput()
		if quit == 1 {
			break
		}
	}

	ansi.DisableMouseReporting()
	ansi.DisableAlternateScreenBuffer()
	ansi.ClearEntireScreen()
}
