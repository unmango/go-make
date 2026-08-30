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

func (p *Parser) isRecipePrefix() bool {
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

// skipNewlines consumes newlines without consuming a recipe prefix.
// Object parsing starts at the first non-newline token so a leading tab
// remains part of the object it introduces.
func (p *Parser) skipNewlines() {
	for p.tok == token.NEWLINE {
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

// parseRef parses a variable reference. An escaped '$$' is not a reference, so
// it returns an ast.Text holding both characters instead of an ast.VarRef.
func (p *Parser) parseRef() ast.Expr {
	if p.tok != token.DOLLAR {
		p.expect(token.DOLLAR)
		return nil
	}

	dollar := p.pos
	p.next()

	open, name := token.ILLEGAL, "_"
	switch p.tok {
	case token.DOLLAR:
		// '$$' escapes a literal '$'. Both characters belong to the value so
		// the printer writes them back unchanged.
		p.next()
		return &ast.Text{
			ValuePos: dollar,
			Value:    "$$",
		}
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
	default:
		// Without this the "_" placeholder would reach the printer.
		p.expectOneOf(token.TEXT, token.DOLLAR, token.LPAREN, token.LBRACE)
		return nil
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
	// The scanner reports everything between the pound and the end of the line,
	// whitespace included. Storing it verbatim keeps "#foo" distinct from
	// "# foo" so the printer can write either one back unchanged.
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
		Tok:     tok,
		TokPos:  pos,
		VarName: arg,
	}
}

func (p *Parser) parseQuotedExpr() *ast.QuotedExpr {
	var (
		quote token.Token
		open  token.Pos
	)

	switch p.tok {
	case token.APOS:
		quote = token.APOS
		open = p.expect(token.APOS)
	case token.QUOTE:
		quote = token.QUOTE
		open = p.expect(token.QUOTE)
	default:
		p.expectOneOf(token.APOS, token.QUOTE)
	}

	value := p.parseText()
	close := p.expect(quote)

	return &ast.QuotedExpr{
		Quote: quote,
		Open:  open,
		Value: value,
		Close: close,
	}
}

func (p *Parser) parseIfeqDir() *ast.IfeqDir {
	pos, tok := p.pos, p.tok
	p.next() // consume ifeq or ifneq

	var (
		lparen, rparen token.Pos
		arg1, arg2     ast.Expr
		comma          token.Pos
	)

	switch p.tok {
	case token.LPAREN:
		lparen = p.expect(token.LPAREN)
		arg1 = p.parseExpression()
		comma = p.expect(token.COMMA)
		arg2 = p.parseExpression()
		rparen = p.expect(token.RPAREN)
	case token.APOS, token.QUOTE:
		arg1 = p.parseQuotedExpr()
		arg2 = p.parseQuotedExpr()
	default:
		p.expectOneOf(token.LPAREN, token.APOS, token.QUOTE)
	}

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

func (p *Parser) parseIfDir() (d ast.IfDir) {
	switch p.tok {
	case token.IFDEF, token.IFNDEF:
		d = p.parseIfdefDir()
	case token.IFEQ, token.IFNEQ:
		d = p.parseIfeqDir()
	}

	return
}

func (p *Parser) parseElseBlock() *ast.ElseBlock {
	pos := p.expect(token.ELSE)
	condition := p.parseIfDir()

	p.skipNewlines()
	text := p.parseObjList()

	return &ast.ElseBlock{
		Else:      pos,
		Condition: condition,
		Text:      text,
	}
}

func (p *Parser) parseIfBlock() *ast.IfBlock {
	ifdir := p.parseIfDir()
	p.skipNewlines()
	text := p.parseObjList()

	var eblocks []*ast.ElseBlock
	for p.tok == token.ELSE {
		b := p.parseElseBlock()
		eblocks = append(eblocks, b)
		p.skipNewlines()

		if b.Condition == nil {
			break
		}
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
		token.IFNDEF_ASSIGN, token.RECURSIVE_ASSIGN, token.SHELL_ASSIGN,
		token.APPEND_ASSIGN:
		if len(l) == 1 {
			return p.parseVar(l[0])
		}
		p.error(p.pos, "variable may have only one name")
		fallthrough
	default:
		return p.parseBadObj(l)
	}
}

// exprText reproduces the source text of an expression parsed from a line
// that turned out to be unsupported.
func exprText(e ast.Expr) string {
	if s, ok := e.(fmt.Stringer); ok {
		return s.String()
	}

	return ""
}

// parseBadObj consumes the remainder of the current line and returns a
// BadObj spanning it, including the expressions already parsed from the
// line. The line is the recovery boundary because make itself is
// line-oriented, so the next line is the first point the parser can be
// confident it is synchronized again.
func (p *Parser) parseBadObj(parsed []ast.Expr) *ast.BadObj {
	from := p.pos
	if len(parsed) > 0 {
		from = parsed[0].Pos()
	}

	b := &strings.Builder{}
	nextPos := from
	write := func(pos token.Pos, text string) {
		for range int(pos - nextPos) {
			b.WriteByte(' ')
		}

		b.WriteString(text)
		nextPos = pos + token.Pos(len(text))
	}

	for _, e := range parsed {
		write(e.Pos(), exprText(e))
	}
	for p.tok != token.NEWLINE && p.tok != token.EOF {
		write(p.pos, p.recipeTokenText())
		p.next()
	}

	return &ast.BadObj{
		From: from,
		To:   nextPos,
		Text: b.String(),
	}
}

func (p *Parser) parseObjList() (l []ast.Obj) {
	for p.tok != token.EOF && p.tok != token.ENDIF && p.tok != token.ELSE {
		l = append(l, p.parseObj())
		p.skipNewlines()
	}

	return
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

// recipeTokenText returns the source text of the current token. Tokens such
// as TEXT and UNSUPPORTED stringify to their name rather than to any source
// text, and the scanner reports the text it read in the literal, so the
// literal wins whenever the token carries one. Operators and directives
// carry no literal and stringify to the syntax they represent.
func (p *Parser) recipeTokenText() string {
	if p.tok == token.COMMENT {
		return "#" + p.lit
	}
	if p.lit != "" {
		return p.lit
	}

	return p.tok.String()
}

func (p *Parser) parseRecipe() *ast.Recipe {
	prefixPos := p.expect(p.recipePrefix)
	prefixText := p.recipePrefix.String()
	prefixWidth := token.Pos(len(prefixText))
	b := &strings.Builder{}
	nextPos := prefixPos + prefixWidth
	for p.tok != token.NEWLINE && p.tok != token.EOF {
		if gap := int(p.pos - nextPos); gap > 0 {
			for range gap {
				b.WriteByte(' ')
			}
		}

		text := p.recipeTokenText()
		b.WriteString(text)
		nextPos = p.pos + token.Pos(len(text))
		p.next()
	}
	if p.tok == token.NEWLINE {
		p.next()
	}

	return &ast.Recipe{
		Prefix:    p.recipePrefix,
		PrefixPos: prefixPos,
		Text: ast.Text{
			Value:    b.String(),
			ValuePos: prefixPos + prefixWidth,
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
		if p.skipNewlines(); p.tok == token.EOF {
			break
		}

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
