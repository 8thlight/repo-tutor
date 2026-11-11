package reader

import "testing"

func TestLinkerVersionOverridesBuildInfo(t *testing.T) {
	version = "v0.2.0"
	t.Cleanup(func() { version = "" })
	if got := Read().Version; got != "v0.2.0" {
		t.Fatalf("version %q, want the linker value", got)
	}
}
