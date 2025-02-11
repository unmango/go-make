package builder

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

func AtRule(node *ast.Rule, f RuleFunc) *ast.Rule {
	if node == nil {
		return nil
	}

	var pos token.Pos = 1
	if len(node.Targets) > 0 {
		pos = node.Pos()
	}

	r := &rule{&builder{pos}, node}
	f(r)
	r.r.Colon = r.pos

	return r.r
}
