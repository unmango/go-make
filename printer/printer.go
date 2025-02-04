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

func (p *printer) fill(c byte, pos token.Pos) {
	p.writeChar(c, int(pos)-(p.pos.Offset+1))
}

func (p *printer) fillSpace(pos token.Pos) {
	p.fill(' ', pos)
}

func (p *printer) fillLines(pos token.Pos) {
	p.fill('\n', pos)
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

func (p *printer) quotedExpr(e *ast.QuotedExpr) {
	p.tok(p.posFor(e.Open), e.Quote)
	p.expr(e.Value)
	p.tok(p.posFor(e.Close), e.Quote)
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
	case *ast.QuotedExpr:
		p.quotedExpr(n)
	case *ast.Recipe:
		p.text(&n.Text)
	case *ast.VarRef:
		p.varRef(n)
	}
}

func (p *printer) exprList(l []ast.Expr) {
	for _, e := range l {
		p.fillSpace(e.Pos())
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
	p.fillSpace(r.Colon)
	p.tok(p.posFor(r.Colon), token.COLON)
	p.prereqList(r.PreReqs)
	if r.Pipe.IsValid() {
		p.fillSpace(r.Pipe)
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

func (p *printer) comment(c *ast.Comment) {
	p.writeString(p.posFor(c.Pound), "#")
	p.fillSpace(c.Pound + 2)
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

func (p *printer) ifeqDir(d *ast.IfeqDir) {
	p.tok(p.posFor(d.TokPos), d.Tok)
	if d.Open.IsValid() {
		p.fillSpace(d.Open)
		p.tok(p.posFor(d.Open), token.LPAREN)
		p.fillSpace(d.Arg1.Pos())
		p.expr(d.Arg1)
		p.fillSpace(d.Comma)
		p.tok(p.posFor(d.Comma), token.COMMA)
		p.fillSpace(d.Arg2.Pos())
		p.expr(d.Arg2)
		p.fillSpace(d.Close)
		p.tok(p.posFor(d.Close), token.RPAREN)
	} else {
		p.fillSpace(d.Arg1.Pos())
		p.expr(d.Arg1)
		p.fillSpace(d.Arg2.Pos())
		p.expr(d.Arg2)
	}
}

func (p *printer) ifdefDir(d *ast.IfdefDir) {
	p.tok(p.posFor(d.TokPos), d.Tok)
	p.fillSpace(d.VarName.Pos())
	p.expr(d.VarName)
}

func (p *printer) ifDir(d ast.IfDir) {
	switch n := d.(type) {
	case *ast.IfeqDir:
		p.ifeqDir(n)
	case *ast.IfdefDir:
		p.ifdefDir(n)
	}
}

func (p *printer) elseBlock(b *ast.ElseBlock) {
	p.tok(p.posFor(b.Else), token.ELSE)
	if b.Condition != nil {
		p.fillSpace(b.Condition.Pos())
		p.ifDir(b.Condition)
	}
	p.writeLine()
	p.objList(b.Text)
}

func (p *printer) ifBlock(b *ast.IfBlock) {
	p.ifDir(b.Directive)
	p.objList(b.Text)
	for _, e := range b.Else {
		p.elseBlock(e)
	}
	p.tok(p.posFor(b.Endif), token.ENDIF)
}

func (p *printer) directive(d ast.Dir) {
	switch n := d.(type) {
	case *ast.IfBlock:
		p.ifBlock(n)
	}
}

func (p *printer) variable(v *ast.Variable) {
	if v == nil {
		return
	}

	p.expr(v.Name)
	p.fillSpace(v.OpPos)
	p.tok(p.posFor(v.OpPos), v.Op)
	if v.Value != nil {
		p.exprList(v.Value)
	}
	p.writeLine()
}

func (p *printer) obj(o ast.Obj) {
	switch n := o.(type) {
	case ast.Dir:
		p.directive(n)
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
		p.fillLines(d.Pos())
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
	case ast.IfDir:
		p.ifDir(n)
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
