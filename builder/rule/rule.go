package rule

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/expr"
	"github.com/unmango/go-make/builder/text"
	"github.com/unmango/go-make/token"
)

func New(pos token.Pos, builder ...builder.Rule) *ast.Rule {
	rule := &ast.Rule{}
	for _, fn := range builder {
		fn(pos, rule)
	}

	return rule
}

func Target(expr func(token.Pos) ast.Expr) builder.Rule {
	return func(p token.Pos, r *ast.Rule) {
		r.Targets = append(r.Targets, expr(p))
	}
}

func TextTarget(value string) builder.Rule {
	return Target(func(p token.Pos) ast.Expr {
		return text.New(p, text.Value(value))
	})
}

// func Copy(rule *ast.Rule) builder.Rule {
// 	return func(p token.Pos, r *ast.Rule) {
// 		for _, t := range rule.Targets {

// 		}
// 	}
// }

func Flat(builders []builder.Rule) builder.Rule {
	return func(p token.Pos, r *ast.Rule) {
		for _, b := range builders {
			b(p, r)
		}
	}
}

func RePos(pos token.Pos, rule *ast.Rule) {
	for _, t := range rule.Targets {
		expr.RePos(pos, t)
		pos = t.End() + 1
	}

	rule.Colon = pos + 1
	pos += 2

	for _, p := range rule.PreReqs {
		expr.RePos(pos, p)
		pos = p.End()
	}
}
