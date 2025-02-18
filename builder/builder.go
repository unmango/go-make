package builder

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

func Noop[T any](T) {}

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

func (b *builder) varRef(name string) *ast.VarRef {
	dollar := b.nextPos()
	_ = b.nextPos() // Open
	_ = b.nextStr(name)
	_ = b.nextPos() // Close

	return &ast.VarRef{
		Dollar: dollar,
		Open:   token.LBRACE,
		Name:   name,
		Close:  token.RBRACE,
	}
}

type State[T ast.Node] struct {
	pos  token.Pos
	Node T
}

// Advance returns the current position and
// increments the position by n
func (s *State[T]) Advance(n int) token.Pos {
	pos := s.pos
	s.pos += token.Pos(n)
	return pos
}

// Increment returns the current position
// and increments the position by 1
func (s *State[T]) Increment() token.Pos {
	return s.Advance(1)
}

type Builder[T ast.Node] = func(token.Pos, T)

type (
	File = Builder[*ast.File]
	Rule = Builder[*ast.Rule]
	Expr = Builder[ast.Expr]
)

func NewFile2(builder ...File) *ast.File {
	var pos token.Pos = 1
	file := &ast.File{}
	for _, fn := range builder {
		fn(pos, file)
		pos = file.End()
	}

	return file
}

func FileRule(builder ...Rule) File {
	return func(p token.Pos, f *ast.File) {
		f.Contents = append(f.Contents, NewRule2(p, builder...))
	}
}

func NewRule2(pos token.Pos, builder ...Rule) *ast.Rule {
	rule := &ast.Rule{}
	for _, fn := range builder {
		fn(pos, rule)
		pos = rule.End()
	}

	return rule
}

func RuleTargetText(name string) Rule {
	return func(p token.Pos, r *ast.Rule) {
		r.Targets = append(r.Targets, nil)
	}
}
