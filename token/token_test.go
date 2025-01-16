package token_test

import (
	"testing/quick"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/token"
)

var Literals = []TableEntry{
	Entry(nil, token.IDENT),
}

var Operators = []TableEntry{
	Entry(nil, token.LPAREN),
	Entry(nil, token.LBRACE),
	Entry(nil, token.RPAREN),
	Entry(nil, token.RBRACE),
	Entry(nil, token.DOLLAR),
	Entry(nil, token.COLON),
	Entry(nil, token.COMMA),
	Entry(nil, token.NEWLINE),
	Entry(nil, token.TAB),
	Entry(nil, token.RECURSIVE_ASSIGN),
	Entry(nil, token.SIMPLE_ASSIGN),
	Entry(nil, token.POSIX_ASSIGN),
	Entry(nil, token.IMMEDIATE_ASSIGN),
	Entry(nil, token.IFNDEF_ASSIGN),
	Entry(nil, token.SHELL_ASSIGN),
}

var Directives = []TableEntry{
	Entry(nil, token.DEFINE),
	Entry(nil, token.ENDEF),
	Entry(nil, token.UNDEFINE),
	Entry(nil, token.IFDEF),
	Entry(nil, token.IFNDEF),
	Entry(nil, token.IFEQ),
	Entry(nil, token.IFNEQ),
	Entry(nil, token.ELSE),
	Entry(nil, token.ENDIF),
	Entry(nil, token.INCLUDE),
	Entry(nil, token.DASH_INCLUDE),
	Entry(nil, token.SINCLUDE),
	Entry(nil, token.OVERRIDE),
	Entry(nil, token.EXPORT),
	Entry(nil, token.UNEXPORT),
	Entry(nil, token.PRIVATE),
	Entry(nil, token.VPATH),
}

var Functions = []TableEntry{
	Entry(nil, token.SUBST),
	Entry(nil, token.PATSUBST),
	Entry(nil, token.STRIP),
	Entry(nil, token.FINDSTRING),
	Entry(nil, token.FILTER),
	Entry(nil, token.FILTER_OUT),
	Entry(nil, token.SORT),
	Entry(nil, token.WORD),
	Entry(nil, token.WORDS),
	Entry(nil, token.WORDLIST),
	Entry(nil, token.FIRSTWORD),
	Entry(nil, token.LASTWORD),
	Entry(nil, token.DIR),
	Entry(nil, token.NOTDIR),
	Entry(nil, token.SUFFIX),
	Entry(nil, token.BASENAME),
	Entry(nil, token.ADDSUFFIX),
	Entry(nil, token.ADDPREFIX),
	Entry(nil, token.JOIN),
	Entry(nil, token.WILDCARD),
	Entry(nil, token.REALPATH),
	Entry(nil, token.ABSPATH),
	Entry(nil, token.ERROR),
	Entry(nil, token.WARNING),
	Entry(nil, token.SHELL),
	Entry(nil, token.ORIGIN),
	Entry(nil, token.FLAVOR),
	Entry(nil, token.LET),
	Entry(nil, token.FOREACH),
	Entry(nil, token.IF),
	Entry(nil, token.OR),
	Entry(nil, token.AND),
	Entry(nil, token.INTCMP),
	Entry(nil, token.CALL),
	Entry(nil, token.EVAL),
	Entry(nil, token.FILE),
	Entry(nil, token.VALUE),
}

var _ = Describe("Token", func() {
	Describe("IsLiteral", func() {
		DescribeTable("true", Literals,
			func(tok token.Token) {
				Expect(tok.IsLiteral()).To(BeTrue())
			},
		)

		DescribeTable("false", Operators, Directives, Functions,
			func(tok token.Token) {
				Expect(tok.IsLiteral()).To(BeFalse())
			},
		)
	})

	Describe("IsOperator", func() {
		DescribeTable("true", Operators,
			func(tok token.Token) {
				Expect(tok.IsOperator()).To(BeTrue())
			},
		)

		DescribeTable("false", Literals, Directives, Functions,
			func(tok token.Token) {
				Expect(tok.IsOperator()).To(BeFalse())
			},
		)
	})

	Describe("IsDirective", func() {
		DescribeTable("true", Directives,
			func(tok token.Token) {
				Expect(tok.IsDirective()).To(BeTrue())
			},
		)

		DescribeTable("false", Literals, Operators, Functions,
			func(tok token.Token) {
				Expect(tok.IsDirective()).To(BeFalse())
			},
		)
	})

	Describe("IsBuiltinFunction", func() {
		DescribeTable("true", Functions,
			func(tok token.Token) {
				Expect(tok.IsBuiltinFunction()).To(BeTrue())
			},
		)

		DescribeTable("false", Literals, Directives, Operators,
			func(tok token.Token) {
				Expect(tok.IsBuiltinFunction()).To(BeFalse())
			},
		)
	})

	DescribeTable("String",
		Entry(nil, token.IDENT, "IDENT"),
		Entry(nil, token.LPAREN, "("),
		Entry(nil, token.LBRACE, "{"),
		Entry(nil, token.RPAREN, ")"),
		Entry(nil, token.RBRACE, "}"),
		Entry(nil, token.DOLLAR, "$"),
		Entry(nil, token.COLON, ":"),
		Entry(nil, token.COMMA, ","),
		Entry(nil, token.NEWLINE, "\n"),
		Entry(nil, token.TAB, "\t"),
		Entry(nil, token.RECURSIVE_ASSIGN, "="),
		Entry(nil, token.SIMPLE_ASSIGN, ":="),
		Entry(nil, token.POSIX_ASSIGN, "::="),
		Entry(nil, token.IMMEDIATE_ASSIGN, ":::="),
		Entry(nil, token.IFNDEF_ASSIGN, "?="),
		Entry(nil, token.SHELL_ASSIGN, "!="),
		Entry(nil, token.DEFINE, "define"),
		Entry(nil, token.ENDEF, "endef"),
		Entry(nil, token.UNDEFINE, "undefine"),
		Entry(nil, token.IFDEF, "ifdef"),
		Entry(nil, token.IFNDEF, "ifndef"),
		Entry(nil, token.IFEQ, "ifeq"),
		Entry(nil, token.IFNEQ, "ifneq"),
		Entry(nil, token.ELSE, "else"),
		Entry(nil, token.ENDIF, "endif"),
		Entry(nil, token.INCLUDE, "include"),
		Entry(nil, token.DASH_INCLUDE, "-include"),
		Entry(nil, token.SINCLUDE, "sinclude"),
		Entry(nil, token.OVERRIDE, "override"),
		Entry(nil, token.EXPORT, "export"),
		Entry(nil, token.UNEXPORT, "unexport"),
		Entry(nil, token.PRIVATE, "private"),
		Entry(nil, token.VPATH, "vpath"),
		Entry(nil, token.SUBST, "subst"),
		Entry(nil, token.PATSUBST, "patsubst"),
		Entry(nil, token.STRIP, "strip"),
		Entry(nil, token.FINDSTRING, "findstring"),
		Entry(nil, token.FILTER, "filter"),
		Entry(nil, token.FILTER_OUT, "filter-out"),
		Entry(nil, token.SORT, "sort"),
		Entry(nil, token.WORD, "word"),
		Entry(nil, token.WORDS, "words"),
		Entry(nil, token.WORDLIST, "wordlist"),
		Entry(nil, token.FIRSTWORD, "firstword"),
		Entry(nil, token.LASTWORD, "lastword"),
		Entry(nil, token.DIR, "dir"),
		Entry(nil, token.NOTDIR, "notdir"),
		Entry(nil, token.SUFFIX, "suffix"),
		Entry(nil, token.BASENAME, "basename"),
		Entry(nil, token.ADDSUFFIX, "addsuffix"),
		Entry(nil, token.ADDPREFIX, "addprefix"),
		Entry(nil, token.JOIN, "join"),
		Entry(nil, token.WILDCARD, "wildcard"),
		Entry(nil, token.REALPATH, "realpath"),
		Entry(nil, token.ABSPATH, "abspath"),
		Entry(nil, token.ERROR, "error"),
		Entry(nil, token.WARNING, "warning"),
		Entry(nil, token.SHELL, "shell"),
		Entry(nil, token.ORIGIN, "origin"),
		Entry(nil, token.FLAVOR, "flavor"),
		Entry(nil, token.LET, "let"),
		Entry(nil, token.FOREACH, "foreach"),
		Entry(nil, token.IF, "if"),
		Entry(nil, token.OR, "or"),
		Entry(nil, token.AND, "and"),
		Entry(nil, token.INTCMP, "intcmp"),
		Entry(nil, token.CALL, "call"),
		Entry(nil, token.EVAL, "eval"),
		Entry(nil, token.FILE, "file"),
		Entry(nil, token.VALUE, "value"),
		Entry(nil, token.Token(420), "token(420)"),
		func(tok token.Token, expected string) {
			Expect(tok.String()).To(Equal(expected))
		},
	)

	Describe("Lookup", func() {
		DescribeTable("Lookup keyword",
			Entry(nil, token.DEFINE, "define"),
			Entry(nil, token.ENDEF, "endef"),
			Entry(nil, token.UNDEFINE, "undefine"),
			Entry(nil, token.IFDEF, "ifdef"),
			Entry(nil, token.IFNDEF, "ifndef"),
			Entry(nil, token.IFEQ, "ifeq"),
			Entry(nil, token.IFNEQ, "ifneq"),
			Entry(nil, token.ELSE, "else"),
			Entry(nil, token.ENDIF, "endif"),
			Entry(nil, token.INCLUDE, "include"),
			Entry(nil, token.DASH_INCLUDE, "-include"),
			Entry(nil, token.SINCLUDE, "sinclude"),
			Entry(nil, token.OVERRIDE, "override"),
			Entry(nil, token.EXPORT, "export"),
			Entry(nil, token.UNEXPORT, "unexport"),
			Entry(nil, token.PRIVATE, "private"),
			Entry(nil, token.VPATH, "vpath"),
			Entry(nil, token.SUBST, "subst"),
			Entry(nil, token.PATSUBST, "patsubst"),
			Entry(nil, token.STRIP, "strip"),
			Entry(nil, token.FINDSTRING, "findstring"),
			Entry(nil, token.FILTER, "filter"),
			Entry(nil, token.FILTER_OUT, "filter-out"),
			Entry(nil, token.SORT, "sort"),
			Entry(nil, token.WORD, "word"),
			Entry(nil, token.WORDS, "words"),
			Entry(nil, token.WORDLIST, "wordlist"),
			Entry(nil, token.FIRSTWORD, "firstword"),
			Entry(nil, token.LASTWORD, "lastword"),
			Entry(nil, token.DIR, "dir"),
			Entry(nil, token.NOTDIR, "notdir"),
			Entry(nil, token.SUFFIX, "suffix"),
			Entry(nil, token.BASENAME, "basename"),
			Entry(nil, token.ADDSUFFIX, "addsuffix"),
			Entry(nil, token.ADDPREFIX, "addprefix"),
			Entry(nil, token.JOIN, "join"),
			Entry(nil, token.WILDCARD, "wildcard"),
			Entry(nil, token.REALPATH, "realpath"),
			Entry(nil, token.ABSPATH, "abspath"),
			Entry(nil, token.ERROR, "error"),
			Entry(nil, token.WARNING, "warning"),
			Entry(nil, token.SHELL, "shell"),
			Entry(nil, token.ORIGIN, "origin"),
			Entry(nil, token.FLAVOR, "flavor"),
			Entry(nil, token.LET, "let"),
			Entry(nil, token.FOREACH, "foreach"),
			Entry(nil, token.IF, "if"),
			Entry(nil, token.OR, "or"),
			Entry(nil, token.AND, "and"),
			Entry(nil, token.INTCMP, "intcmp"),
			Entry(nil, token.CALL, "call"),
			Entry(nil, token.EVAL, "eval"),
			Entry(nil, token.FILE, "file"),
			Entry(nil, token.VALUE, "value"),
			func(tok token.Token, keyword string) {
				Expect(token.Lookup(keyword)).To(Equal(tok))
			},
		)

		It("should lookup identifiers", func() {
			err := quick.Check(func(s string) bool {
				return token.Lookup(s) == token.IDENT
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("IsKeyword", func() {
		DescribeTable("keywords",
			Entry(nil, "define"),
			Entry(nil, "endef"),
			Entry(nil, "undefine"),
			Entry(nil, "ifdef"),
			Entry(nil, "ifndef"),
			Entry(nil, "ifeq"),
			Entry(nil, "ifneq"),
			Entry(nil, "else"),
			Entry(nil, "endif"),
			Entry(nil, "include"),
			Entry(nil, "-include"),
			Entry(nil, "sinclude"),
			Entry(nil, "override"),
			Entry(nil, "export"),
			Entry(nil, "unexport"),
			Entry(nil, "private"),
			Entry(nil, "vpath"),
			Entry(nil, "subst"),
			Entry(nil, "patsubst"),
			Entry(nil, "strip"),
			Entry(nil, "findstring"),
			Entry(nil, "filter"),
			Entry(nil, "filter-out"),
			Entry(nil, "sort"),
			Entry(nil, "word"),
			Entry(nil, "words"),
			Entry(nil, "wordlist"),
			Entry(nil, "firstword"),
			Entry(nil, "lastword"),
			Entry(nil, "dir"),
			Entry(nil, "notdir"),
			Entry(nil, "suffix"),
			Entry(nil, "basename"),
			Entry(nil, "addsuffix"),
			Entry(nil, "addprefix"),
			Entry(nil, "join"),
			Entry(nil, "wildcard"),
			Entry(nil, "realpath"),
			Entry(nil, "abspath"),
			Entry(nil, "error"),
			Entry(nil, "warning"),
			Entry(nil, "shell"),
			Entry(nil, "origin"),
			Entry(nil, "flavor"),
			Entry(nil, "let"),
			Entry(nil, "foreach"),
			Entry(nil, "if"),
			Entry(nil, "or"),
			Entry(nil, "and"),
			Entry(nil, "intcmp"),
			Entry(nil, "call"),
			Entry(nil, "eval"),
			Entry(nil, "file"),
			Entry(nil, "value"),
			func(keyword string) {
				Expect(token.IsKeyword(keyword)).To(BeTrue())
			},
		)

		It("should return false for non-keywords", func() {
			err := quick.Check(func(s string) bool {
				return !token.IsKeyword(s)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("IsDirective", func() {
		DescribeTable("directives",
			Entry(nil, "define"),
			Entry(nil, "endef"),
			Entry(nil, "undefine"),
			Entry(nil, "ifdef"),
			Entry(nil, "ifndef"),
			Entry(nil, "ifeq"),
			Entry(nil, "ifneq"),
			Entry(nil, "else"),
			Entry(nil, "endif"),
			Entry(nil, "include"),
			Entry(nil, "-include"),
			Entry(nil, "sinclude"),
			Entry(nil, "override"),
			Entry(nil, "export"),
			Entry(nil, "unexport"),
			Entry(nil, "private"),
			Entry(nil, "vpath"),
			func(directive string) {
				Expect(token.IsDirective(directive)).To(BeTrue())
			},
		)

		It("should return false for non-directive", func() {
			err := quick.Check(func(s string) bool {
				return !token.IsDirective(s)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("IsBuiltinFunction", func() {
		DescribeTable("functions",
			Entry(nil, "subst"),
			Entry(nil, "patsubst"),
			Entry(nil, "strip"),
			Entry(nil, "findstring"),
			Entry(nil, "filter"),
			Entry(nil, "filter-out"),
			Entry(nil, "sort"),
			Entry(nil, "word"),
			Entry(nil, "words"),
			Entry(nil, "wordlist"),
			Entry(nil, "firstword"),
			Entry(nil, "lastword"),
			Entry(nil, "dir"),
			Entry(nil, "notdir"),
			Entry(nil, "suffix"),
			Entry(nil, "basename"),
			Entry(nil, "addsuffix"),
			Entry(nil, "addprefix"),
			Entry(nil, "join"),
			Entry(nil, "wildcard"),
			Entry(nil, "realpath"),
			Entry(nil, "abspath"),
			Entry(nil, "error"),
			Entry(nil, "warning"),
			Entry(nil, "shell"),
			Entry(nil, "origin"),
			Entry(nil, "flavor"),
			Entry(nil, "let"),
			Entry(nil, "foreach"),
			Entry(nil, "if"),
			Entry(nil, "or"),
			Entry(nil, "and"),
			Entry(nil, "intcmp"),
			Entry(nil, "call"),
			Entry(nil, "eval"),
			Entry(nil, "file"),
			Entry(nil, "value"),
			func(function string) {
				Expect(token.IsBuiltinFunction(function)).To(BeTrue())
			},
		)

		It("should return false for non-functions", func() {
			err := quick.Check(func(s string) bool {
				return !token.IsBuiltinFunction(s)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("IsIdentifier", func() {
		DescribeTable("keywords",
			Entry(nil, "define"),
			Entry(nil, "endef"),
			Entry(nil, "undefine"),
			Entry(nil, "ifdef"),
			Entry(nil, "ifndef"),
			Entry(nil, "ifeq"),
			Entry(nil, "ifneq"),
			Entry(nil, "else"),
			Entry(nil, "endif"),
			Entry(nil, "include"),
			Entry(nil, "-include"),
			Entry(nil, "sinclude"),
			Entry(nil, "override"),
			Entry(nil, "export"),
			Entry(nil, "unexport"),
			Entry(nil, "private"),
			Entry(nil, "vpath"),
			Entry(nil, "subst"),
			Entry(nil, "patsubst"),
			Entry(nil, "strip"),
			Entry(nil, "findstring"),
			Entry(nil, "filter"),
			Entry(nil, "filter-out"),
			Entry(nil, "sort"),
			Entry(nil, "word"),
			Entry(nil, "words"),
			Entry(nil, "wordlist"),
			Entry(nil, "firstword"),
			Entry(nil, "lastword"),
			Entry(nil, "dir"),
			Entry(nil, "notdir"),
			Entry(nil, "suffix"),
			Entry(nil, "basename"),
			Entry(nil, "addsuffix"),
			Entry(nil, "addprefix"),
			Entry(nil, "join"),
			Entry(nil, "wildcard"),
			Entry(nil, "realpath"),
			Entry(nil, "abspath"),
			Entry(nil, "error"),
			Entry(nil, "warning"),
			Entry(nil, "shell"),
			Entry(nil, "origin"),
			Entry(nil, "flavor"),
			Entry(nil, "let"),
			Entry(nil, "foreach"),
			Entry(nil, "if"),
			Entry(nil, "or"),
			Entry(nil, "and"),
			Entry(nil, "intcmp"),
			Entry(nil, "call"),
			Entry(nil, "eval"),
			Entry(nil, "file"),
			Entry(nil, "value"),
			Entry(nil, ""),
			func(keyword string) {
				Expect(token.IsIdentifier(keyword)).To(BeFalse())
			},
		)

		DescribeTable("special characters",
			Entry(nil, "("),
			Entry(nil, ")"),
			Entry(nil, "{"),
			Entry(nil, "}"),
			Entry(nil, ":"),
			Entry(nil, ";"),
			Entry(nil, "$"),
			Entry(nil, "#"),
			Entry(nil, ","),
			Entry(nil, "="),
			Entry(nil, ":="),
			Entry(nil, "::="),
			Entry(nil, ":::="),
			Entry(nil, "\n"),
			Entry(nil, "\t"),
			Entry(nil, "?="),
			Entry(nil, "!="),
			Entry(nil, "|"),
			Entry(nil, " "),
			func(keyword string) {
				Expect(token.IsIdentifier(keyword)).To(BeFalse())
			},
		)

		DescribeTable("should return true for non-keywords",
			Entry(nil, "b"),
			Entry(nil, "blah"),
			Entry(nil, "foo"),
			Entry(nil, "foo-bar"),
			Entry(nil, "foo_bar"),
			Entry(nil, "./file/path"),
			Entry(nil, "/abs/path"),
			Entry(nil, "12"),
			func(s string) {
				Expect(token.IsIdentifier(s)).To(BeTrue())
			},
		)
	})
})
