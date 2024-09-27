package fileeditor

import (
	// "bufio"
	// "fmt"
	"os"
	"strings"
)

/*
Writes the current FileBuffer to the opened file
*/
func (f *FileEditor) SaveFile() {
	f.Saved = true

	data := strings.Join(f.FileBuffer, "\n")
	os.WriteFile(f.Filename, []byte(data), 0644)
}

/*
Opens the file or creates a new one if it cannot be found,
reads its content into the buffer
*/
func (f *FileEditor) OpenFile() {
	// file, err := os.OpenFile(f.Filename, os.O_WRONLY, 0644)
	file, err := os.Open(f.Filename)
	if err != nil {
		file, _ = os.Create(f.Filename)
	}

	f.file = file
}

func (f *FileEditor) CloseFile() error {
	err := f.file.Close()
	return err
}
