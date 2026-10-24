package repository

import (
	"context"
	"crypto/sha256"
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
	root, err := filepath.Abs(directory)
	if err != nil {
		return Snapshot{}, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return Snapshot{}, fmt.Errorf("open repository: %w", err)
	}
	information, err := os.Stat(root)
	if err != nil || !information.IsDir() {
		return Snapshot{}, fmt.Errorf("repository must be a directory: %s", root)
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
		return Snapshot{}, fmt.Errorf("scan repository: %w", err)
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
	fingerprint := sha256.New()
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return Snapshot{}, err
		}
		file, err := os.Open(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return Snapshot{}, fmt.Errorf("read %s: %w", name, err)
		}
		data, readError := io.ReadAll(io.LimitReader(file, maximumFileBytes+1))
		closeError := file.Close()
		if readError != nil {
			return Snapshot{}, fmt.Errorf("read %s: %w", name, readError)
		}
		if closeError != nil {
			return Snapshot{}, closeError
		}
		if len(data) > maximumFileBytes || !utf8.Valid(data) || strings.ContainsRune(string(data), 0) {
			warnings = append(warnings, name+": skipped oversized or non-text file")
			continue
		}
		total += len(data)
		if total > maximumSnapshotBytes {
			return Snapshot{}, fmt.Errorf("repository source exceeds the 64 MiB scan limit; choose a smaller module directory")
		}
		files = append(files, File{name, string(data)})
		fmt.Fprintf(fingerprint, "%d:%s:%d:", len(name), name, len(data))
		fingerprint.Write(data)
	}
	result := Analyze(files)
	result.Root = root
	result.Fingerprint = fmt.Sprintf("%x", fingerprint.Sum(nil))
	result.Warnings = append(warnings, result.Warnings...)
	if len(result.Packages) == 0 {
		return result, fmt.Errorf("no readable Go, Python, JavaScript, or TypeScript source found in %s", root)
	}
	changes, available := readChanges(ctx, root)
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	result.History = available
	result.Hotspots = ReadingHints(result, changes)
	return result, nil
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
	if name == "go.mod" || name == "pyproject.toml" || name == "package.json" {
		return true
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".go", ".py", ".js", ".jsx", ".mjs", ".cjs", ".ts", ".tsx", ".md":
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
