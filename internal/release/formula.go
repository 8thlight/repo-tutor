package release

import (
	"fmt"
	"strings"
	"text/template"
)

const repositoryURL = "https://github.com/8thlight/repo-tutor"

type archive struct{ URL, Checksum string }

type formulaFields struct {
	Version                                      string
	DarwinARM, DarwinIntel, LinuxARM, LinuxIntel archive
}

var formulaTemplate = template.Must(template.New("formula").Parse(`class RepoTutor < Formula
  desc "Teaches a code repository as a fast onboarding course in your terminal"
  homepage "` + repositoryURL + `"
  version "{{ .Version }}"
  license "GPL-3.0-or-later"

  on_macos do
    on_arm do
      url "{{ .DarwinARM.URL }}"
      sha256 "{{ .DarwinARM.Checksum }}"
    end
    on_intel do
      url "{{ .DarwinIntel.URL }}"
      sha256 "{{ .DarwinIntel.Checksum }}"
    end
  end

  on_linux do
    on_arm do
      url "{{ .LinuxARM.URL }}"
      sha256 "{{ .LinuxARM.Checksum }}"
    end
    on_intel do
      url "{{ .LinuxIntel.URL }}"
      sha256 "{{ .LinuxIntel.Checksum }}"
    end
  end

  def install
    bin.install "repo-tutor"
  end

  test do
    assert_match "repo-tutor", shell_output("#{bin}/repo-tutor --version")
  end
end
`))

func checksumsByFile(checksums string) map[string]string {
	found := map[string]string{}
	for _, line := range strings.Split(checksums, "\n") {
		if fields := strings.Fields(line); len(fields) == 2 {
			found[fields[1]] = fields[0]
		}
	}
	return found
}

func Formula(tag, checksums string) (string, error) {
	known := checksumsByFile(checksums)
	find := func(system string) (archive, error) {
		name := fmt.Sprintf("repo-tutor_%s_%s.tar.gz", tag, system)
		checksum, found := known[name]
		if !found {
			return archive{}, fmt.Errorf("no checksum for %s", name)
		}
		return archive{URL: fmt.Sprintf("%s/releases/download/%s/%s", repositoryURL, tag, name), Checksum: checksum}, nil
	}
	fields := formulaFields{Version: strings.TrimPrefix(tag, "v")}
	for system, target := range map[string]*archive{"darwin_arm64": &fields.DarwinARM, "darwin_amd64": &fields.DarwinIntel, "linux_arm64": &fields.LinuxARM, "linux_amd64": &fields.LinuxIntel} {
		found, err := find(system)
		if err != nil {
			return "", err
		}
		*target = found
	}
	var formula strings.Builder
	if err := formulaTemplate.Execute(&formula, fields); err != nil {
		return "", err
	}
	return formula.String(), nil
}
