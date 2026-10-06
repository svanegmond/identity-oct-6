package internal_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-4 & TP-6 supplemental: Verify architectural package boundaries and seam isolation.
// - DAO (internal/store) must not import HTTP packages or internal/idp.
// - IdP connector (internal/idp) must not import internal/store.
func TestArchitecturalPackageBoundaries(t *testing.T) {
	root := ".." // internal directory

	// 1. Check internal/store imports
	storeDir := filepath.Join(root, "store")
	err := filepath.Walk(storeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(importPath, "internal/idp") {
				t.Errorf("Seam violation in %s: store package must not import internal/idp (got %s)", path, importPath)
			}
			if strings.Contains(importPath, "internal/api") {
				t.Errorf("Seam violation in %s: store package must not import internal/api (got %s)", path, importPath)
			}
			if importPath == "net/http" {
				t.Errorf("Seam violation in %s: store package must not import net/http (got %s)", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk store dir: %v", err)
	}

	// 2. Check internal/idp imports
	idpDir := filepath.Join(root, "idp")
	err = filepath.Walk(idpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(importPath, "internal/store") {
				t.Errorf("Seam violation in %s: idp package must not import internal/store (got %s)", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk idp dir: %v", err)
	}
}
