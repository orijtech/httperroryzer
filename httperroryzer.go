// Copyright 2020 Orijtech, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Package httperroryzer defines an Analyzer that checks for
// missing terminating statements after invoking http.Error.
package httperroryzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/cfg"
)

const Doc = `check for a missing return after invoking http.Error

A common mistake when using the net/http package is to forget to invoke
return after a call to http.Error

    if err != nil {
         http.Error(w, err.Error(), statusCode)
    }

    // Code that assumes the error was properly handled.
    slurp, _ := ioutil.ReadAll(res.Body)

This checker helps uncover latent nil dereference bugs, or security problems by reporting a
diagnostic for such mistakes.`

var Analyzer = &analysis.Analyzer{
	Name: "httperroryzer",
	Doc:  Doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		ctrlflow.Analyzer,
	},
}

// Imports returns true if path is imported by pkg.
func imports(pkg *types.Package, path string) bool {
	for _, imp := range pkg.Imports() {
		if imp.Path() == path {
			return true
		}
	}
	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Fast path: if the package doesn't import net/http,
	// skip the traversal.
	if !imports(pass.Pkg, "net/http") {
		return nil, nil
	}

	// Grab out the control flow graphs.
	cfgs := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)

	onlyFuncs := []ast.Node{
		(*ast.FuncDecl)(nil),
	}
	inspect.Preorder(onlyFuncs, func(n ast.Node) {
		fnDecl := n.(*ast.FuncDecl)

		if !responseWriterInParams(pass, fnDecl) {
			return
		}

		// Great, we've identified that the function takes http.ResponseWriter as an argument!
		// Let's inspect its control flow graph.
		acfg := cfgs.FuncDecl(fnDecl)

		for _, block := range acfg.Blocks {
			if retStmt := block.Return(); retStmt != nil {
				continue
			}
			firstHTTPErrorIndex := -1
			for i, node := range block.Nodes {
				exprStmt, ok := node.(*ast.ExprStmt)
				if !ok {
					continue
				}
				callExpr, ok := exprStmt.X.(*ast.CallExpr)
				if !ok {
					continue
				}

				var ident *ast.Ident
				switch t := callExpr.Fun.(type) {
				case *ast.Ident:
					ident = t
				case *ast.SelectorExpr:
					ident = t.Sel
				case *ast.CallExpr:
					ident = t.Fun.(*ast.Ident)
				}

				if ident != nil && identMatches(pass, ident, responseWriterRepliers...) {
					firstHTTPErrorIndex = i
					break
				}
			}

			// No invocation of http.Error in this block, can safely continue.
			if firstHTTPErrorIndex == -1 {
				continue
			}

			var succs []*cfg.Block
			tillEndOfBlock := block.Nodes[firstHTTPErrorIndex+1:]
			// First attempt is to try to find any terminating statements in the same block as the
			// http.Error statement, for example:
			//  if cond {
			//      http.Error(rw, msg, code)
			//      ...
			//      ...
			//      panic("panicking here")
			//  }
			for _, node := range tillEndOfBlock {
				if isTerminatingStmt(pass, node) {
					goto done
				}
			}

			// Check if the successors don't have terminating statements.
			succs = append(succs, block.Succs...)
			for len(succs) > 0 {
				succ := succs[0]
				succs = succs[1:]

				for _, node := range succ.Nodes {
					switch node.(type) {
					case *ast.ReturnStmt:
						goto done
					case *ast.DeferStmt:
						continue
					default:
						if isTerminatingStmt(pass, node) {
							goto done
						}
						goto failed
					}
				}
				succs = append(succs, succ.Succs...)
			}

		failed:
			// We did not find a terminating statement in this block.
			pass.ReportRangef(block.Nodes[firstHTTPErrorIndex], "call to http.Error without a terminating statement below it")
		done:
		}
	})
	return nil, nil
}

var responseWriterRepliers = []string{"http.Error", "http.NotFound"}

// Check that the function arguments contain:
//      http.ResponseWriter
func responseWriterInParams(pass *analysis.Pass, fnDecl *ast.FuncDecl) bool {
	params := pass.TypesInfo.Defs[fnDecl.Name].Type().(*types.Signature).Params()
	for i, n := 0, params.Len(); i < n; i++ {
		cur := params.At(i)
		if cur.Type().String() == "net/http.ResponseWriter" {
			return true
		}
	}
	return false
}

func isTerminatingStmt(pass *analysis.Pass, n ast.Node) bool {
	if n == nil {
		return false
	}
	switch t := n.(type) {
	default:
		return false
	case *ast.ReturnStmt:
		return true
	case *ast.CallExpr:
		return isPanicOrKnownExitFunc(pass, t)
	case *ast.ExprStmt:
		callExpr, ok := t.X.(*ast.CallExpr)
		return ok && isPanicOrKnownExitFunc(pass, callExpr)
	}
}

func isPanicOrKnownExitFunc(pass *analysis.Pass, callExpr *ast.CallExpr) bool {
	var ident *ast.Ident
	if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		ident = selExpr.Sel
	} else {
		ident = callExpr.Fun.(*ast.Ident)
	}
	return identMatches(pass, ident, "builtin.panic", "runtime.Goexit", "log.Fatal", "log.Fatalf")
}

func identMatches(pass *analysis.Pass, ident *ast.Ident, anyOfFullNames ...string) bool {
	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}
	pkgName := "builtin"
	objName := obj.Name()
	if pkg := obj.Pkg(); pkg != nil {
		pkgName = pkg.Name()
	}
	identFullName := pkgName + "." + objName
	for _, fullName := range anyOfFullNames {
		if fullName == identFullName {
			return true
		}
	}
	return false
}
