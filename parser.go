package make

import (
	"go/scanner"
	"io"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

type Parser struct {
	s      *Scanner
	file   *token.File
	errors scanner.ErrorList

	pos token.Pos
	tok token.Token // one token look-ahead
	lit string      // token literal
}

func NewParser(r io.Reader, file *token.File) *Parser {
	p := &Parser{
		s:    NewScanner(r, file),
		file: file,
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

	var rules []*ast.Rule
	for p.tok != token.EOF {
		rules = append(rules, p.parseRule())
	}

	return &ast.File{
		Comments:  []*ast.CommentGroup{},
		Rules:     rules,
		FileStart: token.Pos(p.file.Base()),
		FileEnd:   token.Pos(p.file.Base() + p.file.Size()),
	}
}

func (p *Parser) parseRule() *ast.Rule {
	targets := new(ast.TargetList)
	for p.tok != token.COLON && p.tok != token.EOF {
		targets.Add(p.parseFileName())
	}
	if p.errors.Len() > 0 {
		return nil
	}

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

	return &ast.Rule{
		Targets: targets,
		Colon:   colon,
		Pipe:    token.NoPos,
		Semi:    token.NoPos,
		PreReqs: nil,
		Recipes: nil,
	}
}

func (p *Parser) parseFileName() ast.FileName {
	name := p.parseIdent()

	return &ast.LiteralFileName{
		Name: name,
	}
}

func (p *Parser) parseIdent() *ast.Ident {
	pos, name := p.pos, "_"
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
	} else {
		p.expect(token.IDENT)
	}

	return &ast.Ident{
		NamePos: pos,
		Name:    name,
	}
}
