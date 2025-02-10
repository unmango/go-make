package builder

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

type expr struct {
	*builder
	e ast.Expr
}

func (e *expr) VarRef(name string) {
	dollar := e.nextPos()
	_ = e.nextPos() // Open
	_ = e.nextStr(name)
	_ = e.nextPos() // Close

	e.e = &ast.VarRef{
		Dollar: dollar,
		Open:   token.LBRACE,
		Name:   name,
		Close:  token.RBRACE,
	}
}

func (e *expr) Text(t string) {
	e.e = &ast.Text{
		Value:    t,
		ValuePos: e.nextStr(t),
	}
}

type rule struct {
	*builder
	r *ast.Rule
}

func (b *rule) Target(f ExprBuilder) {
	b.space()
	e := &expr{builder: b.builder}
	f(e)
	b.r.Targets = append(b.r.Targets, e.e)
}

type file struct {
	*builder
	f *ast.File
}

func (b *file) Rule(t ExprBuilder, rs ...RuleBuilder) {
	b.f.Contents = append(b.f.Contents,
		newRule(b.builder, t, rs),
	)
}

func (b *file) Start(pos token.Pos) {
	b.f.FileStart = pos
}

func newExpr(b *builder, f ExprBuilder) ast.Expr {
	e := &expr{builder: b}
	f(e)
	return e.e
}

func NewExpr(start token.Pos, f ExprBuilder) ast.Expr {
	return newExpr(&builder{start}, f)
}

func newRule(b *builder, e ExprBuilder, rs []RuleBuilder) *ast.Rule {
	r := &rule{b, &ast.Rule{
		Targets: []ast.Expr{newExpr(b, e)},
	}}
	for _, f := range rs {
		f(r)
	}

	r.r.Colon = r.nextPos()
	return r.r
}

func NewRule(start token.Pos, e ExprBuilder, rs ...RuleBuilder) *ast.Rule {
	return newRule(&builder{start}, e, rs)
}

func NewFile(start token.Pos, fs ...FileBuilder) *ast.File {
	b := &file{&builder{start}, &ast.File{FileStart: start}}
	for _, f := range fs {
		f(b)
	}

	b.f.FileEnd = b.pos
	return b.f
}
