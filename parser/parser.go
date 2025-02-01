package parser

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/scanner"
	"github.com/unmango/go-make/token"
)

type Parser struct {
	s      *scanner.Scanner
	file   *token.File
	errors scanner.ErrorList

	pos token.Pos
	tok token.Token // one token look-ahead
	lit string      // token literal

	recipePrefix token.Token
}

func New(r io.Reader, file *token.File) *Parser {
	if file == nil {
		file = token.NewFileSet().AddFile("", 1, math.MaxInt-2)
	}

	p := &Parser{
		s:    scanner.New(r, file),
		file: file,

		recipePrefix: token.TAB,
	}
	p.next()

	return p
}

func (p Parser) isRecipePrefix() bool {
	return p.tok == p.recipePrefix
}

func (p *Parser) error(pos token.Pos, msg string) {
	epos := p.file.Position(pos)
	p.errors.Add(epos, msg)
}

func (p *Parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if p.pos == pos {
		switch {
		case p.tok.IsLiteral():
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}

	p.error(pos, msg)
}

func (p *Parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}

	p.next()
	return pos
}

func (p *Parser) expectOneOf(tok ...token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok[0] {
		ts := make([]string, len(tok))
		for i, t := range tok {
			ts[i] = fmt.Sprint("'", t, "'")
		}

		p.errorExpected(pos, "one of "+strings.Join(ts, ", "))
	}

	p.next()
	return pos
}

func (p *Parser) next() {
	p.pos, p.tok, p.lit = p.s.Scan()
}

func (p *Parser) parseText() *ast.Text {
	pos, name := p.pos, "_"
	if p.tok == token.TEXT {
		name = p.lit
		p.next()
	} else {
		p.expect(token.TEXT)
	}

	return &ast.Text{
		ValuePos: pos,
		Value:    name,
	}
}

func (p *Parser) parseRef() *ast.VarRef {
	if p.tok != token.DOLLAR {
		p.expect(token.DOLLAR)
		return nil
	}

	dollar := p.pos
	p.next()

	open, name := token.ILLEGAL, "_"
	switch p.tok {
	case token.LPAREN, token.LBRACE:
		open = p.tok
		p.next()
		if p.tok == token.TEXT {
			name = p.lit
			p.next()
		} else {
			p.expect(token.TEXT)
		}
	case token.TEXT:
		if len(p.lit) == 1 {
			name = p.lit
			p.next()
		} else {
			// TODO: This should occur in the scanner
			name = p.lit[:1]
			p.lit = p.lit[1:]
			p.pos++
		}
	}

	close := token.ILLEGAL
	if open != token.ILLEGAL {
		switch p.tok {
		case token.RPAREN, token.RBRACE:
			close = p.tok
			p.next()
		default:
			p.expectOneOf(token.RPAREN, token.RBRACE)
		}
	}

	return &ast.VarRef{
		Dollar: dollar,
		Open:   open,
		Name:   name,
		Close:  close,
	}
}

func (p *Parser) parseExpression() ast.Expr {
	switch p.tok {
	case token.TEXT:
		return p.parseText()
	case token.DOLLAR:
		return p.parseRef()
	default:
		p.expectOneOf(token.TEXT, token.DOLLAR)
		return nil
	}
}

func (p *Parser) parseDecl() ast.Decl {
	var l []ast.Expr
	for p.tok == token.TEXT || p.tok == token.DOLLAR {
		l = append(l, p.parseExpression())
	}

	switch p.tok {
	case token.COLON:
		return p.parseRule(l)
	case token.SIMPLE_ASSIGN, token.POSIX_ASSIGN, token.IMMEDIATE_ASSIGN,
		token.IFNDEF_ASSIGN, token.RECURSIVE_ASSIGN, token.SHELL_ASSIGN:
		if len(l) == 1 {
			return p.parseVar(l[0])
		}
		p.error(p.pos, "variable may have only one name")
		fallthrough
	default:
		p.next() // always progress
		return nil
	}
}

func (p *Parser) parseVar(name ast.Expr) ast.Decl {
	op, opPos := p.tok, p.pos
	p.next()

	var rhs []ast.Expr
	for p.tok != token.NEWLINE && p.tok != token.EOF {
		rhs = append(rhs, p.parseExpression())
	}

	return &ast.Variable{
		Name:  name,
		Op:    op,
		OpPos: opPos,
		Value: rhs,
	}
}

func (p *Parser) parseRecipe() *ast.Recipe {
	if !p.isRecipePrefix() {
		p.expect(p.recipePrefix)
		return nil
	}

	prefixPos := p.pos
	b := &strings.Builder{}
	p.next()
	for p.tok != token.NEWLINE && p.tok != token.EOF {
		if p.pos > prefixPos+1 {
			b.WriteRune(' ')
		}
		b.WriteString(p.lit)
		p.next()
	}
	if p.tok == token.NEWLINE {
		p.next()
	}

	return &ast.Recipe{
		Prefix:    token.TAB,
		PrefixPos: prefixPos,
		Text: ast.Text{
			Value:    b.String(),
			ValuePos: prefixPos + 1,
		},
	}
}

func (p *Parser) parseRule(targets []ast.Expr) *ast.Rule {
	var colon token.Pos
	if p.tok == token.COLON {
		colon = p.pos
		p.next()
	} else {
		p.expect(token.COLON)
	}
	if p.errors.Len() > 0 {
		return nil
	}

	prereqs := []ast.Expr{}
	for p.tok != token.PIPE && p.tok != token.NEWLINE && p.tok != token.EOF {
		prereqs = append(prereqs, p.parseExpression())
	}
	if p.errors.Len() > 0 {
		return nil
	}

	pipe, oprereqs := token.NoPos, []ast.Expr{}
	if p.tok == token.PIPE {
		pipe = p.pos
		p.next()
		for p.tok != token.NEWLINE && p.tok != token.EOF {
			oprereqs = append(oprereqs, p.parseExpression())
		}
	}
	if p.tok == token.NEWLINE {
		p.next()
	}

	recipes := make([]*ast.Recipe, 0)
	for p.isRecipePrefix() && p.tok != token.EOF {
		recipes = append(recipes, p.parseRecipe())
	}
	if p.errors.Len() > 0 {
		return nil
	}

	return &ast.Rule{
		Targets:      targets,
		Colon:        colon,
		PreReqs:      prereqs,
		Pipe:         pipe,
		OrderPreReqs: oprereqs,
		Recipes:      recipes,
	}
}

func (p *Parser) parseFile() *ast.File {
	if p.errors.Len() > 0 {
		return nil
	}

	var decls []ast.Decl
	for p.tok != token.EOF {
		decls = append(decls, p.parseDecl())
	}

	return &ast.File{
		Comments:  []*ast.CommentGroup{},
		Decls:     decls,
		FileStart: token.Pos(p.file.Base()),
		FileEnd:   token.Pos(p.file.Base() + p.file.Size()),
	}
}

func (p *Parser) ParseFile() (*ast.File, error) {
	f := p.parseFile()
	if p.errors.Len() > 0 {
		p.errors.Sort()
		return nil, p.errors.Err()
	} else {
		return f, nil
	}
}
