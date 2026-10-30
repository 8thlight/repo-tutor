package release

import (
	"os"
	"regexp"
	"testing"
)

func TestVersionFileHoldsOneReleaseNumberAndANewline(t *testing.T) {
	data, err := os.ReadFile("../../VERSION")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^\d+\.\d+\.\d+\n$`).Match(data) {
		t.Fatalf("VERSION should hold only MAJOR.MINOR.PATCH and a newline, got %q", data)
	}
}
