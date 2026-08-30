package parser

import "github.com/unmango/go-make/ast"

// ParsePartial parses a file and reports the tree the parser built along with
// any errors it recorded. It exists so tests can assert on the trees error
// recovery produces, which ParseFile discards as soon as an error is recorded.
func (p *Parser) ParsePartial() (*ast.File, error) {
	f := p.parseFile()
	if p.errors.Len() > 0 {
		p.errors.Sort()
		return f, p.errors.Err()
	}

	return f, nil
}
