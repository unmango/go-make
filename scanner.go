package make

import (
	"bufio"
	"bytes"
	"io"

	"github.com/unmango/go-make/token"
)

type Scanner struct {
	file *token.File
	s    *bufio.Scanner

	offset   int
	rdOffset int
	tok      token.Token
	lit      string

	done bool
}

func NewScanner(r io.Reader, file *token.File) *Scanner {
	s := &Scanner{
		s:    bufio.NewScanner(r),
		file: file,
	}
	s.s.Split(ScanTokens)
	s.next()

	return s
}

func (s Scanner) Err() error {
	return s.s.Err()
}

func (s Scanner) Token() token.Token {
	return s.tok
}

func (s Scanner) Literal() string {
	return s.lit
}

func (s Scanner) Pos() token.Pos {
	return s.file.Pos(s.offset)
}

func (s *Scanner) Scan() bool {
	if s.done {
		s.tok = token.EOF
		return false
	}

	var atNewline bool

	s.skipWhitespace()
	switch txt := s.s.Text(); {
	case token.IsIdentifier(txt):
		s.lit = txt
		if len(txt) > 1 {
			s.tok = token.Lookup(txt)
		} else {
			s.tok = token.IDENT
		}
	default:
		switch txt {
		case "=":
			s.tok = token.RECURSIVE_ASSIGN
		case ":=":
			s.tok = token.SIMPLE_ASSIGN
		case "::=":
			s.tok = token.POSIX_ASSIGN
		case ":::=":
			s.tok = token.IMMEDIATE_ASSIGN
		case "?=":
			s.tok = token.IFNDEF_ASSIGN
		case "!=":
			s.tok = token.SHELL_ASSIGN
		case ",":
			s.tok = token.COMMA
		case "\n":
			atNewline = true
			s.tok = token.NEWLINE
		case "\t":
			s.tok = token.TAB
		case "(":
			s.tok = token.LPAREN
		case ")":
			s.tok = token.RPAREN
		case "{":
			s.tok = token.LBRACE
		case "}":
			s.tok = token.RBRACE
		case "$":
			s.tok = token.DOLLAR
		case ":":
			s.tok = token.COLON
		case "#":
			// TODO
			// s.lit = s.scanComment()
			s.tok = token.COMMENT
		default:
			s.tok = token.UNSUPPORTED
			s.lit = txt
		}
	}

	s.next()
	if atNewline && s.done {
		s.tok = token.EOF
		return false
	} else {
		return true
	}
}

func (s *Scanner) next() {
	s.done = !s.s.Scan()
	s.offset = s.rdOffset
	s.rdOffset += len(s.s.Bytes())
}

func (s *Scanner) skipWhitespace() {
	for bytes.ContainsAny(s.s.Bytes(), " \r") {
		s.next()
	}
}
