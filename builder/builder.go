package builder

import (
	"go/token"

	"github.com/unmango/go-make/ast"
)

type builder struct {
	pos token.Pos
}

func (b *builder) nextPos() token.Pos {
	pos := b.pos
	b.pos++
	return pos
}

func (b *builder) nextStr(s string) token.Pos {
	pos := b.pos
	b.pos += token.Pos(len(s))
	return pos
}

func (b *builder) space() {
	_ = b.nextPos()
}

func (b *builder) text(t string) *ast.Text {
	return &ast.Text{
		Value:    t,
		ValuePos: b.nextStr(t),
	}
}

type File interface {
	Start(token.Pos)
	Rule(string, func(Rule))
}

type Rule interface {
	Target(string)
	TargetExpr(func(Expr))
}

type Expr interface {
	VarRef(string)
}

type expr struct {
	*builder
	e ast.Expr
}

func (e *expr) VarRef(name string) {
	e.e = &ast.VarRef{
		Dollar: e.nextPos(),
	}
}

type rule struct {
	*builder
	r *ast.Rule
}

func (b *rule) Target(t string) {
	b.space()
	b.r.Targets = append(b.r.Targets, b.text(t))
}

func (b *rule) TargetExpr(f func(Expr)) {
	e := &expr{builder: b.builder}
	f(e)
	b.space()
	b.r.Targets = append(b.r.Targets, e.e)
}

type file struct {
	*builder
	f *ast.File
}

func (b *file) Rule(target string, f func(Rule)) {
	r := &rule{b.builder, ast.NewRule(b.text(target), 0)}
	f(r)
	r.r.Colon = r.nextPos()
	b.f.Contents = append(b.f.Contents, r.r)
}

func (b *file) Start(pos token.Pos) {
	b.f.FileStart = pos
}

func NewFile(start token.Pos, f func(File)) *ast.File {
	b := &file{&builder{1}, &ast.File{FileStart: start}}
	f(b)
	b.f.FileEnd = b.pos
	return b.f
}

func Noop[T any](T) {}
