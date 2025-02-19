package text

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/token"
)

func New(pos token.Pos, builder ...builder.Text) *ast.Text {
	text := &ast.Text{ValuePos: pos}
	for _, fn := range builder {
		fn(pos, text)
	}

	return text
}

func Value(text string) builder.Text {
	return func(p token.Pos, t *ast.Text) {
		t.Value = text
		t.ValuePos = p
	}
}

func RePos(pos token.Pos, text *ast.Text) {
	text.ValuePos = pos
}
