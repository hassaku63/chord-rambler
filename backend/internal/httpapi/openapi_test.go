package httpapi

import (
	"bufio"
	"os"
	"slices"
	"strings"
	"testing"
)

// openAPIPath is where the API contract lives relative to this package.
const openAPIPath = "../../../docs/api/openapi.yaml"

// TestRoutesMatchOpenAPI keeps the server and the contract in step
// (ADR-0008): every registered route is an operation of
// docs/api/openapi.yaml, and every operation there is registered.
func TestRoutesMatchOpenAPI(t *testing.T) {
	documented := operationsFromOpenAPI(t)

	var implemented []string
	for _, r := range routes {
		implemented = append(implemented, strings.ToLower(r.method)+" "+r.path)
	}
	slices.Sort(implemented)
	slices.Sort(documented)

	if !slices.Equal(implemented, documented) {
		t.Fatalf("the routes and %s disagree\nimplemented:\n  %s\ndocumented:\n  %s",
			openAPIPath, strings.Join(implemented, "\n  "), strings.Join(documented, "\n  "))
	}
}

// operationsFromOpenAPI lists the contract's operations as "method /path".
// It reads the document line by line and relies on its two-space
// indentation (`  /path:` under `paths:`, `    get:` under a path), so that
// no YAML dependency is needed. It fails when the paths block does not have
// that shape.
func operationsFromOpenAPI(t *testing.T) []string {
	t.Helper()
	f, err := os.Open(openAPIPath)
	if err != nil {
		t.Fatalf("opening the contract: %v", err)
	}
	defer func() { _ = f.Close() }()

	methods := []string{"get", "post", "put", "patch", "delete"}
	var operations []string
	inPaths := false
	path := ""
	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		line := scanner.Text()
		switch {
		case line == "paths:":
			inPaths = true
		case !inPaths || strings.TrimSpace(line) == "":
		case strings.Contains(line, "\t"):
			t.Fatalf("%s:%d: a tab in the paths block; indent with two spaces", openAPIPath, n)
		case line[0] != ' ':
			inPaths = false // the next top-level key
		case indent(line) == 2:
			key := strings.TrimSuffix(strings.TrimSpace(line), ":")
			if !strings.HasPrefix(key, "/") || !strings.HasSuffix(line, ":") {
				t.Fatalf("%s:%d: expected a path such as `  /health:` under paths, got %q", openAPIPath, n, line)
			}
			path = key
		case indent(line) == 4:
			key := strings.TrimSuffix(strings.TrimSpace(line), ":")
			if slices.Contains(methods, key) {
				operations = append(operations, key+" "+path)
			}
		case indent(line)%2 != 0:
			t.Fatalf("%s:%d: odd indentation in the paths block; indent with two spaces", openAPIPath, n)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("reading the contract: %v", err)
	}
	if len(operations) == 0 {
		t.Fatalf("no operations found in %s; its paths block may not have the expected shape", openAPIPath)
	}
	return operations
}

func indent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}
