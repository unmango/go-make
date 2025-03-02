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

// func Copy(expr ast.Expr) builder.Expr {
// 	switch n := expr.(type) {
// 	case *ast.Text:
// 		return upcast(text.Copy(n))
// 	default:
// 		return builder.NoOp
// 	}
// }

// func Copy(expr ast.Expr) func(token.Pos) ast.Expr {
// 	return func(p token.Pos) ast.Expr {
// 		switch n := expr.(type) {
// 		case *ast.Text:
// 			return text.Copy(p, n)
// 		default:
// 			panic("unsupported node type")
// 		}
// 	}
// }

func Copy(pos token.Pos, expr ast.Expr) ast.Expr {
	switch n := expr.(type) {
	case *ast.Text:
		return text.Copy(pos, n)
	default:
		panic("unsupported node type")
	}
}

func RePos(pos token.Pos, expr ast.Expr) {
	switch n := expr.(type) {
	case *ast.Text:
		text.RePos(pos, n)
	}
}
