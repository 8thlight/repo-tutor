package repository

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const maximumFileBytes = 512 * 1024
const maximumSnapshotBytes = 64 * 1024 * 1024

func Read(ctx context.Context, directory string) (Snapshot, error) {
	source, err := readSource(ctx, directory)
	if err != nil {
		return Snapshot{}, err
	}
	return Build(source)
}

func readSource(ctx context.Context, directory string) (Source, error) {
	root, err := filepath.Abs(directory)
	if err != nil {
		return Source{}, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return Source{}, fmt.Errorf("open repository: %w", err)
	}
	information, err := os.Stat(root)
	if err != nil || !information.IsDir() {
		return Source{}, fmt.Errorf("repository must be a directory: %s", root)
	}
	var names []string
	var warnings []string
	err = filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkError error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkError != nil {
			return walkError
		}
		if entry.IsDir() {
			if name != root && excludedDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || !supportedFile(entry.Name()) {
			return nil
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		names = append(names, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return Source{}, fmt.Errorf("scan repository: %w", err)
	}
	if ignored, err := ignoredFiles(ctx, root, names); err == nil {
		var included []string
		for _, name := range names {
			if !ignored[name] {
				included = append(included, name)
			}
		}
		names = included
	}
	sort.Strings(names)
	var files []File
	total := 0
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return Source{}, err
		}
		file, err := os.Open(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return Source{}, fmt.Errorf("read %s: %w", name, err)
		}
		data, readError := io.ReadAll(io.LimitReader(file, maximumFileBytes+1))
		closeError := file.Close()
		if readError != nil {
			return Source{}, fmt.Errorf("read %s: %w", name, readError)
		}
		if closeError != nil {
			return Source{}, closeError
		}
		if len(data) > maximumFileBytes || !utf8.Valid(data) || strings.ContainsRune(string(data), 0) {
			warnings = append(warnings, name+": skipped oversized or non-text file")
			continue
		}
		total += len(data)
		if total > maximumSnapshotBytes {
			return Source{}, fmt.Errorf("repository source exceeds the 64 MiB scan limit; choose a smaller module directory")
		}
		files = append(files, File{name, string(data)})
	}
	changeLog, history := readChanges(ctx, root)
	if err := ctx.Err(); err != nil {
		return Source{}, err
	}
	return Source{Root: root, Files: files, Warnings: warnings, ChangeLog: changeLog, History: history}, nil
}

func excludedDirectory(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "vendor", "node_modules", "output", "dist", "build", "result", "testdata", "fixtures", "__pycache__", "coverage":
		return true
	}
	return false
}

func supportedFile(name string) bool {
	if strings.HasPrefix(name, ".") {
		return false
	}
	if name == "go.mod" || manifestFile(name) || sourceExtension(name) {
		return true
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".go", ".md":
		return true
	}
	return false
}

func ignoredFiles(ctx context.Context, root string, names []string) (map[string]bool, error) {
	command := exec.CommandContext(ctx, "git", "check-ignore", "--stdin", "-z")
	command.Dir = root
	command.Stdin = strings.NewReader(strings.Join(names, "\x00") + "\x00")
	output, err := command.Output()
	if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	result := map[string]bool{}
	for _, name := range strings.Split(string(output), "\x00") {
		result[name] = true
	}
	return result, nil
}
