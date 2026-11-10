package repository

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"repo-tutor/internal/repository/analyzer"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const maximumFileBytes = 512 * 1024
const maximumSnapshotBytes = 64 * 1024 * 1024

func Read(ctx context.Context, directory string) (analyzer.Snapshot, error) {
	source, err := readSource(ctx, directory)
	if err != nil {
		return analyzer.Snapshot{}, err
	}
	return analyzer.Build(source)
}

func readSource(ctx context.Context, directory string) (analyzer.Source, error) {
	root, err := filepath.Abs(directory)
	if err != nil {
		return analyzer.Source{}, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return analyzer.Source{}, fmt.Errorf("open repository: %w", err)
	}
	information, err := os.Stat(root)
	if err != nil || !information.IsDir() {
		return analyzer.Source{}, fmt.Errorf("repository must be a directory: %s", root)
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
			if name != root && analyzer.Excluded(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || !analyzer.Supported(entry.Name()) {
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
		return analyzer.Source{}, fmt.Errorf("scan repository: %w", err)
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
	var files []analyzer.File
	total := 0
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return analyzer.Source{}, err
		}
		file, err := os.Open(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return analyzer.Source{}, fmt.Errorf("read %s: %w", name, err)
		}
		data, readError := io.ReadAll(io.LimitReader(file, maximumFileBytes+1))
		closeError := file.Close()
		if readError != nil {
			return analyzer.Source{}, fmt.Errorf("read %s: %w", name, readError)
		}
		if closeError != nil {
			return analyzer.Source{}, closeError
		}
		if len(data) > maximumFileBytes || !utf8.Valid(data) || strings.ContainsRune(string(data), 0) {
			warnings = append(warnings, name+": skipped oversized or non-text file")
			continue
		}
		total += len(data)
		if total > maximumSnapshotBytes {
			return analyzer.Source{}, fmt.Errorf("repository source exceeds the 64 MiB scan limit; choose a smaller module directory")
		}
		files = append(files, analyzer.File{Path: name, Content: string(data)})
	}
	changeLog, history := readChanges(ctx, root)
	if err := ctx.Err(); err != nil {
		return analyzer.Source{}, err
	}
	return analyzer.Source{Root: root, Files: files, Warnings: warnings, ChangeLog: changeLog, History: history}, nil
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

func readChanges(ctx context.Context, root string) (string, bool) {
	bounded, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(bounded, "git", "log", "--relative", "--since=180.days", "--max-count=2000", "--no-merges", "--format=", "--name-only", "-z", "--no-renames", "--", ".")
	command.Dir = root
	data, err := command.Output()
	if err != nil {
		return "", false
	}
	return string(data), true
}
