package internal_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// AC-4 & TP-6 supplemental: Verify architectural package boundaries and seam isolation.
// - DAO (internal/store) must not import HTTP packages or internal/idp or internal/api.
// - IdP connector (internal/idp) must not import internal/store or internal/api.
func TestArchitecturalPackageBoundaries(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current test file path via runtime.Caller")
	}
	internalDir := filepath.Dir(currentFile)

	checkPackage := func(pkgName string, forbiddenImports []string) {
		dir := filepath.Join(internalDir, pkgName)
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("required directory %s does not exist or cannot be accessed: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}

		goFilesChecked := 0
		err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf("walk error at %s: %w", path, walkErr)
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			goFilesChecked++
			fset := token.NewFileSet()
			node, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if parseErr != nil {
				return fmt.Errorf("failed to parse imports in %s: %w", path, parseErr)
			}

			for _, imp := range node.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				for _, forbidden := range forbiddenImports {
					if importPath == forbidden || strings.Contains(importPath, forbidden) {
						t.Errorf("Seam violation in %s: package %s must not import %s (got %s)", path, pkgName, forbidden, importPath)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("failed to walk %s dir: %v", pkgName, err)
		}
		if goFilesChecked == 0 {
			t.Fatalf("expected at least one non-test .go file in %s, got 0", dir)
		}
	}

	// 1. Check internal/store imports: DAO must not import internal/idp, internal/api, or net/http
	checkPackage("store", []string{"internal/idp", "internal/api", "net/http"})

	// 2. Check internal/idp imports: IdP connector must not import internal/store or internal/api
	checkPackage("idp", []string{"internal/store", "internal/api"})
}
