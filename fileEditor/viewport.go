package fileeditor

func (f *FileEditor) GetViewportHeight() int {
	return f.TermHeight - f.StatusBarHeight
}
