package repository

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Hotspot struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Language string `json:"language"`
	Commits  int    `json:"recent_commits"`
	Branches int    `json:"branch_points"`
	Lines    int    `json:"lines"`
	Score    int    `json:"reading_score"`
}

var pythonBranch = regexp.MustCompile(`\b(if|elif|for|while|except|case|and|or)\b`)
var javascriptBranch = regexp.MustCompile(`\b(if|for|while|case|catch)\b|&&|\|\||\?\?`)

func ReadingHints(snapshot Snapshot, changes map[string]int) []Hotspot {
	files := map[string]string{}
	for _, current := range snapshot.Packages {
		for _, name := range current.Files {
			files[name] = current.Language
		}
	}
	var result []Hotspot
	for _, file := range snapshot.Files {
		language, included := files[file.Path]
		if !included {
			continue
		}
		branches, line := branchPoints(file, language)
		lines := strings.Count(strings.TrimSuffix(file.Content, "\n"), "\n") + 1
		commits := changes[file.Path]
		if branches == 0 && commits == 0 {
			continue
		}
		result = append(result, Hotspot{file.Path, line, language, commits, branches, lines, (commits + 1) * (branches + 1)})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score != result[j].Score {
			return result[i].Score > result[j].Score
		}
		return result[i].Path < result[j].Path
	})
	return result
}

func branchPoints(file File, language string) (int, int) {
	count, first := 0, 1
	if language == "Go" {
		positions := token.NewFileSet()
		parsed, err := parser.ParseFile(positions, file.Path, file.Content, parser.SkipObjectResolution)
		if err != nil {
			return 0, 1
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			branch := false
			switch node := node.(type) {
			case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
				branch = true
			case *ast.CaseClause:
				branch = len(node.List) > 0
			case *ast.CommClause:
				branch = node.Comm != nil
			case *ast.BinaryExpr:
				branch = node.Op == token.LAND || node.Op == token.LOR
			}
			if branch {
				if count == 0 {
					first = positions.Position(node.Pos()).Line
				}
				count++
			}
			return true
		})
		return count, first
	}
	pattern := javascriptBranch
	if language == "Python" {
		pattern = pythonBranch
	}
	for index, line := range strings.Split(file.Content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}
		matches := len(pattern.FindAllStringIndex(line, -1))
		if matches > 0 && count == 0 {
			first = index + 1
		}
		count += matches
	}
	return count, first
}

func readChanges(ctx context.Context, root string) (map[string]int, bool) {
	bounded, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(bounded, "git", "log", "--relative", "--since=180.days", "--max-count=2000", "--no-merges", "--format=", "--name-only", "-z", "--no-renames", "--", ".")
	command.Dir = root
	data, err := command.Output()
	if err != nil {
		return nil, false
	}
	return ChangeCounts(string(data)), true
}

func ChangeCounts(history string) map[string]int {
	result := map[string]int{}
	for _, name := range strings.Split(history, "\x00") {
		name = strings.TrimPrefix(name, "\n")
		if name != "" && !strings.HasPrefix(name, "/") && path.Clean(name) == name {
			result[name]++
		}
	}
	return result
}
