package repository_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"repo-tutor/internal/repository"
)

func TestReadExclusionsAndFingerprint(t *testing.T) {
	root := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		full := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module course\n")
	write("main.go", "package main\nfunc main() {}\n")
	write("README.md", "# Course\n")
	for _, directory := range []string{".git", ".scratch", "vendor", "node_modules", "output", "fixtures"} {
		write(directory+"/secret.go", "package ignored\n")
	}
	outside := filepath.Join(t.TempDir(), "outside.go")
	if err := os.WriteFile(outside, []byte("package outside\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.go")); err != nil {
		t.Fatal(err)
	}
	write("large.md", strings.Repeat("x", 512*1024+1))
	first, err := repository.Read(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Files) != 3 || len(first.Warnings) != 1 {
		t.Fatalf("excluded files entered the snapshot: %+v", first)
	}
	second, err := repository.Read(context.Background(), root)
	if err != nil || first.Fingerprint != second.Fingerprint {
		t.Fatalf("unchanged source produced a different fingerprint: %v", err)
	}
	write("main.go", "package main\nfunc main() { println(1) }\n")
	third, err := repository.Read(context.Background(), root)
	if err != nil || first.Fingerprint == third.Fingerprint {
		t.Fatalf("changed source did not invalidate the fingerprint: %v", err)
	}
}

func TestGitIgnoredFilesAreNotRead(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is optional")
	}
	root := t.TempDir()
	if output, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	for name, content := range map[string]string{
		".gitignore":             "ignored/\n",
		"main.py":                "print('hello')\n",
		"ignored/credentials.py": "secret = 'never include me'\n",
	} {
		full := filepath.Join(root, name)
		os.MkdirAll(filepath.Dir(full), 0700)
		if err := os.WriteFile(full, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := repository.Read(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Files) != 1 || snapshot.Files[0].Path != "main.py" {
		t.Fatalf("Git ignored source was included: %v", snapshot.Files)
	}
}

func TestHistoryCountsLocalCommits(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is optional")
	}
	root := t.TempDir()
	git := func(arguments ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-c", "user.name=Tutor Test", "-c", "user.email=tutor@example.invalid", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null"}, arguments...)...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", arguments, err, output)
		}
	}
	git("init", "-q")
	for _, name := range []string{"alpha.py", "beta.py"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("if True:\n    pass\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	git("add", ".")
	git("commit", "-qm", "initial source")
	if err := os.WriteFile(filepath.Join(root, "alpha.py"), []byte("if False:\n    pass\n"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-qm", "change alpha")
	snapshot, err := repository.Read(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.History || len(snapshot.Hotspots) != 2 || snapshot.Hotspots[0].Path != "alpha.py" || snapshot.Hotspots[0].Commits != 2 || snapshot.Hotspots[1].Commits != 1 {
		t.Fatalf("local Git history did not reach the reading hints: %+v", snapshot.Hotspots)
	}
}
