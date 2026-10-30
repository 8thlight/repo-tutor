package main

import (
	"fmt"
	"os"

	"repo-tutor/internal/release"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: formula CHECKSUMS_FILE TAG")
		os.Exit(2)
	}
	checksums, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	formula, err := release.Formula(os.Args[2], string(checksums))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(formula)
}
