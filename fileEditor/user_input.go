package fileeditor

import (
	"strconv"
	"strings"

	"github.com/Asiandayboy/CLITextEditor/util/ansi"
)

/*
This file is responsible for reading and parsing the stdin
and delegating each input to the appropriate function.

It reads for mouse input, regular keyboard input, and other
escape sequences
*/

/* SGR mouse reporting mode

With SGR, mouse events are encoded in a more detailed
and structured way, which will avoid capping our mouse position
at 127, 98 whenever we try to read from it.

This means, we must enable \x1b[?1003h (to enable tracking any mouse event)
and \x1b[?1006h (to use SGR encoding when reporting mouse event)

With SGR, the mouse code sequence is always prefixed with 91 and 60,
aka "[<"", which is always a length of 2.
This is how we know if the input is a mouse event

Ex: The buffer is an array of bytes, represented by the ASCII codes

Example buffer = [91, 60, 51, 53, 59, 54, 52, 59, 51, 51, 109]

Using the ASCII chart, we can decode this
	91 = [
	60 = <
	51 = 3
	53 = 5
	59 = ;

	54 = 6
	52 = 4
	59 = ;

	51 = 3
	51 = 3
	109 = m

	And we get: [<35;64;33m
	This means our mouse is at position x = 64, y =33,
	with no buttons currently pressed

	Casting our buffer into a string will give us the decoded string,
	and from it we can parse it for the event type and mouse position




The first parameter indicates the type of mouse event it is.
Here are all the mouse events:
	0 = Left click
	1 = Mouse wheel click
	2 = Right click

	32 = Left drag
	33 = Mouse wheel drag
	34 = Right drag

	64 = Scroll up
	65 = Scroll down

*/

const SGR_MOUSE_PREFIX string = "[<"

const (
	MouseEventLeftClick byte = iota
	MouseEventWheelClick
	MouseEventRightClick
	MouseEventLeftDrag
	MouseEventWheelDrag
	MouseEventRightDrag
	MouseEventScrollUp
	MouseEventScrollDown
)

type MouseInput struct {
	Event byte
	X, Y  int
}

/*
Reads the buffer and returns true if the escape sequence is a mouse
input, in which case a MouseInput is returned as well. Otherwise, the
function will return false, and an empty MouseInput{}
*/
func ReadEscSequence(buf []byte, numBytesReadFromBuf int) (bool, MouseInput) {
	var seq string = string(buf[1:numBytesReadFromBuf])

	if strings.HasPrefix(seq, SGR_MOUSE_PREFIX) {
		// ending at len(seq) - 1 to ignore the 'm' at the end
		parts := strings.Split(seq[2:len(seq)-1], ";")

		event, _ := strconv.Atoi(parts[0])
		x, _ := strconv.Atoi(parts[1])
		y, _ := strconv.Atoi(parts[2])

		return true, MouseInput{byte(event), x, y}
	}

	return false, MouseInput{}
}

// using a debounce to ignore release event
var lastMouseInputEvent byte = 100

func HandleMouseInput(editor *FileEditor, m MouseInput) byte {
	if m.Event == MouseEventLeftClick && m.Event != lastMouseInputEvent {
		lastMouseInputEvent = m.Event
		return editor.SetCursorPosition(m)
	}

	if lastMouseInputEvent == m.Event {
		lastMouseInputEvent = 100
	}

	return 0
}

func HandleEscapeInput(editor *FileEditor, buf []byte, n int) byte {
	if n == 3 && buf[2] == Up || buf[2] == Down || buf[2] == Right || buf[2] == Left {
		editor.Keybindings.MapKeybindToAction(buf[2], true, editor)
		return CursorPositionChange
	}

	// Return to Command mode
	if n == 1 && buf[0] == Escape {
		editor.EditorMode = EditorCommandMode
		return EditorModeChange
	}

	return 0
}

func HandleKeyboardInput(editor *FileEditor, key byte) byte {
	if key == 'q' {
		return Quit
	}

	const asciiLowerDif uint8 = 32

	if ansi.IsAlphaChar(key) {
		// Transition to View or Edit mode
		if editor.EditorMode == EditorCommandMode {
			if key == EditorEditMode || key == EditorEditMode+asciiLowerDif {
				editor.EditorMode = EditorEditMode
				return EditorModeChange
			} else if key == EditorViewMode || key == EditorViewMode+asciiLowerDif {
				editor.EditorMode = EditorViewMode
				return EditorModeChange
			}
		}

		if editor.EditorMode == EditorEditMode {
			editor.actionTyping(key)
		}

	} else {
		if editor.EditorMode == EditorEditMode {
			if key == NewLine {
				editor.actionNewLine(key)
			} else if key == Backspace {
				editor.actionDeleteText(key)
			}
		}
	}

	// editor.Keybindings.MapKeybindToAction(key, false, *editor)
	return KeyboardInput
}
