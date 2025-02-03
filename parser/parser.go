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

func (p *Parser) isWhitespace() bool {
	return p.tok == token.NEWLINE || p.tok == token.TAB
}

func (p *Parser) skipWhitespace() {
	for p.tok != token.EOF && p.isWhitespace() {
		p.next()
	}
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

func (p *Parser) parseComment() *ast.Comment {
	pos, lit := p.pos, p.lit
	p.next()

	return &ast.Comment{
		Pound: pos,
		Text:  lit,
	}
}

func (p *Parser) parseCommentGroup() *ast.CommentGroup {
	g := &ast.CommentGroup{}
	for p.tok == token.COMMENT {
		g.List = append(g.List, p.parseComment())
		if p.tok == token.NEWLINE {
			p.next() // potentially more comments in group
		}
	}

	return g
}

func (p *Parser) parseIfdefDir() *ast.IfdefDir {
	pos, tok := p.pos, p.tok
	p.next()
	arg := p.parseExpression()

	return &ast.IfdefDir{
		Tok:    tok,
		TokPos: pos,
		Arg:    arg,
	}
}

func (p *Parser) parseIfeqDir() *ast.IfeqDir {
	pos, tok := p.pos, p.tok
	p.next()
	lparen := p.expect(token.LPAREN)
	arg1 := p.parseExpression()
	comma := p.expect(token.COMMA)
	arg2 := p.parseExpression()
	rparen := p.expect(token.RPAREN)

	return &ast.IfeqDir{
		Tok:    tok,
		TokPos: pos,
		Open:   lparen,
		Arg1:   arg1,
		Comma:  comma,
		Arg2:   arg2,
		Close:  rparen,
	}
}

func (p *Parser) parseElseBlock() *ast.ElseBlock {
	pos := p.expect(token.ELSE)

	return &ast.ElseBlock{
		Else: pos,
	}
}

func (p *Parser) parseIfBlock() *ast.IfBlock {
	var ifdir ast.IfDir
	switch p.tok {
	case token.IFDEF, token.IFNDEF:
		ifdir = p.parseIfdefDir()
	case token.IFEQ, token.IFNEQ:
		ifdir = p.parseIfeqDir()
	}
	p.skipWhitespace()

	var text []ast.Obj
	for p.tok != token.EOF && p.tok != token.ENDIF && p.tok != token.ELSE {
		text = append(text, p.parseObj())
		p.skipWhitespace()
	}

	var eblocks []*ast.ElseBlock
	if p.tok == token.ELSE {
		eblocks = append(eblocks, p.parseElseBlock())
		p.skipWhitespace()
	}

	endif := p.expect(token.ENDIF)

	return &ast.IfBlock{
		Directive: ifdir,
		Text:      text,
		Else:      eblocks,
		Endif:     endif,
	}
}

func (p *Parser) parseObj() ast.Obj {
	switch p.tok {
	case token.COMMENT:
		return p.parseCommentGroup()
	case token.IFDEF, token.IFNDEF, token.IFEQ, token.IFNEQ:
		return p.parseIfBlock()
	}

	// TODO: refactor to improve the error message
	// we expect one expression, then we expect one
	// of (Expr | COLON | *_ASSIGN)
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
		p.next()   // always progress
		return nil // TODO: BadObj?
	}
}

func (p *Parser) parseVar(name ast.Expr) ast.Obj {
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
	prefixPos := p.expect(p.recipePrefix)
	b := &strings.Builder{}
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
	colon := p.expect(token.COLON)
	prereqs := []ast.Expr{}
	for p.tok != token.PIPE && p.tok != token.NEWLINE && p.tok != token.EOF {
		prereqs = append(prereqs, p.parseExpression())
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
	var content []ast.Obj
	for p.tok != token.EOF {
		p.skipWhitespace()
		content = append(content, p.parseObj())
	}

	return &ast.File{
		Contents:  content,
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
