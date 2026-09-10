package text

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/token"
)

func New(pos token.Pos, builder ...builder.Text) *ast.Text {
	text := &ast.Text{ValuePos: pos}
	for _, fn := range builder {
		pos = fn(pos, text)
	}

	return text
}

func Value(text string) builder.Text {
	return func(p token.Pos, t *ast.Text) token.Pos {
		t.Value = text
		t.ValuePos = p
		return t.End()
	}
}

func Copy(pos token.Pos, text *ast.Text) *ast.Text {
	return &ast.Text{
		Value:    text.Value,
		ValuePos: pos,
	}
}
