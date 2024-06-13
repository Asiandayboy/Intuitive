package fileeditor

func (f *FileEditor) GetWordWrappedLines(line string, maxWidth int) (lines []string) {
	length := len(line)

	for length >= maxWidth {
		dif := maxWidth - length - 1
		cutoffIndex := length + dif
		lines = append(lines, line[:cutoffIndex])
		line = line[cutoffIndex:]
		length = len(line)
	}

	/*
		when the loop breaks, we need to add the last remaining line,
		which will be less than maxWidth
	*/
	lines = append(lines, line)

	return lines
}

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
