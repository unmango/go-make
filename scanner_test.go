package make_test

import (
	"bytes"
	gotoken "go/token"
	"math"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/internal/testing"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Scanner", func() {
	var file *token.File

	BeforeEach(func() {
		file = gotoken.NewFileSet().AddFile("test", 1, math.MaxInt-2)
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
			s := make.NewScanner(buf, file)

			Expect(s.Scan()).To(BeTrueBecause("scanned a token"))
			Expect(s.Token()).To(Equal(token.IDENT))
			Expect(s.Literal()).To(Equal(strings.TrimSpace(input)))
			Expect(s.Scan()).To(BeFalseBecause("at EOF"))
			Expect(s.Token()).To(Equal(token.EOF))
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
			s := make.NewScanner(buf, file)

			Expect(s.Scan()).To(BeTrueBecause("scanned a token"))
			Expect(s.Token()).To(Equal(token.IDENT))
			Expect(s.Literal()).To(Equal(strings.TrimSpace(input)))
			Expect(s.Scan()).To(BeFalseBecause("at EOF"))
			Expect(s.Token()).To(Equal(token.EOF))
		},
	)

	DescribeTable("Scan ident followed by token",
		Entry(nil, "ident $"),
		Entry(nil, "ident:"),
		Entry(nil, "ident :"),
		Entry(nil, "ident ;"),
		Entry(nil, "ident |"),
		Entry(nil, "ident ="),
		Entry(nil, "ident :="),
		Entry(nil, "ident ::="),
		Entry(nil, "ident :::="),
		Entry(nil, "ident ?="),
		Entry(nil, "ident !="),
		Entry(nil, "ident ("),
		Entry(nil, "ident )"),
		Entry(nil, "ident {"),
		Entry(nil, "ident }"),
		Entry(nil, "ident ,"),
		Entry(nil, "ident\n\t"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			s := make.NewScanner(buf, file)

			ok := s.Scan()

			Expect(s.Token()).To(Equal(token.IDENT))
			Expect(s.Literal()).To(Equal("ident"))
			Expect(ok).To(BeTrueBecause("scanned a token"))
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
		Entry(nil, "(", token.LPAREN),
		Entry(nil, ")", token.RPAREN),
		Entry(nil, "{", token.LBRACE),
		Entry(nil, "}", token.RBRACE),
		Entry(nil, ",", token.COMMA),
		Entry(nil, "\t", token.TAB),
		func(input string, expected token.Token) {
			buf := bytes.NewBufferString(input)
			s := make.NewScanner(buf, file)

			ok := s.Scan()

			Expect(s.Token()).To(Equal(expected))
			Expect(ok).To(BeTrueBecause("scanned a token"))
		},
	)

	DescribeTable("Scan comment tokens",
		Entry(nil, "#", token.COMMENT),
		func(input string, expected token.Token) {
			buf := bytes.NewBufferString(input)
			s := make.NewScanner(buf, file)

			ok := s.Scan()

			Expect(s.Token()).To(Equal(expected))
			Expect(ok).To(BeTrueBecause("scanned a token"))
		},
	)

	It("should scan newline followed by token", func() {
		buf := bytes.NewBufferString("\n ident")
		s := make.NewScanner(buf, file)

		Expect(s.Scan()).To(BeTrue())
		Expect(s.Token()).To(Equal(token.NEWLINE))
	})

	Describe("Pos", func() {
		DescribeTable("Starting token",
			Entry(nil, "$", 2),
			Entry(nil, ":", 2),
			Entry(nil, ";", 2),
			Entry(nil, "|", 2),
			Entry(nil, "=", 2),
			Entry(nil, ":=", 3),
			Entry(nil, "::=", 4),
			Entry(nil, ":::=", 5),
			Entry(nil, "?=", 3),
			Entry(nil, "!=", 3),
			Entry(nil, "(", 2),
			Entry(nil, ")", 2),
			Entry(nil, "{", 2),
			Entry(nil, "}", 2),
			Entry(nil, ",", 2),
			Entry(nil, "\t", 2),
			Entry(nil, "identifier", 11),
			Entry(nil, "$ foo", 2),
			Entry(nil, ": foo", 2),
			Entry(nil, "; foo", 2),
			Entry(nil, "| foo", 2),
			Entry(nil, "= foo", 2),
			Entry(nil, ":= foo", 3),
			Entry(nil, "::= foo", 4),
			Entry(nil, ":::= foo", 5),
			Entry(nil, "?= foo", 3),
			Entry(nil, "!= foo", 3),
			Entry(nil, "( foo", 2),
			Entry(nil, ") foo", 2),
			Entry(nil, "{ foo", 2),
			Entry(nil, "} foo", 2),
			Entry(nil, ", foo", 2),
			Entry(nil, "\t foo", 2),
			Entry(nil, "identifier foo", 11),
			func(input string, expected int) {
				buf := bytes.NewBufferString(input)
				s := make.NewScanner(buf, file)

				Expect(s.Scan()).To(BeTrueBecause("scanned a token"))
				Expect(s.Pos()).To(Equal(token.Pos(expected)))
			},
		)

		DescribeTable("Second token",
			Entry(nil, "$ foo", 6),
			Entry(nil, ": foo", 6),
			Entry(nil, "; foo", 6),
			Entry(nil, "| foo", 6),
			Entry(nil, "= foo", 6),
			Entry(nil, ":= foo", 7),
			Entry(nil, "::= foo", 8),
			Entry(nil, ":::= foo", 9),
			Entry(nil, "?= foo", 7),
			Entry(nil, "!= foo", 7),
			Entry(nil, "( foo", 6),
			Entry(nil, ") foo", 6),
			Entry(nil, "{ foo", 6),
			Entry(nil, "} foo", 6),
			Entry(nil, ", foo", 6),
			Entry(nil, "\t foo", 6),
			Entry(nil, "identifier foo", 15),
			Entry(nil, "$ foo bar", 6),
			Entry(nil, ": foo bar", 6),
			Entry(nil, "; foo bar", 6),
			Entry(nil, "| foo bar", 6),
			Entry(nil, "= foo bar", 6),
			Entry(nil, ":= foo bar", 7),
			Entry(nil, "::= foo bar", 8),
			Entry(nil, ":::= foo bar", 9),
			Entry(nil, "?= foo bar", 7),
			Entry(nil, "!= foo bar", 7),
			Entry(nil, "( foo bar", 6),
			Entry(nil, ") foo bar", 6),
			Entry(nil, "{ foo bar", 6),
			Entry(nil, "} foo bar", 6),
			Entry(nil, ", foo bar", 6),
			Entry(nil, "\t foo bar", 6),
			Entry(nil, "identifier foo bar", 15),
			func(input string, expected int) {
				buf := bytes.NewBufferString(input)
				s := make.NewScanner(buf, file)

				Expect(s.Scan()).To(BeTrueBecause("scanned a token"))
				Expect(s.Scan()).To(BeTrueBecause("scanned another token"))
				Expect(s.Pos()).To(Equal(token.Pos(expected)))
			},
		)

		DescribeTable("Scan newline",
			Entry(nil, "$\nfoo", 3),
			Entry(nil, ":\nfoo", 3),
			Entry(nil, ";\nfoo", 3),
			Entry(nil, "|\nfoo", 3),
			Entry(nil, "=\nfoo", 3),
			Entry(nil, ":=\nfoo", 4),
			Entry(nil, "::=\nfoo", 5),
			Entry(nil, ":::=\nfoo", 6),
			Entry(nil, "?=\nfoo", 4),
			Entry(nil, "!=\nfoo", 4),
			Entry(nil, "(\nfoo", 3),
			Entry(nil, ")\nfoo", 3),
			Entry(nil, "{\nfoo", 3),
			Entry(nil, "}\nfoo", 3),
			Entry(nil, ",\nfoo", 3),
			Entry(nil, "\t\nfoo", 3),
			Entry(nil, "identifier\nfoo", 12),
			func(input string, expected int) {
				buf := bytes.NewBufferString(input)
				s := make.NewScanner(buf, file)

				Expect(s.Scan()).To(BeTrueBecause("scanned first token"))
				Expect(s.Scan()).To(BeTrueBecause("scanned newline token"))
				Expect(s.Pos()).To(Equal(token.Pos(expected)))
				Expect(file.PositionFor(s.Pos(), false)).To(Equal(token.Position{
					Filename: file.Name(),
					Offset:   expected - file.Base(),
					Line:     2,
					Column:   2,
				}))
			},
		)

		DescribeTable("Scan final newline",
			Entry(nil, "$\n", 3),
			Entry(nil, ":\n", 3),
			Entry(nil, ";\n", 3),
			Entry(nil, "|\n", 3),
			Entry(nil, "=\n", 3),
			Entry(nil, ":=\n", 4),
			Entry(nil, "::=\n", 5),
			Entry(nil, ":::=\n", 6),
			Entry(nil, "?=\n", 4),
			Entry(nil, "!=\n", 4),
			Entry(nil, "(\n", 3),
			Entry(nil, ")\n", 3),
			Entry(nil, "{\n", 3),
			Entry(nil, "}\n", 3),
			Entry(nil, ",\n", 3),
			Entry(nil, "\t\n", 3),
			Entry(nil, "identifier\n", 12),
			func(input string, expected int) {
				buf := bytes.NewBufferString(input)
				s := make.NewScanner(buf, file)

				Expect(s.Scan()).To(BeTrueBecause("scanned a token"))
				Expect(s.Scan()).To(BeFalseBecause("scanned final newline"))
				Expect(s.Pos()).To(Equal(token.Pos(expected)))
				Expect(file.PositionFor(s.Pos(), false)).To(Equal(token.Position{
					Filename: file.Name(),
					Offset:   expected - file.Base(),
					Line:     2,
					Column:   2,
				}))
			},
		)
	})

	It("should return IO errors", func() {
		r := testing.ErrReader("io error")
		s := make.NewScanner(r, file)

		Expect(s.Scan()).To(BeFalse())
		Expect(s.Err()).To(MatchError("io error"))
	})
})
