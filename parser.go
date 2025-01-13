package make

import (
	"io"

	"github.com/unmango/go-make/token"
)

type Parser struct {
	s *Scanner

	tok token.Token // one token look-ahead
	lit string      // token literal
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		s: NewScanner(r),
	}
}
