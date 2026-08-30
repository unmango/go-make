package scanner_test

import (
	"bytes"
	gotoken "go/token"
	"math"
	"strings"
	"testing/quick"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/internal/testing"
	"github.com/unmango/go-make/scanner"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Scanner", func() {
	var file *token.File

	BeforeEach(func() {
		file = gotoken.NewFileSet().AddFile("test", 1, math.MaxInt-2)
	})

	Describe("Position", func() {
		It("should be equivalent to calling file.PositionFor(p, false)", func() {
			s := scanner.New(&bytes.Buffer{}, file)

			err := quick.Check(func(p int) bool {
				pos := token.Pos(p)

				expected := file.PositionFor(pos, false)
				actual := s.Position(pos)

				return actual == expected
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	DescribeTable("Scan identifier",
		Entry(nil, "ident"),
		Entry(nil, "./file/path"),
		Entry(nil, "/abs/path"),
		Entry(nil, "foo_bar"),
		Entry(nil, "foo-bar"),
		Entry(nil, "foo123"),
		Entry(nil, "123"),
		Entry(nil, "_foo"),
		Entry(nil, "foo_"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			pos, tok, lit := s.Scan()
			Expect(tok).To(Equal(token.TEXT))
			Expect(lit).To(Equal(input))
			Expect(pos).To(Equal(token.Pos(1)))

			pos, tok, lit = s.Scan()
			Expect(tok).To(Equal(token.EOF))
			Expect(pos).To(Equal(token.Pos(len(input) + 1)))
		},
	)

	DescribeTable("Scan identifier with trailing newline",
		Entry(nil, "ident\n"),
		Entry(nil, "./file/path\n"),
		Entry(nil, "/abs/path\n"),
		Entry(nil, "foo_bar\n"),
		Entry(nil, "foo-bar\n"),
		Entry(nil, "foo123\n"),
		Entry(nil, "123\n"),
		Entry(nil, "_foo\n"),
		Entry(nil, "foo_\n"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			pos, tok, lit := s.Scan()
			Expect(tok).To(Equal(token.TEXT))
			Expect(lit).To(Equal(strings.TrimSpace(input)))
			Expect(pos).To(Equal(token.Pos(1)))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   0,
				Line:     1,
				Column:   1,
			}))

			pos, tok, lit = s.Scan()
			Expect(tok).To(Equal(token.EOF))
			Expect(pos).To(Equal(token.Pos(len(input))))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   len(input) - 1,
				Line:     1,
				Column:   len(input),
			}))
		},
	)

	DescribeTable("Scan ident followed by whitespace",
		Entry(nil, "ident $"),
		Entry(nil, "ident :"),
		Entry(nil, "ident ;"),
		Entry(nil, "ident |"),
		Entry(nil, "ident ="),
		Entry(nil, "ident :="),
		Entry(nil, "ident ::="),
		Entry(nil, "ident :::="),
		Entry(nil, "ident ?="),
		Entry(nil, "ident !="),
		Entry(nil, "ident +="),
		Entry(nil, "ident ("),
		Entry(nil, "ident )"),
		Entry(nil, "ident {"),
		Entry(nil, "ident }"),
		Entry(nil, "ident ,"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			pos, tok, lit := s.Scan()
			Expect(tok).To(Equal(token.TEXT))
			Expect(lit).To(Equal("ident"))
			Expect(pos).To(Equal(token.Pos(1)))

			pos, tok, lit = s.Scan()
			// File base + Length of the identifier + whitespace
			Expect(pos).To(Equal(token.Pos(7)))
			Expect(tok).NotTo(Equal(token.TEXT))
		},
	)

	DescribeTable("Scan non-ident tokens",
		Entry(nil, "$", token.DOLLAR),
		Entry(nil, ":", token.COLON),
		Entry(nil, ";", token.SEMI),
		Entry(nil, "|", token.PIPE),
		Entry(nil, "=", token.RECURSIVE_ASSIGN),
		Entry(nil, ":=", token.SIMPLE_ASSIGN),
		Entry(nil, "::=", token.POSIX_ASSIGN),
		Entry(nil, ":::=", token.IMMEDIATE_ASSIGN),
		Entry(nil, "?=", token.IFNDEF_ASSIGN),
		Entry(nil, "!=", token.SHELL_ASSIGN),
		Entry(nil, "+=", token.APPEND_ASSIGN),
		Entry(nil, "(", token.LPAREN),
		Entry(nil, ")", token.RPAREN),
		Entry(nil, "{", token.LBRACE),
		Entry(nil, "}", token.RBRACE),
		Entry(nil, ",", token.COMMA),
		Entry(nil, "'", token.APOS),
		Entry(nil, `"`, token.QUOTE),
		Entry(nil, "\t", token.TAB),
		Entry(nil, "\n\n", token.NEWLINE),
		func(input string, expected token.Token) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			pos, tok, _ := s.Scan()
			Expect(tok).To(Equal(expected))
			Expect(pos).To(Equal(token.Pos(1)))

			pos, tok, _ = s.Scan()
			Expect(tok).To(Equal(token.EOF))
		},
	)

	DescribeTable("Scan comment tokens",
		Entry(nil, "#", ""),
		Entry(nil, "#some text", "some text"),
		Entry(nil, "# some text", " some text"),
		Entry(nil, "#\n", ""),
		Entry(nil, "#some text\n", "some text"),
		Entry(nil, "# some text\n", " some text"),
		func(input, expected string) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			pos, tok, lit := s.Scan()

			Expect(tok).To(Equal(token.COMMENT))
			Expect(lit).To(Equal(expected))
			Expect(pos).To(Equal(token.Pos(1)))
		},
	)

	DescribeTable("should scan newline followed by token",
		Entry(nil, "\nident"),
		Entry(nil, "\n,"),
		Entry(nil, "\n'"),
		Entry(nil, "\n\""),
		Entry(nil, "\n$"),
		Entry(nil, "\n;"),
		Entry(nil, "\n:"),
		Entry(nil, "\n:="),
		Entry(nil, "\n::="),
		Entry(nil, "\n:::="),
		Entry(nil, "\n="),
		Entry(nil, "\n?="),
		Entry(nil, "\n!="),
		Entry(nil, "\n|"),
		Entry(nil, "\n\t"),
		Entry(nil, "\n{"),
		Entry(nil, "\n}"),
		Entry(nil, "\n("),
		Entry(nil, "\n)"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			pos, tok, _ := s.Scan()
			Expect(tok).To(Equal(token.NEWLINE))
			Expect(pos).To(Equal(token.Pos(1)))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   0,
				Line:     1,
				Column:   1,
			}))

			pos, tok, _ = s.Scan()
			Expect(pos).To(Equal(token.Pos(2)))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   1,
				Line:     2,
				Column:   1,
			}))
		},
	)

	DescribeTable("space separated tokens",
		Entry(nil, "$ foo", 3),
		Entry(nil, ": foo", 3),
		Entry(nil, "; foo", 3),
		Entry(nil, "| foo", 3),
		Entry(nil, "= foo", 3),
		Entry(nil, ":= foo", 4),
		Entry(nil, "::= foo", 5),
		Entry(nil, ":::= foo", 6),
		Entry(nil, "?= foo", 4),
		Entry(nil, "!= foo", 4),
		Entry(nil, "( foo", 3),
		Entry(nil, ") foo", 3),
		Entry(nil, "{ foo", 3),
		Entry(nil, "} foo", 3),
		Entry(nil, ", foo", 3),
		Entry(nil, "' foo", 3),
		Entry(nil, "\" foo", 3),
		Entry(nil, "\t foo", 3),
		Entry(nil, "identifier foo", 12),
		Entry(nil, "$ foo bar", 3),
		Entry(nil, ": foo bar", 3),
		Entry(nil, "; foo bar", 3),
		Entry(nil, "| foo bar", 3),
		Entry(nil, "= foo bar", 3),
		Entry(nil, ":= foo bar", 4),
		Entry(nil, "::= foo bar", 5),
		Entry(nil, ":::= foo bar", 6),
		Entry(nil, "?= foo bar", 4),
		Entry(nil, "!= foo bar", 4),
		Entry(nil, "( foo bar", 3),
		Entry(nil, ") foo bar", 3),
		Entry(nil, "{ foo bar", 3),
		Entry(nil, "} foo bar", 3),
		Entry(nil, ", foo bar", 3),
		Entry(nil, "\t foo bar", 3),
		Entry(nil, "identifier foo bar", 12),
		func(input string, expected int) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			_, _, _ = s.Scan()
			pos, tok, lit := s.Scan()
			Expect(tok).To(Equal(token.TEXT))
			Expect(pos).To(Equal(token.Pos(expected)))
			Expect(lit).To(Equal("foo"))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   expected - file.Base(),
				Line:     1,
				Column:   expected,
			}))
		},
	)

	DescribeTable("newline separated tokens",
		Entry(nil, "$\nfoo", 2),
		Entry(nil, ":\nfoo", 2),
		Entry(nil, ";\nfoo", 2),
		Entry(nil, "|\nfoo", 2),
		Entry(nil, "=\nfoo", 2),
		Entry(nil, ":=\nfoo", 3),
		Entry(nil, "::=\nfoo", 4),
		Entry(nil, ":::=\nfoo", 5),
		Entry(nil, "?=\nfoo", 3),
		Entry(nil, "!=\nfoo", 3),
		Entry(nil, "(\nfoo", 2),
		Entry(nil, ")\nfoo", 2),
		Entry(nil, "{\nfoo", 2),
		Entry(nil, "}\nfoo", 2),
		Entry(nil, ",\nfoo", 2),
		Entry(nil, "'\nfoo", 2),
		Entry(nil, "\"\nfoo", 2),
		Entry(nil, "\t\nfoo", 2),
		Entry(nil, "identifier\nfoo", 11),
		func(input string, nlPos int) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			_, _, _ = s.Scan()
			pos, tok, _ := s.Scan()
			Expect(tok).To(Equal(token.NEWLINE))
			Expect(pos).To(Equal(token.Pos(nlPos)))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   nlPos - file.Base(),
				Line:     1,
				Column:   nlPos,
			}))

			pos, tok, lit := s.Scan()
			Expect(tok).To(Equal(token.TEXT))
			Expect(pos).To(Equal(token.Pos(nlPos + 1)))
			Expect(lit).To(Equal("foo"))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   nlPos,
				Line:     2,
				Column:   1,
			}))
		},
	)

	DescribeTable("directives",
		Entry(nil, "ifeq"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			pos, tok, lit := s.Scan()
			Expect(tok).To(Equal(token.Lookup(input)))
			Expect(lit).To(Equal(input))
			Expect(pos).To(Equal(token.Pos(1)))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   0,
				Line:     1,
				Column:   1,
			}))
		},
	)

	DescribeTable("Scan trailing whitespace without a newline",
		Entry(nil, "a: b  "),
		Entry(nil, "target: dep "),
		Entry(nil, "ident "),
		Entry(nil, " "),
		Entry(nil, "  "),
		func(input string) {
			buf := bytes.NewBufferString(input)
			s := scanner.New(buf, file)

			var (
				pos token.Pos
				tok token.Token
				lit string
			)
			for tok != token.EOF {
				pos, tok, lit = s.Scan()
				Expect(tok).NotTo(Equal(token.UNSUPPORTED),
					"expected a supported token, got literal %q", lit,
				)
			}

			Expect(pos).To(Equal(token.Pos(len(input) + 1)))
		},
	)

	Describe("CRLF input", func() {
		type scanned struct {
			Tok token.Token
			Lit string
		}

		scanAll := func(input string) []scanned {
			s := scanner.New(bytes.NewBufferString(input), file)

			toks := []scanned{}
			for {
				_, tok, lit := s.Scan()
				toks = append(toks, scanned{tok, lit})
				if tok == token.EOF {
					return toks
				}
			}
		}

		It("should not drop text adjacent to a carriage return", func() {
			Expect(scanAll("a: b\r\nc: d\n")).To(Equal([]scanned{
				{token.TEXT, "a"},
				{token.COLON, ""},
				{token.TEXT, "b"},
				{token.NEWLINE, "\r\n"},
				{token.TEXT, "c"},
				{token.COLON, ""},
				{token.TEXT, "d"},
				{token.EOF, ""},
			}))
		})

		It("should scan a CRLF file", func() {
			Expect(scanAll("VAR := val\r\n# comment\r\ntarget: prereq\r\n\techo hi\r\n")).To(Equal([]scanned{
				{token.TEXT, "VAR"},
				{token.SIMPLE_ASSIGN, ""},
				{token.TEXT, "val"},
				{token.NEWLINE, "\r\n"},
				{token.COMMENT, " comment"},
				{token.NEWLINE, "\r\n"},
				{token.TEXT, "target"},
				{token.COLON, ""},
				{token.TEXT, "prereq"},
				{token.NEWLINE, "\r\n"},
				{token.TAB, ""},
				{token.TEXT, "echo"},
				{token.TEXT, "hi"},
				{token.EOF, ""},
			}))
		})

		It("should scan a CRLF newline in a recipe body", func() {
			Expect(scanAll("target:\r\n\techo hi\r\n\techo bye\r\n")).To(Equal([]scanned{
				{token.TEXT, "target"},
				{token.COLON, ""},
				{token.NEWLINE, "\r\n"},
				{token.TAB, ""},
				{token.TEXT, "echo"},
				{token.TEXT, "hi"},
				{token.NEWLINE, "\r\n"},
				{token.TAB, ""},
				{token.TEXT, "echo"},
				{token.TEXT, "bye"},
				{token.EOF, ""},
			}))
		})

		It("should report a carriage return that does not end a line", func() {
			Expect(scanAll("a: b\rc\n")).To(Equal([]scanned{
				{token.TEXT, "a"},
				{token.COLON, ""},
				{token.TEXT, "b"},
				{token.UNSUPPORTED, "\r"},
				{token.TEXT, "c"},
				{token.EOF, ""},
			}))
		})

		It("should report a trailing carriage return", func() {
			Expect(scanAll("a\r")).To(Equal([]scanned{
				{token.TEXT, "a"},
				{token.UNSUPPORTED, "\r"},
				{token.EOF, ""},
			}))
		})

		It("should track positions across CRLF newlines", func() {
			s := scanner.New(bytes.NewBufferString("a: b\r\nc: d\n"), file)

			for range 3 {
				_, _, _ = s.Scan() // a, :, b
			}

			pos, tok, lit := s.Scan()
			Expect(tok).To(Equal(token.NEWLINE))
			Expect(lit).To(Equal("\r\n"))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   4,
				Line:     1,
				Column:   5,
			}))

			pos, tok, lit = s.Scan()
			Expect(tok).To(Equal(token.TEXT))
			Expect(lit).To(Equal("c"))
			Expect(s.Position(pos)).To(Equal(token.Position{
				Filename: file.Name(),
				Offset:   6,
				Line:     2,
				Column:   1,
			}))
		})
	})

	Describe("LineEnding", func() {
		lineEnding := func(input string) string {
			s := scanner.New(bytes.NewBufferString(input), file)
			for {
				if _, tok, _ := s.Scan(); tok == token.EOF {
					return s.LineEnding()
				}
			}
		}

		DescribeTable("should report the line ending of the input",
			func(input, expected string) {
				Expect(lineEnding(input)).To(Equal(expected))
			},
			Entry("LF", "a: b\nc: d\n", "\n"),
			Entry("CRLF", "a: b\r\nc: d\r\n", "\r\n"),
			Entry("a single CRLF line", "a: b\r\n", "\r\n"),
			Entry("no line ending at all", "a: b", "\n"),
			Entry("mostly CRLF", "a: b\r\nc: d\r\ne: f\n", "\r\n"),
			Entry("mostly LF", "a: b\nc: d\ne: f\r\n", "\n"),
			Entry("a tie", "a: b\r\nc: d\n", "\n"),
		)

		It("should count the line ending folded into EOF", func() {
			s := scanner.New(bytes.NewBufferString("a: b\r\n"), file)

			// The final newline is reported as EOF rather than as a NEWLINE,
			// so nothing in the token stream records it.
			for {
				if _, tok, _ := s.Scan(); tok == token.EOF {
					break
				}
			}

			Expect(s.LineEnding()).To(Equal("\r\n"))
		})
	})

	It("should return IO errors", func() {
		r := testing.ErrReader("io error")
		s := scanner.New(r, file)

		_, _, _ = s.Scan()
		Expect(s.Err()).To(MatchError("io error"))
	})

	It("should support a nil *token.File value", func() {
		buf := bytes.NewBufferString("target:")
		s := scanner.New(buf, nil)

		pos, tok, lit := s.Scan()
		Expect(tok).To(Equal(token.TEXT))
		Expect(lit).To(Equal("target"))
		Expect(pos).To(Equal(token.Pos(1)))
	})
})
