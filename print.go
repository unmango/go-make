package make

import (
	"io"

	"github.com/unmango/go-make/ast"
)

func Fprint(writer io.Writer, node ast.Node) (err error) {
	w := NewWriter(writer)
	switch node := node.(type) {
	case ast.Expr:
		_, err = w.WriteExpr(node)
	case *ast.Recipe:
		_, err = WriteRecipe(w, node)
	case *ast.Rule:
		_, err = WriteRule(w, node)
	case *ast.File:
		_, err = WriteFile(w, node)
	}

	return
}
