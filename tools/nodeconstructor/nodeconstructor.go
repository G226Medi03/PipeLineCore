// Package nodeconstructor provides a golangci-lint module plugin that forbids
// direct construction of PipeLineCore Node values outside the core package.
package nodeconstructor

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

const corePackagePath = "github.com/G226Medi03/PipeLineCore"

// Analyzer forbids direct construction of core.Node.
var Analyzer = &analysis.Analyzer{
	Name: "nodeconstructor",
	Doc:  "forbid direct construction of core.Node",
	Run:  run,
}

func init() {
	register.Plugin("nodeconstructor", New)
}

// Plugin is the golangci-lint module plugin entry point.
type Plugin struct{}

// New creates the nodeconstructor plugin.
func New(any) (register.LinterPlugin, error) {
	return &Plugin{}, nil
}

// BuildAnalyzers returns the nodeconstructor analyzer.
func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{Analyzer}, nil
}

// GetLoadMode requests type information for the analyzer.
func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Path() == corePackagePath {
		return nil, nil
	}
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.CompositeLit:
				reportNodeType(pass, pass.TypesInfo.TypeOf(value), value.Pos())
			case *ast.ValueSpec:
				if value.Type != nil && len(value.Values) == 0 {
					reportNodeType(pass, pass.TypesInfo.TypeOf(value.Type), value.Pos())
				}
			case *ast.CallExpr:
				reportNewNode(pass, value)
			}
			return true
		})
	}
	return nil, nil
}

func reportNewNode(pass *analysis.Pass, call *ast.CallExpr) {
	identifier, ok := call.Fun.(*ast.Ident)
	if !ok || identifier.Name != "new" || len(call.Args) != 1 {
		return
	}
	if _, ok := pass.TypesInfo.Uses[identifier].(*types.Builtin); !ok {
		return
	}
	reportNodeType(pass, pass.TypesInfo.TypeOf(call.Args[0]), call.Pos())
}

func reportNodeType(pass *analysis.Pass, typ types.Type, pos token.Pos) {
	if typ == nil {
		return
	}
	named, ok := types.Unalias(typ).(*types.Named)
	if !ok {
		return
	}
	object := named.Obj()
	if object == nil || object.Pkg() == nil {
		return
	}
	if object.Pkg().Path() != corePackagePath || object.Name() != "Node" {
		return
	}
	pass.Reportf(pos, "direct construction of core.Node is forbidden; use core.NewNode")
}
