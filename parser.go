package make

import (
	"fmt"
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

func NewParser(r io.Reader) *Parser {
	return &Parser{
		s:    NewScanner(r),
		file: &token.File{},
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

func (p *Parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.error(pos, "'"+tok.String()+"'")
	}

	p.next()
	return pos
}

func (p *Parser) error(pos token.Pos, msg string) {
	epos := p.file.Position(pos)
	p.errors.Add(epos, msg)
}

func (p *Parser) next() {
	if p.s.Scan() {
		// TODO: p.pos
		p.tok, p.lit = p.s.Token(), p.s.Literal()
		fmt.Println("set tok:", p.tok)
	} else {
		p.tok = token.EOF
	}
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
	var names []ast.FileName
	for p.tok != token.COLON && p.tok != token.EOF {
		names = append(names, p.parseFileName())
	}

	return &ast.Rule{
		Targets: &ast.TargetList{
			List: names,
		},
	}
}

func (p *Parser) parseFileName() ast.FileName {
	return &ast.LiteralFileName{
		Name: p.parseIdent(),
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
