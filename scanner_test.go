package make_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/internal/testing"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Scanner", func() {
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
			s := make.NewScanner(buf)

			Expect(s.Scan()).To(BeTrueBecause("more to scan"))
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
			s := make.NewScanner(buf)

			Expect(s.Scan()).To(BeTrueBecause("more to scan"))
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
			s := make.NewScanner(buf)

			more := s.Scan()

			Expect(s.Token()).To(Equal(token.IDENT))
			Expect(s.Literal()).To(Equal("ident"))
			Expect(more).To(BeTrueBecause("more to scan"))
		},
	)

	DescribeTable("Scan non-ident tokens",
		Entry(nil, "$", token.DOLLAR),
		Entry(nil, ":", token.COLON),
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
			s := make.NewScanner(buf)

			more := s.Scan()

			Expect(s.Token()).To(Equal(expected))
			Expect(more).To(BeTrueBecause("more to scan"))
		},
	)

	DescribeTable("Scan comment tokens",
		Entry(nil, "#", token.COMMENT),
		func(input string, expected token.Token) {
			buf := bytes.NewBufferString(input)
			s := make.NewScanner(buf)

			more := s.Scan()

			Expect(s.Token()).To(Equal(expected))
			Expect(more).To(BeTrueBecause("more to scan"))
		},
	)

	It("should scan newline followed by token", func() {
		buf := bytes.NewBufferString("\n ident")
		s := make.NewScanner(buf)

		Expect(s.Scan()).To(BeTrue())
		Expect(s.Token()).To(Equal(token.NEWLINE))
	})

	Describe("Pos", func() {
		DescribeTable("Starting token",
			Entry(nil, "$", 2),
			Entry(nil, ":", 2),
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
			func(input string, expected int) {
				buf := bytes.NewBufferString(input)
				s := make.NewScanner(buf)

				Expect(s.Scan()).To(BeTrueBecause("scanned a token"))
				Expect(s.Pos()).To(Equal(token.Pos(expected)))
			},
		)
	})

	It("should return IO errors", func() {
		s := make.NewScanner(testing.ErrReader("io error"))

		Expect(s.Scan()).To(BeFalse())
		Expect(s.Err()).To(MatchError("io error"))
	})
})
