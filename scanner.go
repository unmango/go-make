package make

import (
	"bufio"
	"bytes"
	"io"
	"math"

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
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanTokens)

	if file == nil {
		file = token.NewFileSet().AddFile("", 1, math.MaxInt-2)
	}
	s := &Scanner{
		s:    scanner,
		file: file,
	}
	s.next()

	return s
}

func (s Scanner) Err() error {
	return s.s.Err()
}

func (s Scanner) Position(pos token.Pos) token.Position {
	return token.PositionFor(s.file, pos)
}

func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	if s.done {
		pos = s.file.Pos(s.offset)
		tok = token.EOF
		return
	}

	s.skipWhitespace()

	// current token start
	pos = s.file.Pos(s.offset)
	var atNewline bool

	switch txt := s.s.Text(); {
	case token.IsIdentifier(txt):
		lit = txt
		s.next()
		if len(txt) > 1 {
			tok = token.Lookup(txt)
		} else {
			tok = token.IDENT
		}
	default:
		s.next()
		switch txt {
		case "=":
			tok = token.RECURSIVE_ASSIGN
		case ":=":
			tok = token.SIMPLE_ASSIGN
		case "::=":
			tok = token.POSIX_ASSIGN
		case ":::=":
			tok = token.IMMEDIATE_ASSIGN
		case "?=":
			tok = token.IFNDEF_ASSIGN
		case "!=":
			tok = token.SHELL_ASSIGN
		case ",":
			tok = token.COMMA
		case "\n":
			atNewline = true
			tok = token.NEWLINE
		case "\t":
			tok = token.TAB
		case "(":
			tok = token.LPAREN
		case ")":
			tok = token.RPAREN
		case "{":
			tok = token.LBRACE
		case "}":
			tok = token.RBRACE
		case "$":
			tok = token.DOLLAR
		case ":":
			tok = token.COLON
		case ";":
			tok = token.SEMI
		case "|":
			tok = token.PIPE
		case "#":
			// TODO
			// s.lit = s.scanComment()
			tok = token.COMMENT
		default:
			tok = token.UNSUPPORTED
			s.lit = txt
		}
	}

	if atNewline && s.done {
		tok = token.EOF
	}

	return
}

func (s *Scanner) next() {
	s.offset = s.rdOffset
	if bytes.ContainsRune(s.s.Bytes(), '\n') {
		s.file.AddLine(s.offset)
	}
	s.done = !s.s.Scan()
	s.rdOffset += len(s.s.Bytes())
}

func (s *Scanner) skipWhitespace() {
	for bytes.ContainsAny(s.s.Bytes(), " \r") {
		s.next()
	}
}
