package builder

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/build"
	"github.com/unmango/go-make/token"
)

func Noop[T any](T) {}

type (
	File = build.File
	Rule = build.Rule
	Expr = build.Expr
)

type builder struct {
	pos token.Pos
}

func (b *builder) nextPos() token.Pos {
	pos := b.pos
	b.pos++
	return pos
}

func (b *builder) nextStr(s string) token.Pos {
	pos := b.pos
	b.pos += token.Pos(len(s))
	return pos
}

func (b *builder) space() {
	_ = b.nextPos()
}

func (b *builder) text(t string) *ast.Text {
	return &ast.Text{
		Value:    t,
		ValuePos: b.nextStr(t),
	}
}

func (b *builder) varRef(name string) *ast.VarRef {
	dollar := b.nextPos()
	_ = b.nextPos() // Open
	_ = b.nextStr(name)
	_ = b.nextPos() // Close

	return &ast.VarRef{
		Dollar: dollar,
		Open:   token.LBRACE,
		Name:   name,
		Close:  token.RBRACE,
	}
}
