package recipe

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

func Copy(pos token.Pos, r *ast.Recipe) *ast.Recipe {
	return &ast.Recipe{
		Prefix:    r.Prefix,
		PrefixPos: pos,
		Text: ast.Text{
			Value:    r.Value,
			ValuePos: pos + 1,
		},
	}
}
