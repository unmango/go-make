package expr

import (
	"go/token"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/text"
)

// func Builder(expr ast.Expr) builder.Expr {
// 	return func(p token.Pos, e ast.Expr) {
// 		switch n := expr.(type) {
// 		case *ast.Text:
// 			text.Value()
// 		case *ast.VarRef:
// 			e.VarRef(n.Name)
// 		}
// 	}
// }

func RePos(pos token.Pos, expr ast.Expr) {
	switch n := expr.(type) {
	case *ast.Text:
		text.RePos(pos, n)
	}
}
