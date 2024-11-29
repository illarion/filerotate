package filerotate

import (
	"fmt"
	"os"
	"sync"
)

// Options for the file rotation
type Options struct {
	// FilePath full path to the log file (i.e. our.log)
	FilePath string
	// Rotate log file count times before removing. If Rotate count is 0, old versions are removed rather than rotated, so that only our.log is present
	Rotate int
	// Size of the file to grow. When exceeded, file is rotated.
	Size int64
	// File mode, like 0600
	Mode os.FileMode
}

var DefaultOptions = Options{
	Rotate: 5,
	Size:   10 * 1024 * 1024, // 10MB
	Mode:   0644,
}

type Writer struct {
	options Options
	mu      sync.Mutex
	f       *os.File // current file
}

// NewWriter creates a new Writer
func NewWriter(options Options) (*Writer, error) {

	if options.FilePath == "" {
		return nil, fmt.Errorf("file path is empty")
	}

	if options.Mode == 0 {
		options.Mode = DefaultOptions.Mode
	}

	if options.Rotate == 0 {
		options.Rotate = DefaultOptions.Rotate
	}

	if options.Size == 0 {
		options.Size = DefaultOptions.Size
	}

	f, err := os.OpenFile(options.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, options.Mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new file: %v", err)
	}

	return &Writer{
		options: options,
		f:       f,
	}, nil
}

// Write writes the data to the file. If the file size exceeds the limit, it rotates the file.
func (w *Writer) Write(p []byte) (n int, err error) {

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.f == nil {
		return 0, fmt.Errorf("file is closed")
	}

	if w.options.Size > 0 {
		fi, err := w.f.Stat()
		if err != nil {
			return 0, err
		}
		if fi.Size() > w.options.Size {
			if err := w.rotate(); err != nil {
				return 0, err
			}
		}
	}

	return w.f.Write(p)
}

func (w *Writer) rotate() error {

	if w.f != nil {
		err := w.f.Close()
		if err != nil {
			return fmt.Errorf("failed to close the file: %v", err)
		}
	}

	// file named filePath.N where N is Rotate - is removed
	// file named filePath.N-1 is renamed to filePath.N
	// ...
	// file named filePath is renamed to filePath.1

	// remove the last file
	removePath := fmt.Sprintf("%s.%d", w.options.FilePath, w.options.Rotate)
	if _, err := os.Stat(removePath); err == nil {
		err = os.Remove(removePath)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %v", removePath, err)
		}
	}

	for i := w.options.Rotate - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.options.FilePath, i)

		if _, err := os.Stat(oldPath); err != nil {
			// file does not exist, skip
			continue
		}

		newPath := fmt.Sprintf("%s.%d", w.options.FilePath, i+1)
		err := os.Rename(oldPath, newPath)
		if err != nil {
			return fmt.Errorf("failed to rename %s to %s: %v", oldPath, newPath, err)
		}
	}

	// rename the current file
	err := os.Rename(w.options.FilePath, w.options.FilePath+".1")
	if err != nil {
		return fmt.Errorf("failed to rename %s to %s: %v", w.options.FilePath, w.options.FilePath+".1", err)
	}

	// create a new file
	f, err := os.OpenFile(w.options.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, w.options.Mode)
	if err != nil {
		return fmt.Errorf("failed to create a new file: %v", err)
	}

	w.f = f
	return nil

}

// Close closes the file
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.f != nil {
		err := w.f.Close()
		w.f = nil
		return err
	}

	return nil
}
