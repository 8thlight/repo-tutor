package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runIn(t *testing.T, directory string, environment []string, name string, arguments ...string) string {
	t.Helper()
	command := exec.Command(name, arguments...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, arguments, err, output)
	}
	return string(output)
}

func TestUpdateHomebrewTapPushesTheFormulaToTheNamedBranchOnly(t *testing.T) {
	if _, err := exec.LookPath("just"); err != nil {
		t.Skip("just is not installed")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	tap := filepath.Join(t.TempDir(), "tap.git")
	runIn(t, "", nil, "git", "init", "--quiet", "--bare", "--initial-branch=main", tap)
	seed := t.TempDir()
	runIn(t, seed, nil, "git", "init", "--quiet", "--initial-branch=main")
	os.WriteFile(filepath.Join(seed, "README.md"), []byte("tap\n"), 0o600)
	runIn(t, seed, nil, "git", "add", ".")
	runIn(t, seed, nil, "git", "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "--quiet", "-m", "start")
	runIn(t, seed, nil, "git", "push", "--quiet", tap, "main")

	formulaDirectory := filepath.Join(root, "output", "release")
	os.MkdirAll(formulaDirectory, 0o755)
	formulaPath := filepath.Join(formulaDirectory, "repo-tutor.rb")
	previous, notThere := os.ReadFile(formulaPath)
	t.Cleanup(func() {
		if notThere == nil {
			os.WriteFile(formulaPath, previous, 0o644)
		} else {
			os.Remove(formulaPath)
		}
	})
	if err := os.WriteFile(formulaPath, []byte("class RepoTutor < Formula\nend\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	environment := []string{"TAP_DEPLOY_KEY=not-a-real-key", "TAP_URL=" + tap}
	runIn(t, root, environment, "just", "update-homebrew-tap", "v0.0.0-dryrun", "dryrun")

	if branches := runIn(t, tap, nil, "git", "branch", "--list"); !strings.Contains(branches, "dryrun") {
		t.Fatalf("the dryrun branch should exist:\n%s", branches)
	}
	if listing := runIn(t, tap, nil, "git", "ls-tree", "-r", "--name-only", "main"); strings.Contains(listing, "Formula") {
		t.Fatalf("main must not receive the formula:\n%s", listing)
	}
	if listing := runIn(t, tap, nil, "git", "ls-tree", "-r", "--name-only", "dryrun"); !strings.Contains(listing, "Formula/repo-tutor.rb") {
		t.Fatalf("the dryrun branch should hold the formula:\n%s", listing)
	}
	runIn(t, root, environment, "just", "update-homebrew-tap", "v0.1.0", "main")
	if again := runIn(t, root, environment, "just", "update-homebrew-tap", "v0.1.0", "main"); !strings.Contains(again, "already has this formula") {
		t.Fatalf("a second run for the same formula should change nothing:\n%s", again)
	}
}
