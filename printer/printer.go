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
}

type Op func(*printer)

func WithFile(f *token.File) Op {
	return func(p *printer) {
		p.f = f
	}
}

func (p *printer) posFor(pos token.Pos) token.Position {
	return token.PositionFor(p.f, pos)
}

func (p *printer) writeLine() {
	p.out = append(p.out, '\n')
	p.pos.Line++
	p.pos.Offset++
}

func fillSpace(p *printer, pos token.Pos) {
	p.writeSpace(int(pos) - (p.pos.Offset + 1))
}

func (p *printer) writeChar(r byte, n int) {
	for range n {
		p.out = append(p.out, r)
	}

	p.pos.Offset += n
	p.pos.Column += n
}

func (p *printer) writeSpace(n int) {
	p.writeChar(' ', n)
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

func (p *printer) recipe(r *ast.Recipe) {
	pos := p.posFor(r.PrefixPos)
	p.tok(pos, r.Prefix)
	p.expr(r)
	p.writeLine()
}

func (p *printer) varRef(v *ast.VarRef) {
	p.tok(p.posFor(v.Dollar), token.DOLLAR)
	if v.Open != token.ILLEGAL {
		p.tok(p.pos, v.Open)
	}
	p.writeString(p.pos, v.Name)
	if v.Close != token.ILLEGAL {
		p.tok(p.pos, v.Close)
	}
}

func (p *printer) expr(expr ast.Expr) {
	switch n := expr.(type) {
	case *ast.Text:
		p.text(n)
	case *ast.Recipe:
		p.text(&n.Text)
	case *ast.VarRef:
		p.varRef(n)
	}
}

func (p *printer) exprList(l []ast.Expr) {
	for _, e := range l {
		fillSpace(p, e.Pos())
		p.expr(e)
	}
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

func (p *printer) recipeList(l []*ast.Recipe) {
	for _, r := range l {
		p.recipe(r)
	}
}

func (p *printer) rule(r *ast.Rule) {
	if r == nil {
		return
	}

	p.targetList(r.Targets)
	fillSpace(p, r.Colon)
	p.tok(p.posFor(r.Colon), token.COLON)
	p.prereqList(r.PreReqs)
	if r.Pipe.IsValid() {
		fillSpace(p, r.Pipe)
		p.tok(p.posFor(r.Pipe), token.PIPE)
	}
	if len(r.OrderPreReqs) > 0 {
		p.exprList(r.OrderPreReqs)
	}
	if len(r.Recipes) > 0 {
		if r.Recipes[0].Prefix != token.SEMI {
			p.writeLine()
		}
		p.recipeList(r.Recipes)
	} else {
		p.writeLine()
	}
}

func (p *printer) variable(v *ast.Variable) {
	if v == nil {
		return
	}

	p.expr(v.Name)
	fillSpace(p, v.OpPos)
	p.tok(p.posFor(v.OpPos), v.Op)
	if v.Value != nil {
		p.exprList(v.Value)
	}
	p.writeLine()
}

func (p *printer) comment(c *ast.Comment) {
	p.writeString(p.posFor(c.Pound), "#")
	fillSpace(p, c.Pound+2)
	p.writeString(p.pos, c.Text)
}

func (p *printer) commentGroup(g *ast.CommentGroup) {
	if g == nil {
		return
	}

	for _, c := range g.List {
		p.comment(c)
		p.writeLine()
	}
}

func (p *printer) obj(o ast.Obj) {
	switch n := o.(type) {
	case *ast.CommentGroup:
		p.commentGroup(n)
	case *ast.Rule:
		p.rule(n)
	case *ast.Variable:
		p.variable(n)
	}
}

func (p *printer) objList(l []ast.Obj) {
	for _, d := range l {
		p.obj(d)
	}
}

func (p *printer) file(f *ast.File) {
	if f != nil {
		p.objList(f.Contents)
	}
}

func (p *printer) printNode(node any) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case ast.Expr:
		p.expr(n)
	case ast.Obj:
		p.obj(n)
	case []ast.Expr:
		p.exprList(n)
	case []ast.Obj:
		p.objList(n)
	case *ast.File:
		p.file(n)
	default:
		return fmt.Errorf("unsupported node: %#v", node)
	}

	return nil
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
