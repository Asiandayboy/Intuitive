package fileeditor

func (f *FileEditor) GetViewportHeight() int {
	return f.TermHeight - f.StatusBarHeight
}

func (f *FileEditor) GetViewportWidth() int {
	return f.TermWidth - EditorLeftMargin + 1
}
