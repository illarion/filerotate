package filerotate

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func TestNewWriter(t *testing.T) {
	// tmp dir:
	basePath, err := os.MkdirTemp("", "filerotate-test-*")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %v", err)
	}

	w, err := NewWriter(Options{
		FilePath: path.Join(basePath, "test.log"),
		Rotate:   5,
		Size:     1000,
		Mode:     0644,
	})

	if err != nil {
		t.Fatalf("failed to create a new writer: %v", err)
	}

	if w == nil {
		t.Fatalf("writer is nil")
	}

	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Fatalf("failed to write to the file: %v", err)
	}

	if n != 4 {
		t.Fatalf("invalid number of bytes written: %d", n)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("failed to close the writer: %v", err)
	}

}

func TestWriterRotates(t *testing.T) {
	// tmp dir:
	basePath, err := os.MkdirTemp("", "filerotate-test-*")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %v", err)
	}

	w, err := NewWriter(Options{
		FilePath: path.Join(basePath, "test.log"),
		Rotate:   5,
		Size:     1000,
		Mode:     0644,
	})

	if err != nil {
		t.Fatalf("failed to create a new writer: %v", err)
	}

	// write (w.options.Rotate * 2) * w.options.Size bytes
	c := int64(w.options.Rotate*2) * w.options.Size
	for c > 0 {
		n, err := w.Write([]byte("test"))
		if err != nil {
			t.Fatalf("failed to write to the file: %v", err)
		}
		c -= int64(n)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("failed to close the writer: %v", err)
	}

	// check the number of files
	files, err := os.ReadDir(basePath)

	if err != nil {
		t.Fatalf("failed to read the directory: %v", err)
	}

	if len(files) != w.options.Rotate+1 {
		t.Fatalf("invalid number of files: %d", len(files))
	}

	// check the names of the files
	set := make(map[string]struct{})
	for _, file := range files {
		set[file.Name()] = struct{}{}
	}

	if _, ok := set["test.log"]; !ok {
		t.Fatalf("file test.log is missing")
	}

	for i := 0; i < w.options.Rotate; i++ {
		name := fmt.Sprintf("test.log.%d", i+1)
		if _, ok := set[name]; !ok {
			t.Fatalf("file %s is missing", name)
		}
	}

}

func TestWriterRotatesOnSeparator(t *testing.T) {
	// tmp dir:
	basePath, err := os.MkdirTemp("", "filerotate-test-*")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %v", err)
	}

	w, err := NewWriter(Options{
		FilePath:      path.Join(basePath, "test.log"),
		Rotate:        5,
		Size:          1000,
		Mode:          0644,
		LineSeparator: LineSeparatorUnix,
	})

	testLine := "12345678901\n"

	if err != nil {
		t.Fatalf("failed to create a new writer: %v", err)
	}

	// write 2 * w.options.Size bytes
	c := int64(2) * w.options.Size

	for c > 0 {
		n, err := w.Write([]byte(testLine))
		if err != nil {
			t.Fatalf("failed to write to the file: %v", err)
		}
		c -= int64(n)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("failed to close the writer: %v", err)
	}

	// check content of the test.log.1 file - it should end with a whole testLine

	f, err := os.Open(path.Join(basePath, "test.log.1"))
	if err != nil {
		t.Fatalf("failed to open the file: %v", err)
	}

	defer f.Close()

	buf := make([]byte, len(testLine))

	info, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to get the file info: %v", err)
	}

	_, err = f.ReadAt(buf, info.Size()-int64(len(testLine)))
	if err != nil {
		t.Fatalf("failed to read the file: %v", err)
	}

	if string(buf) != testLine {
		t.Fatalf("invalid content of the file: %s", buf)
	}

}
