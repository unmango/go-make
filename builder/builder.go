package builder

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

type Builder[T ast.Node] = func(token.Pos, T) token.Pos

type (
	File   = Builder[*ast.File]
	Rule   = Builder[*ast.Rule]
	Expr   = Builder[ast.Expr]
	Text   = Builder[*ast.Text]
	VarRef = Builder[*ast.VarRef]
)

func Flat[T ast.Node](builders []Builder[T]) Builder[T] {
	return func(p token.Pos, t T) token.Pos {
		for _, build := range builders {
			p = build(p, t)
		}

		return p
	}
}

func NoOp[T ast.Node](pos token.Pos, _ T) token.Pos {
	return pos
}
