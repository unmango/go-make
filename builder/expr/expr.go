package expr

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/build"
)

func Text(text string) func(build.Expr) {
	return func(b build.Expr) {
		b.Text(text)
	}
}

func VarRef(name string) func(build.Expr) {
	return func(b build.Expr) {
		b.VarRef(name)
	}
}

func Builder(expr ast.Expr) func(build.Expr) {
	return func(e build.Expr) {
		switch n := expr.(type) {
		case *ast.Text:
			e.Text(n.Value)
		case *ast.VarRef:
			e.VarRef(n.Name)
		}
	}
}
