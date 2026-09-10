package recipe

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/expr"
	"github.com/unmango/go-make/token"
)

// Copy returns a copy of r positioned at pos.
func Copy(pos token.Pos, r *ast.Recipe) *ast.Recipe {
	return expr.Copy(pos, r).(*ast.Recipe)
}
