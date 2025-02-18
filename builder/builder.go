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
	File   = Builder[*ast.File]
	Rule   = Builder[*ast.Rule]
	Expr   = Builder[ast.Expr]
	Text   = Builder[*ast.Text]
	VarRef = Builder[*ast.VarRef]
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

func FileInsertRule(i int, builder ...Rule) File {
	return func(p token.Pos, f *ast.File) {
		// Simply re-write the entire contents starting at p?
		for j, o := range f.Contents {
			switch {
			case j < i:
				continue
			case j == i:
			}
		}
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

func RuleTarget(fn func(token.Pos) ast.Expr) Rule {
	return func(p token.Pos, r *ast.Rule) {
		r.Targets = append(r.Targets, fn(p))
	}
}

func RuleTargetText(name string) Rule {
	return RuleTarget(func(p token.Pos) ast.Expr {
		return NewText(p, TextValue(name))
	})
}

func NewText(pos token.Pos, builder ...Text) *ast.Text {
	text := &ast.Text{ValuePos: pos}
	for _, fn := range builder {
		fn(pos, text)
	}

	return text
}

func TextValue(text string) Text {
	return func(p token.Pos, t *ast.Text) {
		t.Value = text
	}
}

// func NewText(text string) Text {
// 	return func(p token.Pos, t *ast.Text) {
// 		t.Value = text
// 		t.ValuePos = p
// 	}
// }

// func NewVarRef(name string) VarRef {
// 	return func(p token.Pos, v *ast.VarRef) {
// 		v.Dollar = p
// 		v.Open = token.LPAREN
// 		v.Name = name
// 		v.Close = token.RPAREN
// 	}
// }
