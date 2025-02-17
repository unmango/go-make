package internal

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

type Builder[T ast.Node] func(*State) T

type (
	File Builder[*ast.File]
	Rule Builder[*ast.Rule]
	Expr Builder[ast.Expr]
)

type State struct {
	pos token.Pos
}

func (b *State) NextPos() token.Pos {
	pos := b.pos
	b.pos++
	return pos
}

func (b *State) NextStr(s string) token.Pos {
	pos := b.pos
	b.pos += token.Pos(len(s))
	return pos
}

func (b *State) Space() {
	_ = b.NextPos()
}

func (b *State) Text(t string) *ast.Text {
	return &ast.Text{
		Value:    t,
		ValuePos: b.NextStr(t),
	}
}

func (b *State) VarRef(name string) *ast.VarRef {
	dollar := b.NextPos()
	_ = b.NextPos() // Open
	_ = b.NextStr(name)
	_ = b.NextPos() // Close

	return &ast.VarRef{
		Dollar: dollar,
		Open:   token.LBRACE,
		Name:   name,
		Close:  token.RBRACE,
	}
}
