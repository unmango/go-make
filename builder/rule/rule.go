package rule

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/expr"
	"github.com/unmango/go-make/builder/recipe"
	"github.com/unmango/go-make/builder/text"
	"github.com/unmango/go-make/token"
)

func New(pos token.Pos, builder ...builder.Rule) *ast.Rule {
	rule := &ast.Rule{}
	for _, fn := range builder {
		fn(pos, rule)
	}

	if n := len(rule.Targets); n > 0 {
		rule.Colon = rule.Targets[n-1].End()
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

func Copy(pos token.Pos, r *ast.Rule) *ast.Rule {
	rule := &ast.Rule{}
	for _, t := range r.Targets {
		t = expr.Copy(pos, t)
		rule.Targets = append(rule.Targets, t)
		pos = t.End() + 1
	}

	if len(r.Targets) > 0 {
		rule.Colon = pos - 1
	} else {
		rule.Colon = pos
	}
	pos += 2

	for _, p := range r.PreReqs {
		p = expr.Copy(pos, p)
		rule.PreReqs = append(rule.PreReqs, p)
		pos = p.End() + 1
	}

	if r.Pipe.IsValid() && len(r.OrderPreReqs) > 0 {
		rule.Pipe = pos + 1
		for _, p := range r.OrderPreReqs {
			p = expr.Copy(pos, p)
			rule.OrderPreReqs = append(rule.OrderPreReqs, p)
			pos = p.End()
		}
	}

	if len(r.Recipes) > 0 {
		for _, r := range r.Recipes {
			r = recipe.Copy(pos, r)
			rule.Recipes = append(rule.Recipes, r)
			pos = r.End()
		}
	}

	return rule
}
