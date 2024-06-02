package fileeditor

import (
	"fmt"

	"github.com/Asiandayboy/CLITextEditor/util/ansi"
)

type actionFunc func(key byte)

func (f *FileEditor) actionDeleteText(key byte) {
	fmt.Println("delete text")
}

func (f *FileEditor) actionCursorLeft(key byte) {
	fmt.Println("cursor left")
}

func (f *FileEditor) actionCursorRight(key byte) {
	fmt.Println("cursor right")
}

func (f *FileEditor) actionCursorUp(key byte) {
	fmt.Println("cursor up")
}

func (f *FileEditor) actionCursorDown(key byte) {
	fmt.Println("cursor down")
}

func (f *FileEditor) actionNewLine(key byte) {
	fmt.Println("new line")
}

func (f *FileEditor) actionTyping(key byte) {
	if ansi.IsAlphaChar(key) {
		fmt.Println(string(key))
	} else { // typing a char that is not printable
		fmt.Println("typing")
	}
}
