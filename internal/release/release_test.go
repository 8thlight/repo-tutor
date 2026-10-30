package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type releaseFixture struct {
	t                   *testing.T
	work, remote, tools string
}

func gitIn(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com"}, arguments...)...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
	return strings.TrimSpace(string(output))
}

func newReleaseFixture(t *testing.T) releaseFixture {
	t.Helper()
	root := t.TempDir()
	fixture := releaseFixture{t: t, work: filepath.Join(root, "work"), remote: filepath.Join(root, "remote.git"), tools: filepath.Join(root, "tools")}
	for _, directory := range []string{fixture.work, fixture.tools} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	gitIn(t, root, "init", "--quiet", "--bare", "--initial-branch=main", fixture.remote)
	gitIn(t, fixture.work, "init", "--quiet", "--initial-branch=main")
	gitIn(t, fixture.work, "remote", "add", "origin", fixture.remote)
	os.WriteFile(filepath.Join(fixture.work, "VERSION"), []byte("0.1.0\n"), 0o644)
	os.WriteFile(filepath.Join(fixture.work, "justfile"), []byte("check:\n    @echo checked\n"), 0o644)
	gitIn(t, fixture.work, "add", ".")
	gitIn(t, fixture.work, "commit", "--quiet", "-m", "start")
	gitIn(t, fixture.work, "push", "--quiet", "origin", "main")
	fixture.fakeGitHub("success")
	return fixture
}

func (fixture releaseFixture) fakeGitHub(conclusion string) {
	script := "#!/bin/sh\necho " + conclusion + "\n"
	if err := os.WriteFile(filepath.Join(fixture.tools, "gh"), []byte(script), 0o755); err != nil {
		fixture.t.Fatal(err)
	}
}

func (fixture releaseFixture) release(version string) (string, error) {
	script, err := filepath.Abs("release.sh")
	if err != nil {
		fixture.t.Fatal(err)
	}
	command := exec.Command("bash", script, version)
	command.Dir = fixture.work
	command.Env = append(os.Environ(), "PATH="+fixture.tools+string(os.PathListSeparator)+os.Getenv("PATH"), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	output, err := command.CombinedOutput()
	return string(output), err
}

func (fixture releaseFixture) remoteTags() string {
	return gitIn(fixture.t, fixture.remote, "tag", "--list")
}

func TestReleasePushesTheBumpAndTheTagTogether(t *testing.T) {
	if _, err := exec.LookPath("just"); err != nil {
		t.Skip("just is not installed")
	}
	fixture := newReleaseFixture(t)
	if output, err := fixture.release("0.2.0"); err != nil {
		t.Fatalf("release failed: %v\n%s", err, output)
	}
	if got := fixture.remoteTags(); got != "v0.2.0" {
		t.Fatalf("the remote should have the tag, got %q", got)
	}
	if got := gitIn(t, fixture.remote, "show", "main:VERSION"); got != "0.2.0" {
		t.Fatalf("the pushed main should hold the new VERSION, got %q", got)
	}
	if tagged, head := gitIn(t, fixture.remote, "rev-parse", "v0.2.0^{commit}"), gitIn(t, fixture.remote, "rev-parse", "main"); tagged != head {
		t.Fatal("the tag should point at the bump commit")
	}
	if kind := gitIn(t, fixture.remote, "cat-file", "-t", "v0.2.0"); kind != "tag" {
		t.Fatalf("the tag should be annotated, got %s", kind)
	}
	if subject := gitIn(t, fixture.remote, "log", "-1", "--format=%s", "main"); subject != "Release v0.2.0" {
		t.Fatalf("unexpected commit subject %q", subject)
	}
}

func TestReleaseRefusesAndPushesNothing(t *testing.T) {
	if _, err := exec.LookPath("just"); err != nil {
		t.Skip("just is not installed")
	}
	for name, scenario := range map[string]struct {
		version string
		arrange func(releaseFixture)
		message string
	}{
		"a version that is not a release number": {"0.2", func(releaseFixture) {}, "not MAJOR.MINOR.PATCH"},
		"the same version":                       {"0.1.0", func(releaseFixture) {}, "not newer"},
		"an older version":                       {"0.0.9", func(releaseFixture) {}, "not newer"},
		"a dirty working tree":                   {"0.2.0", func(f releaseFixture) { os.WriteFile(filepath.Join(f.work, "stray.txt"), []byte("x"), 0o644) }, "has changes"},
		"another branch":                         {"0.2.0", func(f releaseFixture) { gitIn(f.t, f.work, "switch", "--quiet", "-c", "topic") }, "switch to main"},
		"a main that is behind its remote": {"0.2.0", func(f releaseFixture) {
			other := filepath.Join(f.t.TempDir(), "other")
			gitIn(f.t, filepath.Dir(other), "clone", "--quiet", f.remote, other)
			os.WriteFile(filepath.Join(other, "new.txt"), []byte("x"), 0o644)
			gitIn(f.t, other, "add", ".")
			gitIn(f.t, other, "commit", "--quiet", "-m", "ahead")
			gitIn(f.t, other, "push", "--quiet", "origin", "main")
		}, "not level"},
		"an existing tag":                     {"0.2.0", func(f releaseFixture) { gitIn(f.t, f.work, "tag", "v0.2.0") }, "already exists"},
		"a commit whose check has not passed": {"0.2.0", func(f releaseFixture) { f.fakeGitHub("failure") }, "has not passed"},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newReleaseFixture(t)
			scenario.arrange(fixture)
			before := gitIn(t, fixture.remote, "rev-parse", "main")
			output, err := fixture.release(scenario.version)
			if err == nil || !strings.Contains(output, scenario.message) {
				t.Fatalf("expected a refusal containing %q, got err=%v\n%s", scenario.message, err, output)
			}
			if gitIn(t, fixture.remote, "rev-parse", "main") != before || strings.Contains(fixture.remoteTags(), "v0.2.0") {
				t.Fatal("a refused release must not push anything")
			}
		})
	}
}
