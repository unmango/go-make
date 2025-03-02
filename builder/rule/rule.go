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

	if n := len(rule.Targets); n > 0 {
		rule.Colon = rule.Targets[n-1].End() + 1
	} else {
		rule.Colon = pos
	}

	return rule
}

func PreReq(expr func(token.Pos) ast.Expr) builder.Rule {
	return func(p token.Pos, r *ast.Rule) {
		r.PreReqs = append(r.PreReqs, expr(p))
	}
}

func TextPreReq(value string) builder.Rule {
	return PreReq(func(p token.Pos) ast.Expr {
		return text.New(p, text.Value(value))
	})
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

func Copy(r *ast.Rule) builder.Rule {
	builders := []builder.Rule{}
	for _, t := range r.Targets {
		builders = append(builders, Target(expr.Copy(t)))
	}

	return builder.Flat(builders)
}

func RePos(pos token.Pos, rule *ast.Rule) {
	for _, t := range rule.Targets {
		expr.RePos(pos, t)
		pos = t.End() + 1
	}

	rule.Colon = pos
	pos += 2

	for _, p := range rule.PreReqs {
		expr.RePos(pos, p)
		pos = p.End()
	}
}
