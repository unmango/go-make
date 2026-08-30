package rule

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/obj"
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
	return obj.Copy(pos, r).(*ast.Rule)
}

// SetPos lays r out beginning at pos, assigning every position the printer
// reads when it writes r. It returns [End] of the moved rule.
//
// Targets and prerequisites are separated by a single space, the colon
// immediately follows the last target, and a single space separates the colon
// from the first prerequisite.
//
// A rule is an [ast.Obj], and a conditional in its recipe list holds objects
// of every kind, so the layout of the whole tree lives in [obj] and this
// forwards to it.
func SetPos(pos token.Pos, r *ast.Rule) token.Pos {
	return obj.SetPos(pos, r)
}

// End returns the position of the first character following the newline that
// terminates r.
func End(r *ast.Rule) token.Pos {
	return obj.End(r)
}
