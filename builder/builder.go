package builder

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
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

func (b *rule) Target(f func(Expr)) {
	b.space()
	e := &expr{builder: b.builder}
	f(e)
	b.r.Targets = append(b.r.Targets, e.e)
}

type file struct {
	*builder
	f *ast.File
}

func (b *file) Rule(t ExprBuilder, fs ...RuleBuilder) {
	e := &expr{builder: b.builder}
	t(e)

	r := &rule{b.builder, ast.NewRule(e.e, 0)}
	for _, f := range fs {
		f(r)
	}

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
