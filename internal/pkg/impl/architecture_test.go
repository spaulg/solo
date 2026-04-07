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

	archgo "github.com/arch-go/arch-go/api"
	archgo_config "github.com/arch-go/arch-go/api/configuration"
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

func (t *ArchitectureTestSuite) TestLayerDependencies() {
	t.T().Skipf("Skipping architecture test as the architecture is still evolving." +
		" Re-enable once the architecture is stable.")

	configuration := archgo_config.Config{
		DependenciesRules: []*archgo_config.DependenciesRule{
			// Commands
			{
				Package: t.moduleName + "/cmd/solo.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/cmd/solo.**",
						t.moduleName + "/internal/pkg/impl/host.**",
					},
				},
			},

			{
				Package: t.moduleName + "/cmd/solo-entrypoint.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/cmd/solo-entrypoint.**",
						t.moduleName + "/internal/pkg/impl/entrypoint.**",
					},
				},
			},

			// Host
			// App -> App/Domain/Infra
			{
				Package: t.moduleName + "/internal/pkg/impl/host/app.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/internal/pkg/types/host.**",
						t.moduleName + "/internal/pkg/impl/host.**",
						t.moduleName + "/internal/pkg/impl/common.**",
					},
				},
			},

			// Infra -> domain
			{
				Package: t.moduleName + "/internal/pkg/impl/host/infra.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/internal/pkg/types/host/infra.**",
						t.moduleName + "/internal/pkg/impl/host/infra.**",
						t.moduleName + "/internal/pkg/impl/common/infra.**",
						t.moduleName + "/internal/pkg/types/host/domain.**",
						t.moduleName + "/internal/pkg/impl/host/domain.**",
						t.moduleName + "/internal/pkg/impl/common/domain.**",
					},
				},
			},

			// Domain -> nothing
			{
				Package: t.moduleName + "/internal/pkg/impl/host/domain.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/internal/pkg/types/host/domain.**",
						t.moduleName + "/internal/pkg/impl/host/domain.**",
						t.moduleName + "/internal/pkg/impl/common/domain.**",
					},
				},
			},

			// Entrypoint
			// App -> App/Domain/Infra
			{
				Package: t.moduleName + "/internal/pkg/impl/entrypoint/app.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/internal/pkg/types/entrypoint.**",
						t.moduleName + "/internal/pkg/impl/entrypoint.**",
						t.moduleName + "/internal/pkg/impl/common.**",
					},
				},
			},

			// Infra -> domain
			{
				Package: t.moduleName + "/internal/pkg/impl/entrypoint/infra.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/internal/pkg/types/entrypoint/infra.**",
						t.moduleName + "/internal/pkg/impl/entrypoint/infra.**",
						t.moduleName + "/internal/pkg/impl/common/infra.**",
					},
				},
			},

			// Common
			{
				Package: t.moduleName + "/internal/pkg/impl/common/app.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/internal/pkg/impl/common.**",
					},
				},
			},

			{
				Package: t.moduleName + "/internal/pkg/impl/common/infra.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/internal/pkg/impl/common/infra.**",
						t.moduleName + "/internal/pkg/impl/common/domain.**",
					},
				},
			},

			{
				Package: t.moduleName + "/internal/pkg/impl/common/domain.**",
				ShouldOnlyDependsOn: &archgo_config.Dependencies{
					Internal: []string{
						t.moduleName + "/internal/pkg/impl/common/domain.**",
					},
				},
			},
		},
	}

	moduleInfo := archgo_config.Load(t.moduleName)
	if result := archgo.CheckArchitecture(moduleInfo, configuration); !result.Pass {
		if result.DependenciesRuleResult != nil {
			for _, depRule := range result.DependenciesRuleResult.Results {
				if depRule.Passes {
					continue
				}

				for _, verification := range depRule.Verifications {
					if verification.Passes {
						continue
					}

					t.Failf("Incorrect layer dependency in package "+verification.Package,
						strings.Join(verification.Details, "\n"))
				}
			}
		}
	}
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
