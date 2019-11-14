package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
)

type Snippet struct {
	Code     string
	Filename string
	Line     int
}

func (s Snippet) String() string {
	return fmt.Sprintf("%s:%d '%s'", s.Filename, s.Line, s.Code)
}

type Router struct {
	Snippet
	Routes []*Route
}

type Route struct {
	Snippet
	Path string
}

func NewSnippet(fset *token.FileSet, n ast.Node) Snippet {
	var b bytes.Buffer
	printer.Fprint(&b, fset, n)
	pos := fset.Position(n.Pos())
	return Snippet{
		Code:     b.String(),
		Filename: pos.Filename,
		Line:     pos.Line,
	}
}

type analyser struct {
	fileset       *token.FileSet
	info          *types.Info
	Routers       []*Router
	OutGoingCalls []*OutGoingCall
}

func NewAnalyser(fset *token.FileSet) *analyser {
	return &analyser{
		fileset: fset,
	}
}

func (a *analyser) Run(f *ast.File) {
	ast.Walk(a, f)
}

func (a *analyser) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.SelectorExpr:
		if c := detectHTTPCalls(n); c != nil {
			a.OutGoingCalls = append(a.OutGoingCalls, c)
		}
		return a
	case *ast.FuncDecl:
		a.collectHTTPHandlerAsFuncParams(n)
		return a
	case *ast.BlockStmt:
		ast.Inspect(n, func(root ast.Node) bool {
			switch r := root.(type) {
			case *ast.AssignStmt:
				a.collectHTTPHandlerAssignments(n, r)
				return true
			}
			return true
		})
	}

	return a
}

type OutGoingCall struct {
	code     string
	kind     string
	node     ast.Node
	pos      token.Pos
	position token.Position
}

func (c *OutGoingCall) String() string {
	return fmt.Sprintf("http call at line %s:%d (%s)", c.position.Filename, c.position.Line, c.kind)
}

func (a *analyser) collectHTTPHandlerAsFuncParams(f *ast.FuncDecl) {
	if f.Type != nil {
		if params := f.Type.Params; params != nil {
			for _, field := range params.List {
				switch expr := field.Type.(type) {
				case *ast.StarExpr:
					switch sel := expr.X.(type) {
					case *ast.SelectorExpr:
						if isSelectorExpr(sel, "mux", "Router") || isSelectorExpr(sel, "http", "ServeMux") {
							a.Routers = append(a.Routers, &Router{
								Snippet: NewSnippet(a.fileset, f.Name),
								Routes:  a.detectHTTPRoutes(f.Body, field.Names...),
							})
						}
					}
				}
			}
		}
	}
}

func (a *analyser) detectHTTPRoutes(b *ast.BlockStmt, names ...*ast.Ident) (routes []*Route) {
	var objectName string
	if len(names) > 0 {
		objectName = names[0].Name
	} else {
		return
	}

	ast.Inspect(b, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			switch sel := node.Fun.(type) {
			case *ast.SelectorExpr:
				if extractIdent(sel.X) == objectName && (sel.Sel.Name == "Handle" || sel.Sel.Name == "HandleFunc") {
					var url string
					switch val := node.Args[0].(type) {
					case *ast.BasicLit:
						url = val.Value
					}
					route := &Route{
						Snippet: NewSnippet(a.fileset, n),
						Path:    url,
					}
					routes = append(routes, route)
				}
			}
		}
		return true
	})
	return routes
}

func detectHTTPCalls(sel *ast.SelectorExpr) *OutGoingCall {
	if isSelectorExpr(sel, "http", "Get") {
		return &OutGoingCall{kind: "http.Get", pos: sel.Pos()}
	}
	if isSelectorExpr(sel, "http", "Head") {
		return &OutGoingCall{kind: "http.Head", pos: sel.Pos()}
	}
	if isSelectorExpr(sel, "http", "Post") {
		return &OutGoingCall{kind: "http.Post", pos: sel.Pos()}
	}

	return nil
}

func (a *analyser) collectHTTPHandlerAssignments(block *ast.BlockStmt, assign *ast.AssignStmt) {
	switch rhs := assign.Rhs[0].(type) {
	case *ast.CallExpr:
		switch sel := rhs.Fun.(type) {
		case *ast.SelectorExpr:
			if isSelectorExpr(sel, "mux", "NewRouter") || isSelectorExpr(sel, "http", "NewServeMux") {
				a.Routers = append(a.Routers, &Router{
					Snippet: NewSnippet(a.fileset, assign),
					Routes:  a.detectHTTPRoutes(block, assign.Lhs[0].(*ast.Ident)),
				})
			}
		}
	}
}

func isSelectorExpr(sel *ast.SelectorExpr, exprIdent, selIdent string) bool {
	return extractIdent(sel.X) == exprIdent && sel.Sel.Name == selIdent
}

func extractIdent(n ast.Node) string {
	switch ident := n.(type) {
	case *ast.Ident:
		return ident.Name
	}
	return ""
}
