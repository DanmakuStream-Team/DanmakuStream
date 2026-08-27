// Command userdomaincoverage reports statement coverage for UC01/07/08/11 only.
// It excludes unrelated handlers that happen to share the user package and the
// perpetual expiration-worker bootstrap, which is infrastructure rather than a
// request path.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type functionRange struct {
	name       string
	start, end int
}

type coverageBlock struct {
	statements int
	covered    bool
}

var selectedFiles = map[string]bool{
	"internal/handler/v1/auth/auth_handler.go":                 true,
	"internal/logic/auth/auth_logic.go":                        true,
	"internal/handler/v1/notification/notification_handler.go": true,
	"internal/handler/v1/message/message_handler.go":           true,
	"internal/handler/v1/user/relationship_handler.go":         true,
	"internal/logic/chat/hub.go":                               true,
}

var selectedUserFunctions = map[string]bool{
	"ProfileHandler": true, "getOptionalUserID": true, "FollowHandler": true,
	"UpdateMeHandler": true, "FollowingListHandler": true, "UploadAvatarHandler": true,
}

func main() {
	profile := flag.String("profile", "user-domain.cover.out", "Go cover profile")
	root := flag.String("root", ".", "backend module root")
	flag.Parse()

	functions := map[string][]functionRange{}
	blocks := map[string]coverageBlock{}
	file, err := os.Open(*profile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			log.Fatalf("invalid profile row: %s", line)
		}
		colon := strings.LastIndex(fields[0], ":")
		if colon < 0 {
			log.Fatalf("invalid source range: %s", fields[0])
		}
		source := strings.TrimPrefix(filepath.ToSlash(fields[0][:colon]), "danmakustream/backend/")
		startText := strings.SplitN(fields[0][colon+1:], ",", 2)[0]
		startLine, err := strconv.Atoi(strings.SplitN(startText, ".", 2)[0])
		if err != nil {
			log.Fatal(err)
		}

		if _, loaded := functions[source]; !loaded {
			functions[source] = parseFunctions(filepath.Join(*root, filepath.FromSlash(source)))
		}
		name := containingFunction(functions[source], startLine)
		if !selected(source, name) {
			continue
		}
		statements, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatal(err)
		}
		count, err := strconv.Atoi(fields[2])
		if err != nil {
			log.Fatal(err)
		}
		key := fields[0]
		block := blocks[key]
		if block.statements != 0 && block.statements != statements {
			log.Fatalf("inconsistent statement count for %s", key)
		}
		block.statements = statements
		block.covered = block.covered || count > 0
		blocks[key] = block
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	covered, total := 0, 0
	for _, block := range blocks {
		total += block.statements
		if block.covered {
			covered += block.statements
		}
	}
	if total == 0 {
		log.Fatal("no UC01/07/08/11 statements found")
	}
	fmt.Printf("UC01/07/08/11 statement coverage: %.1f%% (%d/%d)\n", float64(covered)*100/float64(total), covered, total)
}

func parseFunctions(path string) []functionRange {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil
	}
	result := make([]functionRange, 0)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		result = append(result, functionRange{
			name: function.Name.Name, start: fset.Position(function.Pos()).Line, end: fset.Position(function.End()).Line,
		})
	}
	return result
}

func containingFunction(functions []functionRange, line int) string {
	for _, function := range functions {
		if line >= function.start && line <= function.end {
			return function.name
		}
	}
	return ""
}

func selected(source, function string) bool {
	if selectedFiles[source] {
		return true
	}
	if source == "internal/handler/v1/membership/membership_handler.go" {
		return function != "" && function != "StartExpirationWorker"
	}
	if source == "internal/handler/v1/user/user_handler.go" {
		return selectedUserFunctions[function]
	}
	return false
}
