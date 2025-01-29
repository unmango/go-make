package make

import (
	"io"

	"github.com/unmango/go-make/ast"
)

func Fprint(writer io.Writer, node ast.Node) (err error) {
	w := NewWriter(writer)
	switch node := node.(type) {
	case ast.Expr:
		_, err = WriteExpr(w, node)
	case *ast.TargetList:
		_, err = WriteTargetList(w, node)
	case *ast.PreReqList:
		_, err = WritePreReqList(w, node)
	case *ast.Recipe:
		_, err = WriteRecipe(w, node)
	case *ast.Rule:
		_, err = WriteRule(w, node)
	}

	return
}
