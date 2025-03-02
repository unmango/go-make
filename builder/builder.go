package builder

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

type Builder[T ast.Node] = func(token.Pos, T)

type (
	File   = Builder[*ast.File]
	Rule   = Builder[*ast.Rule]
	Expr   = Builder[ast.Expr]
	Text   = Builder[*ast.Text]
	VarRef = Builder[*ast.VarRef]
)

func Flat[T ast.Node](builders []Builder[T]) Builder[T] {
	return func(p token.Pos, t T) {
		for _, build := range builders {
			build(p, t)
		}
	}
}

func NoOp[T ast.Node](token.Pos, T) {}
