package printer

import (
	"fmt"
	"io"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

type printer struct {
	w       io.Writer
	lastTok token.Token
	pos     token.Pos
	lastPos token.Pos
}

func (p *printer) setPos(pos token.Pos) {
	p.pos = pos
}

func (p *printer) expr(expr ast.Expr) {
	switch n := expr.(type) {
	case *ast.Text:
		io.WriteString(p.w, n.Value)
	}
}

func (p *printer) exprList(l []ast.Expr) {
	for _, e := range l {
		p.expr(e)
	}
}

func (p *printer) recipe(r *ast.Recipe) {
}

func (p *printer) rule(r *ast.Rule) {
	for _, t := range r.Targets.List {
		p.expr(t)
	}
	for _, t := range r.PreReqs.List {
		p.expr(t)
	}
	for _, r := range r.Recipes {
		p.recipe(r)
	}
}

func (p *printer) variable(v *ast.Variable) {
	p.expr(v.Name)
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
	}

	return fmt.Errorf("unsupported node: %v", node)
}
