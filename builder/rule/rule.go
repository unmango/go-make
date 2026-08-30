package rule

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/expr"
	"github.com/unmango/go-make/builder/text"
	"github.com/unmango/go-make/token"
)

// New returns a rule laid out at pos.
func New(pos token.Pos, builder ...builder.Rule) *ast.Rule {
	rule := &ast.Rule{}
	Build(pos, rule, builder...)
	return rule
}

// Build applies builder to rule and lays the result out at pos, returning
// [End] of the built rule.
//
// Each builder is given the position returned by the builder before it, and
// [SetPos] assigns the final layout once every builder has run. The colon
// separating targets from prerequisites is only known after the last target
// has been built, so the positions a builder sees are provisional.
func Build(pos token.Pos, rule *ast.Rule, builder ...builder.Rule) token.Pos {
	p := pos
	for _, fn := range builder {
		p = fn(p, rule) + 1 // separator
	}

	return SetPos(pos, rule)
}

func PreReq(expr func(token.Pos) ast.Expr) builder.Rule {
	return func(pos token.Pos, r *ast.Rule) token.Pos {
		p := expr(pos)
		r.PreReqs = append(r.PreReqs, p)
		return p.End()
	}
}

func TextPreReq(value string) builder.Rule {
	return PreReq(func(p token.Pos) ast.Expr {
		return text.New(p, text.Value(value))
	})
}

func Target(expr func(token.Pos) ast.Expr) builder.Rule {
	return func(p token.Pos, r *ast.Rule) token.Pos {
		t := expr(p)
		r.Targets = append(r.Targets, t)
		return t.End()
	}
}

func TextTarget(value string) builder.Rule {
	return Target(func(p token.Pos) ast.Expr {
		return text.New(p, text.Value(value))
	})
}

// Copy returns a deep copy of r laid out at pos.
func Copy(pos token.Pos, r *ast.Rule) *ast.Rule {
	c := clone(r)
	SetPos(pos, c)
	return c
}

// SetPos lays r out beginning at pos, assigning every position the printer
// reads when it writes r. It returns [End] of the moved rule.
//
// Targets and prerequisites are separated by a single space, the colon
// immediately follows the last target, and a single space separates the colon
// from the first prerequisite.
func SetPos(pos token.Pos, r *ast.Rule) token.Pos {
	for _, t := range r.Targets {
		pos = expr.SetPos(pos, t) + 1 // ' '
	}

	if len(r.Targets) > 0 {
		r.Colon = pos - 1 // the colon follows the last target directly
	} else {
		r.Colon = pos
	}

	pos = r.Colon + 2 // ':' and ' '
	for _, p := range r.PreReqs {
		pos = expr.SetPos(pos, p) + 1 // ' '
	}

	if r.Pipe.IsValid() {
		r.Pipe = pos
		pos = r.Pipe + 2 // '|' and ' '
	}
	for _, p := range r.OrderPreReqs {
		pos = expr.SetPos(pos, p) + 1 // ' '
	}

	for i, recipe := range r.Recipes {
		if i == 0 && recipe.Prefix == token.SEMI {
			pos-- // a semicolon recipe follows the rule on the same line
		}

		pos = expr.SetPos(pos, recipe) + 1 // '\n'
	}

	return pos
}

// End returns the position of the first character following the newline that
// terminates r.
func End(r *ast.Rule) token.Pos {
	if n := len(r.Recipes); n > 0 {
		return expr.End(r.Recipes[n-1]) + 1
	}
	if n := len(r.OrderPreReqs); n > 0 {
		return expr.End(r.OrderPreReqs[n-1]) + 1
	}
	if n := len(r.PreReqs); n > 0 {
		return expr.End(r.PreReqs[n-1]) + 1
	}
	if r.Pipe.IsValid() {
		return r.Pipe + 2
	}

	return r.Colon + 2
}

func clone(r *ast.Rule) *ast.Rule {
	c := &ast.Rule{Colon: r.Colon, Pipe: r.Pipe}
	for _, t := range r.Targets {
		c.Targets = append(c.Targets, expr.Copy(t.Pos(), t))
	}
	for _, p := range r.PreReqs {
		c.PreReqs = append(c.PreReqs, expr.Copy(p.Pos(), p))
	}
	for _, p := range r.OrderPreReqs {
		c.OrderPreReqs = append(c.OrderPreReqs, expr.Copy(p.Pos(), p))
	}
	for _, recipe := range r.Recipes {
		c.Recipes = append(c.Recipes, expr.Copy(recipe.Pos(), recipe).(*ast.Recipe))
	}

	return c
}
