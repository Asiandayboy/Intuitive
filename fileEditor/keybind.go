package fileeditor

// Definitons for action names that can be rebinded
const (
	ActionHighlightText  string = "HighlightText"
	ActionCopyHighlight  string = "CopyHighlight"
	ActionMoveHighlight  string = "MoveHighlight"
	ActionPasteText      string = "PasteText"
	ActionDeleteText     string = "DeleteText"
	ActionToggleTextWrap string = "ToggleTextWrap"
	ActionScrollUp       string = "ScrollUp"
	ActionScrollDown     string = "ScrollDown"
	ActionScrollLeft     string = "ScrollLeft"
	ActionScrollRight    string = "ScrollRight"
	ActionToggleFileTree string = "ToggleFileTree"
)

const (
	CtrlZ        byte = 0
	CtrlA        byte = 1
	CtrlB        byte = 2
	CtrlC        byte = 3
	CtrlD        byte = 4
	CtrlE        byte = 5
	CtrlH        byte = 8
	Tab          byte = 9
	CtrlL        byte = 12
	NewLine      byte = 13
	CtrlN        byte = 14
	CtrlO        byte = 15
	CtrlR        byte = 18
	CtrlS        byte = 19
	CtrlT        byte = 20
	CtrlV        byte = 22
	CtrlX        byte = 24
	CtrlY        byte = 25
	Space        byte = 32
	LBracket     byte = 91
	RBracket     byte = 93
	Escape       byte = 0x1b
	ForwardSlash byte = 47
	Backspace    byte = 127

	UpArrowKey    byte = 'A' // ASCII decimal -> 65
	DownArrowKey  byte = 'B' // ASCII decimal -> 66
	RightArrowKey byte = 'C' // ASCII decimal -> 67
	LeftArrowKey  byte = 'D' // ASCII decimal -> 68
)

// Represents the user's keybindings for each action
type Keybind struct {
	HighlightText  byte
	CopyHighlight  byte
	MoveHighlight  byte
	PasteText      byte
	ToggleTextWrap byte
	ScrollUp       byte
	ScrollDown     byte
	ScrollLeft     byte
	ScrollRight    byte
	ToggleFileTree byte

	// these keybinds cannot be changed
	cursorLeft  byte
	cursorRight byte
	cursorUp    byte
	cursorDown  byte
}

/*
Returns a default keybind
*/
func NewKeybind() Keybind {
	return Keybind{
		cursorUp:    'A',
		cursorDown:  'B',
		cursorRight: 'C',
		cursorLeft:  'D',
	}
}

/*
Executes the action binded to the key based on the user's
keybindings
*/
func (k Keybind) MapKeybindToAction(key byte, isArrowKey bool, editor *FileEditor) {
	keybindings := map[byte]actionFunc{
		k.cursorLeft:  editor.actionCursorLeft,
		k.cursorRight: editor.actionCursorRight,
		k.cursorUp:    editor.actionCursorUp,
		k.cursorDown:  editor.actionCursorDown,
	}

	action, exists := keybindings[key]
	if exists {
		if key == UpArrowKey || key == DownArrowKey || key == RightArrowKey || key == LeftArrowKey {
			if isArrowKey { // typing arrow keys, which are also A, B, C or D
				action()
			}
		}
	}
}

func (k *Keybind) ChangeKeybind(action string, keybind byte) {
	switch action {
	case ActionHighlightText:
		k.HighlightText = keybind
	case ActionCopyHighlight:
		k.CopyHighlight = keybind
	case ActionMoveHighlight:
		k.MoveHighlight = keybind
	case ActionPasteText:
		k.PasteText = keybind
	case ActionToggleTextWrap:
		k.ToggleTextWrap = keybind
	case ActionScrollUp:
		k.ScrollUp = keybind
	case ActionScrollDown:
		k.ScrollDown = keybind
	case ActionScrollLeft:
		k.ScrollLeft = keybind
	case ActionScrollRight:
		k.ScrollRight = keybind
	case ActionToggleFileTree:
		k.ToggleFileTree = keybind
	default:
		return
	}
}
