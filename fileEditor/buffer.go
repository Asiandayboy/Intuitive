package fileeditor

func (f *FileEditor) RefreshVisualBuffers() {
	f.VisualBuffer = []string{}
	f.VisualBufferMapped = []int{}

	var end int = 1

	for _, line := range f.FileBuffer {
		if len(line) >= f.EditorWidth {
			wordWrappedLines := f.GetWordWrappedLines(line, f.EditorWidth)

			end += len(wordWrappedLines) - 1

			f.VisualBuffer = append(f.VisualBuffer, wordWrappedLines...)
			f.VisualBufferMapped = append(f.VisualBufferMapped, end)
		} else {
			f.VisualBuffer = append(f.VisualBuffer, line)
			f.VisualBufferMapped = append(f.VisualBufferMapped, end)
		}
		end++
	}
}
