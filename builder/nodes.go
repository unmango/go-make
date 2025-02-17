package builder

import (
	"slices"

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

func (b *rule) Target(fn func(build.Expr), fs ...func(build.Expr)) {
	if b.r != nil && len(b.r.Targets) > 0 {
		b.space()
	}

	e := &expr{builder: b.builder}
	fn(e)
	for _, f := range fs {
		f(e)
	}
	b.r.Targets = append(b.r.Targets, e.e)
}

type file struct {
	*builder
	f *ast.File
}

func (b *file) insert(i int, obj ast.Obj) {
	b.f.Contents = slices.Insert(b.f.Contents, i, obj)
	for i := i; i < len(b.f.Contents); i++ {
		switch n := b.f.Contents[i].(type) {
		case *ast.Rule:
			b.f.Contents[i] = ApplyRule(n, func(r build.Rule) {})
		}
	}
}

func (b *file) AddRule(fn func(build.Rule), fs ...func(build.Rule)) {
	var o ast.Obj = newRule(b.builder, fn, fs)
	b.f.Contents = append(b.f.Contents, o)
}

func (b *file) InsertRule(i int, fn func(build.Rule), fs ...func(build.Rule)) {
	b.insert(i, newRule(b.builder, fn, fs))
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

func newRule(b *builder, fn func(build.Rule), fs []func(build.Rule)) *ast.Rule {
	r := &rule{b, &ast.Rule{}}
	fn(r)
	for _, f := range fs {
		f(r)
	}

	r.r.Colon = r.nextPos()
	return r.r
}

func NewRule(start token.Pos, fn func(build.Rule), fs ...func(build.Rule)) *ast.Rule {
	return newRule(&builder{start}, fn, fs)
}

func NewFile(start token.Pos, fs ...func(build.File)) *ast.File {
	b := &file{&builder{start}, &ast.File{FileStart: start}}
	for _, f := range fs {
		f(b)
	}

	b.f.FileEnd = b.pos
	return b.f
}
