package writer_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"repo-tutor/internal/file/writer"
)

func entries(t *testing.T, directory string) []string {
	t.Helper()
	found, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range found {
		names = append(names, entry.Name())
	}
	return names
}

func TestWriteStoresTheBytesInTheNamedFile(t *testing.T) {
	directory := t.TempDir()
	if err := writer.Write(directory, "target.txt", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "target.txt"))
	if err != nil || string(data) != "hello" {
		t.Fatalf("file holds %q, %v", data, err)
	}
}

func TestWriteReplacesAnExistingFileWhole(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "target.txt")
	if err := os.WriteFile(path, []byte("a much longer earlier text"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writer.Write(directory, "target.txt", []byte("new")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "new" {
		t.Fatalf("file holds %q, %v", data, err)
	}
}

func TestWriteLeavesNoTemporaryFileAfterSuccess(t *testing.T) {
	directory := t.TempDir()
	if err := writer.Write(directory, "target.txt", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	names := entries(t, directory)
	if len(names) != 1 || names[0] != "target.txt" {
		t.Fatalf("directory holds %v", names)
	}
}

func TestWriteInAMissingDirectoryFailsAndLeavesNothingBehind(t *testing.T) {
	parent := t.TempDir()
	if err := writer.Write(filepath.Join(parent, "missing"), "target.txt", []byte("hello")); err == nil {
		t.Fatal("write into a missing directory succeeded")
	}
	if names := entries(t, parent); len(names) != 0 {
		t.Fatalf("parent holds %v", names)
	}
}

func TestWriteKeepsTheOldFileAndLeavesNoTemporaryFileWhenTheRenameFails(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "inside.txt"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writer.Write(directory, "target", []byte("new")); err == nil {
		t.Fatal("rename over a non-empty directory succeeded")
	}
	data, err := os.ReadFile(filepath.Join(target, "inside.txt"))
	if err != nil || string(data) != "old" {
		t.Fatalf("old content holds %q, %v", data, err)
	}
	if names := entries(t, directory); len(names) != 1 || names[0] != "target" {
		t.Fatalf("directory holds %v", names)
	}
}

func TestWriteGivesTheSameBytesForTheSameInput(t *testing.T) {
	data := []byte("same input")
	first, second := t.TempDir(), t.TempDir()
	if err := writer.Write(first, "target.txt", data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Write(second, "target.txt", data); err != nil {
		t.Fatal(err)
	}
	one, _ := os.ReadFile(filepath.Join(first, "target.txt"))
	two, _ := os.ReadFile(filepath.Join(second, "target.txt"))
	if !bytes.Equal(one, two) || !bytes.Equal(one, data) {
		t.Fatalf("files differ: %q %q", one, two)
	}
}
