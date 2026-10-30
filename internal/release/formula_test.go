package release

import (
	"strings"
	"testing"
)

const sampleChecksums = `1111111111111111111111111111111111111111111111111111111111111111  repo-tutor_v0.1.0_darwin_amd64.tar.gz
2222222222222222222222222222222222222222222222222222222222222222  repo-tutor_v0.1.0_darwin_arm64.tar.gz
3333333333333333333333333333333333333333333333333333333333333333  repo-tutor_v0.1.0_linux_amd64.tar.gz
4444444444444444444444444444444444444444444444444444444444444444  repo-tutor_v0.1.0_linux_arm64.tar.gz
`

func TestFormulaNamesEveryArchiveWithItsChecksum(t *testing.T) {
	formula, err := Formula("v0.1.0", sampleChecksums)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		`version "0.1.0"`,
		`license "GPL-3.0-or-later"`,
		`url "https://github.com/8thlight/repo-tutor/releases/download/v0.1.0/repo-tutor_v0.1.0_darwin_arm64.tar.gz"`,
		`sha256 "2222222222222222222222222222222222222222222222222222222222222222"`,
		`url "https://github.com/8thlight/repo-tutor/releases/download/v0.1.0/repo-tutor_v0.1.0_darwin_amd64.tar.gz"`,
		`sha256 "1111111111111111111111111111111111111111111111111111111111111111"`,
		`url "https://github.com/8thlight/repo-tutor/releases/download/v0.1.0/repo-tutor_v0.1.0_linux_arm64.tar.gz"`,
		`sha256 "4444444444444444444444444444444444444444444444444444444444444444"`,
		`url "https://github.com/8thlight/repo-tutor/releases/download/v0.1.0/repo-tutor_v0.1.0_linux_amd64.tar.gz"`,
		`sha256 "3333333333333333333333333333333333333333333333333333333333333333"`,
		`bin.install "repo-tutor"`,
	} {
		if !strings.Contains(formula, expected) {
			t.Fatalf("the formula is missing %q:\n%s", expected, formula)
		}
	}
}

func TestFormulaPlacesEachChecksumUnderItsOwnPlatform(t *testing.T) {
	formula, _ := Formula("v0.1.0", sampleChecksums)
	for platform, want := range map[string]string{"on_macos": "darwin_arm64", "on_linux": "linux_amd64"} {
		block := formula[strings.Index(formula, platform):]
		if !strings.Contains(block, want) {
			t.Fatalf("%s should hold %s", platform, want)
		}
	}
	if strings.Index(formula, "on_macos") > strings.Index(formula, "on_linux") {
		t.Fatal("the macOS block should come first")
	}
}

func TestFormulaRefusesAReleaseThatLacksAnArchive(t *testing.T) {
	missing := strings.ReplaceAll(sampleChecksums, "linux_arm64", "linux_riscv")
	if _, err := Formula("v0.1.0", missing); err == nil || !strings.Contains(err.Error(), "linux_arm64") {
		t.Fatalf("a missing archive should be named in the error: %v", err)
	}
}

func TestFormulaVersionDropsTheTagPrefix(t *testing.T) {
	checksums := strings.ReplaceAll(sampleChecksums, "v0.1.0", "v0.2.0-rc1")
	formula, err := Formula("v0.2.0-rc1", checksums)
	if err != nil || !strings.Contains(formula, `version "0.2.0-rc1"`) {
		t.Fatalf("version should be 0.2.0-rc1: %v\n%s", err, formula)
	}
}
