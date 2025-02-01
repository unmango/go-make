package make

import (
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

func NewParser(r io.Reader, file *token.File) *Parser {
	if file == nil {
		file = token.NewFileSet().AddFile("", 1, math.MaxInt-2)
	}

	p := &Parser{
		s:    scanner.NewScanner(r, file),
		file: file,

		recipePrefix: token.TAB,
	}
	p.next()

	return p
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

func (p *Parser) next() {
	p.pos, p.tok, p.lit = p.s.Scan()
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

func (p *Parser) parseDecl() ast.Decl {
	var l []ast.Expr
	for p.tok == token.TEXT {
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
	for p.tok != token.NEWLINE && p.tok != token.EOF {
		prereqs = append(prereqs, p.parseExpression())
	}
	if p.errors.Len() > 0 {
		return nil
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
		Targets: targets,
		Colon:   colon,
		Pipe:    token.NoPos,
		PreReqs: prereqs,
		Recipes: recipes,
	}
}

func (p Parser) isRecipePrefix() bool {
	return p.tok == p.recipePrefix
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

func (p *Parser) parseExpression() ast.Expr {
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
