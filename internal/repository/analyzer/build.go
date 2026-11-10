package analyzer

import (
	"crypto/sha256"
	"fmt"
)

type Source struct {
	Root      string
	Files     []File
	Warnings  []string
	ChangeLog string
	History   bool
}

func Build(source Source) (Snapshot, error) {
	result := Analyze(source.Files)
	fingerprint := sha256.New()
	for _, file := range result.Files {
		fmt.Fprintf(fingerprint, "%d:%s:%d:", len(file.Path), file.Path, len(file.Content))
		fingerprint.Write([]byte(file.Content))
	}
	result.Root = source.Root
	result.Fingerprint = fmt.Sprintf("%x", fingerprint.Sum(nil))
	result.Warnings = append(append([]string(nil), source.Warnings...), result.Warnings...)
	if len(result.Packages) == 0 {
		return result, fmt.Errorf("no readable Go, Python, JavaScript, TypeScript, or Java source found in %s", source.Root)
	}
	result.History = source.History
	result.Hotspots = ReadingHints(result, ChangeCounts(source.ChangeLog))
	return result, nil
}
