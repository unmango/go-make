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

func NewFile(builder ...File) *ast.File {
	var pos token.Pos = 1
	file := &ast.File{}
	for _, fn := range builder {
		fn(pos, file)
		pos = file.End()
	}

	return file
}

func NewRule(pos token.Pos, builder ...Rule) *ast.Rule {
	rule := &ast.Rule{}
	for _, fn := range builder {
		fn(pos, rule)
		pos = rule.End()
	}

	return rule
}

func Flat[T ast.Node](builders []Builder[T]) Builder[T] {
	return func(p token.Pos, t T) {
		for _, build := range builders {
			build(p, t)
		}
	}
}

func RePos(pos token.Pos, node ast.Node) {
	
}

func NoOp[T ast.Node](token.Pos, T) {}
