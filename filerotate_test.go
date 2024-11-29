package filerotate

import (
	"fmt"
	"os"
	"testing"
)

func TestNewWriter(t *testing.T) {
	// tmp dir:
	basePath, err := os.MkdirTemp("", "filerotate-test-*")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %v", err)
	}

	w, err := NewWriter(Options{
		FilePath: basePath + "/test.log",
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

func TestWriter_Rotates(t *testing.T) {
	// tmp dir:
	basePath, err := os.MkdirTemp("", "filerotate-test-*")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %v", err)
	}

	w, err := NewWriter(Options{
		FilePath: basePath + "/test.log",
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

	// output file names for debug purposes:
	for k := range set {
		fmt.Println(k)
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
