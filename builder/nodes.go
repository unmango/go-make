package builder

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/build"
	"github.com/unmango/go-make/token"
)

type expr struct {
	*builder
	e ast.Expr
}

func (b *expr) VarRef(name string) {
	b.e = b.varRef(name)
}

func (b *expr) Text(t string) {
	b.e = b.text(t)
}

type rule struct {
	*builder
	r *ast.Rule
}

func (b *rule) Target(f func(build.Expr)) {
	if b.r != nil && len(b.r.Targets) > 0 {
		b.space()
	}

	e := &expr{builder: b.builder}
	f(e)
	b.r.Targets = append(b.r.Targets, e.e)
}

type file struct {
	*builder
	f *ast.File
}

func (b *file) Rule(t func(build.Expr), rs ...func(build.Rule)) {
	b.f.Contents = append(b.f.Contents,
		newRule(b.builder, t, rs),
	)
}

func (b *file) Start(pos token.Pos) {
	b.f.FileStart = pos
}

func newExpr(b *builder, f func(build.Expr)) ast.Expr {
	e := &expr{builder: b}
	f(e)
	return e.e
}

func NewExpr(start token.Pos, f func(build.Expr)) ast.Expr {
	return newExpr(&builder{start}, f)
}

func newRule(b *builder, e func(build.Expr), rs []func(build.Rule)) *ast.Rule {
	r := &rule{b, &ast.Rule{
		Targets: []ast.Expr{newExpr(b, e)},
	}}
	for _, f := range rs {
		f(r)
	}

	r.r.Colon = r.nextPos()
	return r.r
}

func NewRule(start token.Pos, e func(build.Expr), rs ...func(build.Rule)) *ast.Rule {
	return newRule(&builder{start}, e, rs)
}

func NewFile(start token.Pos, fs ...func(build.File)) *ast.File {
	b := &file{&builder{start}, &ast.File{FileStart: start}}
	for _, f := range fs {
		f(b)
	}

	b.f.FileEnd = b.pos
	return b.f
}
