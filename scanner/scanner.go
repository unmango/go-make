package scanner

import (
	"bufio"
	"bytes"
	"go/scanner"
	"io"
	"math"
	"strings"

	"github.com/unmango/go-make/token"
)

type ErrorList = scanner.ErrorList

// space is the single byte token produced by [ScanTokens] for a space.
var space = []byte{' '}

type Scanner struct {
	file *token.File
	s    *bufio.Scanner

	offset   int
	rdOffset int

	lf, crlf int

	done bool
}

func New(r io.Reader, file *token.File) *Scanner {
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

func (s *Scanner) next() {
	s.offset = s.rdOffset
	if bytes.ContainsRune(s.s.Bytes(), '\n') {
		s.file.AddLine(s.offset)
	}
	s.done = !s.s.Scan()
	s.rdOffset += len(s.s.Bytes())
}

// skipWhitespace consumes space tokens. The comparison is exact so that
// a token which merely contains a space is never discarded along with the
// text around it.
func (s *Scanner) skipWhitespace() {
	for bytes.Equal(s.s.Bytes(), space) {
		s.next()
	}
}

func (s *Scanner) scanComment() string {
	b := strings.Builder{}
	for !s.done && !bytes.ContainsRune(s.s.Bytes(), '\n') {
		b.Write(s.s.Bytes())
		s.next()
	}

	return b.String()
}

// LineEnding reports the line ending used by the input scanned so far, "\n"
// or "\r\n". A file may mix the two, so the ending that terminates the
// majority of its lines wins and a tie resolves to "\n". Input with no line
// ending at all reports "\n".
//
// The answer is only complete once [Scanner.Scan] has reported [token.EOF],
// because every newline before that point is still unread.
func (s *Scanner) LineEnding() string {
	if s.crlf > s.lf {
		return "\r\n"
	}

	return "\n"
}

func (s *Scanner) Err() error {
	return s.s.Err()
}

func (s *Scanner) Position(pos token.Pos) token.Position {
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

	// Trailing whitespace exhausted the input.
	if s.done {
		tok = token.EOF
		return
	}

	// newline holds the text of the line ending when the token is one, so the
	// tally below can tell CRLF from LF after the switch has run.
	var newline string

	switch txt := s.s.Text(); {
	case txt == "\r\n":
		// token.IsLit reports true for a carriage return, so CRLF is
		// handled before the literal case.
		newline = txt
		tok = token.NEWLINE
		lit = txt
		s.next()
	case txt == "\r":
		// A carriage return that does not terminate a line has no meaning
		// in make syntax, report it rather than dropping it.
		tok = token.UNSUPPORTED
		lit = txt
		s.next()
	case token.IsLit(txt):
		lit = txt
		s.next()
		if len(txt) > 1 {
			tok = token.Lookup(txt)
		} else {
			tok = token.TEXT
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
		case "+=":
			tok = token.APPEND_ASSIGN
		case ",":
			tok = token.COMMA
		case "'":
			tok = token.APOS
		case `"`:
			tok = token.QUOTE
		case "\n":
			newline = txt
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
			lit = s.scanComment()
			tok = token.COMMENT
		default:
			tok = token.UNSUPPORTED
			lit = txt
		}
	}

	if newline != "" {
		// The newline that ends the final line is folded into EOF below, so
		// the tally happens first and counts every line ending in the input.
		if newline == "\r\n" {
			s.crlf++
		} else {
			s.lf++
		}
		if s.done {
			tok, lit = token.EOF, ""
		}
	}

	return
}
