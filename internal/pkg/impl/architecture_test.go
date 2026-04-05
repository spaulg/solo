package impl

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/test"
)

func TestArchitectureTestSuite(t *testing.T) {
	suite.Run(t, new(ArchitectureTestSuite))
}

type ArchitectureTestSuite struct {
	suite.Suite

	projectRoot string
	moduleName  string

	isPackageInterfaceCacheMap map[string]map[string]bool
}

func (t *ArchitectureTestSuite) SetupTest() {
	var err error

	if t.projectRoot, err = test.GetProjectRoot(); err != nil {
		t.FailNowf("failed to detect project root", "%v", err)
	}

	if t.moduleName, err = test.GetModuleName(); err != nil {
		t.FailNowf("failed to detect module name", "%v", err)
	}

	t.isPackageInterfaceCacheMap = make(map[string]map[string]bool)
}

// TestConstructorsDontReturnInterfaces checks that constructors
// do not return interface types.
func (t *ArchitectureTestSuite) TestConstructorsDontReturnInterfaces() {
	fileset := token.NewFileSet()
	var violations []string

	for _, subDirectory := range []string{"cmd", "internal", "test"} {
		scanPath := filepath.Join(t.projectRoot, subDirectory)

		err := filepath.WalkDir(scanPath, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}

			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			parsedFile, err := parser.ParseFile(fileset, path, nil, parser.SkipObjectResolution)
			if err != nil {
				return err
			}

			imports := buildImportAliasMap(parsedFile)

			for _, declaration := range parsedFile.Decls {
				// Ignore non-function declarations
				fn, ok := declaration.(*ast.FuncDecl)
				if !ok {
					continue
				}

				// Ignore non-constructor functions
				if !strings.HasPrefix(fn.Name.Name, "New") {
					continue
				}

				// Ignore methods on structs
				if fn.Recv != nil && fn.Recv.NumFields() > 0 {
					continue
				}

				// Ignore functions that return nothing
				if fn.Type.Results == nil {
					continue
				}

				// For each returned parameter
				for _, result := range fn.Type.Results.List {
					pkgPath, isInterface, err := t.resolveInterfaceTypeInPackage(result.Type, imports)
					if err != nil {
						t.FailNowf("failed to resolve types package", "%v", err)
					}

					if isInterface {
						pos := fileset.Position(fn.Pos())
						filePath := strings.TrimPrefix(pos.Filename, t.projectRoot+string(filepath.Separator))

						violations = append(violations, fmt.Sprintf(
							"%s:%d: %s() returns interface type from %s",
							filePath, pos.Line, fn.Name.Name, pkgPath,
						))
					}
				}
			}

			return nil
		})

		if err != nil {
			t.FailNowf("failed to check constructor functions for returning interfaces", "%v", err)
		}
	}

	for _, v := range violations {
		t.Failf("Interface type returned from constructor function", "%s", v)
	}
}

// buildImportAliasMap builds a map of import aliases
// to their corresponding import paths.
func buildImportAliasMap(f *ast.File) map[string]string {
	importAliasMap := make(map[string]string)

	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		alias := ""

		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}

		if alias != "_" && alias != "." {
			importAliasMap[alias] = importPath
		}
	}

	return importAliasMap
}

// resolveInterfaceTypeInPackage checks if the given expression
// is an interface type from an imported package.
func (t *ArchitectureTestSuite) resolveInterfaceTypeInPackage(expr ast.Expr, imports map[string]string) (string, bool, error) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", false, nil
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", false, nil
	}

	pkgPath, found := imports[ident.Name]
	if !found {
		return "", false, nil
	}

	// Ignore third-party and stdlib packages
	if !strings.HasPrefix(pkgPath, t.moduleName) {
		return "", false, nil
	}

	// Check for interface types
	isInterface, err := t.isInterfaceTypeInPackage(pkgPath, sel.Sel.Name)
	if err != nil {
		return "", false, fmt.Errorf("failed to detect if interface type is in package: %v", err)
	}

	return pkgPath, isInterface, nil
}

// isInterfaceTypeInPackage checks if the specified type in the given package is an interface type.
// It uses a cache to avoid redundant parsing of packages.
func (t *ArchitectureTestSuite) isInterfaceTypeInPackage(pkgPath string, typeName string) (bool, error) {
	if typeMap, ok := t.isPackageInterfaceCacheMap[pkgPath]; ok {
		if isInterface, ok := typeMap[typeName]; ok {
			return isInterface, nil
		}
	} else {
		t.isPackageInterfaceCacheMap[pkgPath] = make(map[string]bool)
	}

	relPath := strings.TrimPrefix(pkgPath, t.moduleName+"/")
	dir := filepath.Join(t.projectRoot, relPath)

	fset := token.NewFileSet()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("failed to read package directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.SkipObjectResolution)
		if err != nil {
			return false, fmt.Errorf("failed to parse file %s: %v", entry.Name(), err)
		}

		for _, declaration := range file.Decls {
			genDecl, ok := declaration.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name.Name != typeName {
					continue
				}

				_, isInterface := ts.Type.(*ast.InterfaceType)
				t.isPackageInterfaceCacheMap[pkgPath][typeName] = isInterface

				return isInterface, nil
			}
		}
	}

	t.isPackageInterfaceCacheMap[pkgPath][typeName] = false
	return false, nil
}
