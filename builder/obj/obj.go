package obj

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/rule"
	"github.com/unmango/go-make/token"
)

func Copy(pos token.Pos, obj ast.Obj) ast.Obj {
	switch n := obj.(type) {
	case *ast.Rule:
		return rule.Copy(pos, n)
	default:
		panic("unsupported node type")
	}
}
