package printer

import (
	"fmt"
	"io"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
	"github.com/unmango/go/option"
)

type printer struct {
	f   *token.File
	out []byte
	pos token.Position
	err error
}

type Op func(*printer)

func WithFile(f *token.File) Op {
	return func(p *printer) {
		p.f = f
	}
}

func (p *printer) setPos(pos token.Pos) {
	if pos.IsValid() {
		p.pos = p.posFor(pos)
	}
}

func (p *printer) posFor(pos token.Pos) token.Position {
	return token.PositionFor(p.f, pos)
}

func (p *printer) error(msg string, a ...any) {
	p.err = fmt.Errorf(msg, a...)
}

func (p *printer) writeLine() {
	p.out = append(p.out, '\n')
	p.pos.Line++
	p.pos.Offset++
}

func fillSpace(p *printer, pos token.Pos) {
	p.writeSpace(int(pos) - (p.pos.Offset + 1))
}

func (p *printer) writeSpace(n int) {
	for range n {
		p.out = append(p.out, ' ')
	}

	p.pos.Offset += n
	p.pos.Column += n
}

func (p *printer) writeString(pos token.Position, s string) {
	if pos.IsValid() {
		p.pos = pos
	}

	p.out = append(p.out, s...)
	p.pos.Offset += len(s)
	p.pos.Column += len(s)
}

func (p *printer) tok(pos token.Position, t token.Token) {
	p.writeString(pos, t.String())
}

func (p *printer) text(t *ast.Text) {
	pos := p.posFor(t.Pos())
	p.writeString(pos, t.Value)
}

func (p *printer) expr(expr ast.Expr) {
	switch n := expr.(type) {
	case *ast.Text:
		p.text(n)
	}
}

func (p *printer) exprList(l []ast.Expr) {
	for _, e := range l {
		fillSpace(p, e.Pos())
		p.expr(e)
	}
}

func (p *printer) recipe(r *ast.Recipe) {
	pos := p.posFor(r.PrefixPos)
	p.writeString(pos, r.Prefix.String())
	p.writeString(pos, r.Text)
}

func (p *printer) targetList(l []ast.Expr) {
	if l != nil {
		p.exprList(l)
	}
}

func (p *printer) prereqList(l []ast.Expr) {
	if l != nil {
		p.exprList(l)
	}
}

func (p *printer) rule(r *ast.Rule) {
	if r.Targets == nil {
		p.error("no targets in rule")
		return
	}

	p.targetList(r.Targets)
	fillSpace(p, r.Colon)
	p.tok(p.posFor(r.Colon), token.COLON)
	p.prereqList(r.PreReqs)
	if len(r.Recipes) > 0 && r.Recipes[0].Prefix != token.SEMI {
		p.writeLine()
	}
	for _, r := range r.Recipes {
		p.recipe(r)
	}
	if len(r.Recipes) > 0 {
		p.writeLine()
	}
}

func (p *printer) variable(v *ast.Variable) {
	p.expr(v.Name)
	p.tok(p.posFor(v.OpPos), v.Op)
	if v.Value == nil {
		return
	}
	for _, x := range v.Value {
		p.expr(x)
	}
}

func (p *printer) decl(decl ast.Decl) {
	switch n := decl.(type) {
	case *ast.Rule:
		p.rule(n)
	case *ast.Variable:
		p.variable(n)
	}
}

func (p *printer) declList(l []ast.Decl) {
	for _, d := range l {
		p.decl(d)
	}
}

func (p *printer) file(file *ast.File) {
	p.declList(file.Decls)
}

func (p *printer) printNode(node any) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case ast.Expr:
		p.expr(n)
	case ast.Decl:
		p.decl(n)
	case []ast.Expr:
		p.exprList(n)
	case []ast.Decl:
		p.declList(n)
	case *ast.File:
		p.file(n)
	default:
		return fmt.Errorf("unsupported node: %#v", node)
	}

	return p.err
}

func Fprint(w io.Writer, node any, opts ...Op) (n int, err error) {
	p := &printer{f: &token.File{}}
	option.ApplyAll(p, opts)

	if err = p.printNode(node); err != nil {
		return
	} else {
		return w.Write(p.out)
	}
}
