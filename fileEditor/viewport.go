package fileeditor

// The viewport height does not include the space occupied by the status bar
func (f *FileEditor) GetViewportHeight() int {
	return f.TermHeight - f.StatusBarHeight
}

// The viewport width does not include the margin space to the left
func (f *FileEditor) GetViewportWidth() int {
	return f.TermWidth - EditorLeftMargin + 1
}
