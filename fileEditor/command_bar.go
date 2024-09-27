package fileeditor

import (
	"fmt"

	"github.com/Asiandayboy/CLITextEditor/render"
	"github.com/Asiandayboy/CLITextEditor/util/ansi"
)

const CMDBAR_SAVE string = "save"
const CMDBAR_SAVE_AS string = "save as"

const cmdBarWidth int = 35
const cmdBarHeight int = 3
const cmdBarLeftPadding int = 3
const cmdBarCursorXBoundary int = cmdBarWidth - 10

// 5 is the length of "CMD> "; be sure to change this is you're changing the prefix
const cmdBarPrefixLength int = 5

func drawCommandBar(f FileEditor) {
	xPos := f.GetViewportWidth()/2 - cmdBarWidth/2 + cmdBarPrefixLength
	yPos := f.GetViewportHeight()/2 - cmdBarHeight/2
	blueRGB := modeColors[f.EditorMode]
	blue := blueRGB.ToFgColorANSI()

	render.DrawBox(render.Box{
		Width: cmdBarWidth, Height: cmdBarHeight,
		X:           xPos,
		Y:           yPos,
		BorderStyle: "single",
		BorderColor: blueRGB,
	}, true)

	ansi.MoveCursor(yPos+2, xPos+cmdBarLeftPadding)
	fmt.Printf("%sCMD> %s%s", blue, f.CommandBarBuffer, Reset)

	if f.CommandBarCursorX >= 0 {
		ansi.MoveCursor(yPos+2, xPos+cmdBarLeftPadding)
		ansi.MoveCursorRight(cmdBarPrefixLength + f.CommandBarCursorX)
	}
}

func executeCommandBarStr(cmdString string) {
	switch cmdString {
	case CMDBAR_SAVE:
	case CMDBAR_SAVE_AS:
	default:
	}
}

func (f *FileEditor) UpdateCommandBarState() {
	drawCommandBar(*f)
}

func (f *FileEditor) ToggleCommandBar(toggled bool) {
	f.CommandBarToggled = toggled

	if toggled {
		drawCommandBar(*f)
	} else {
		executeCommandBarStr(f.CommandBarBuffer)
	}

	f.CommandBarBuffer = ""
	f.CommandBarCursorX = 0
}

/*
Inserts text from the current cursor's position in the command bar
*/
func (f *FileEditor) commandBarTyping(key byte) {
	if f.CommandBarCursorX > cmdBarCursorXBoundary || len(f.CommandBarBuffer) > cmdBarCursorXBoundary {
		return
	}

	if len(f.CommandBarBuffer) == 0 {
		f.CommandBarBuffer = string(key)
		f.CommandBarCursorX = 1
	} else {
		before := f.CommandBarBuffer[:f.CommandBarCursorX]
		after := f.CommandBarBuffer[f.CommandBarCursorX:]

		f.CommandBarBuffer = before + string(key) + after
		f.CommandBarCursorX++
	}

}

/*
Delete text from the current cursor's position in the command bar
*/
func (f *FileEditor) commandBarDeleteText() {
	if f.CommandBarCursorX > 0 && len(f.CommandBarBuffer) > 0 {
		before := f.CommandBarBuffer[:f.CommandBarCursorX-1]
		after := f.CommandBarBuffer[f.CommandBarCursorX:]

		f.CommandBarBuffer = before + after
		f.CommandBarCursorX--
	}
}

/*
Moves the cursor left or right in the command bar, depending on the arrow key
pressed; Does not move it up or down since the command bar is just one line.
*/
func (f *FileEditor) commandBarMoveCursor(key byte) {
	if key == LeftArrowKey {
		if f.CommandBarCursorX > 0 {
			ansi.MoveCursorLeft(1)
			f.CommandBarCursorX--
		}
	} else {
		if f.CommandBarCursorX < len(f.CommandBarBuffer) && f.CommandBarCursorX <= cmdBarCursorXBoundary {
			ansi.MoveCursorRight(1)
			f.CommandBarCursorX++
		}
	}
}
