package expr

import (
	"go/token"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/text"
)

func Copy(pos token.Pos, expr ast.Expr) ast.Expr {
	switch n := expr.(type) {
	case *ast.Text:
		return text.Copy(pos, n)
	default:
		panic("unsupported node type")
	}
}
