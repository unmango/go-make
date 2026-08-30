package scanner_test

import (
	"bufio"
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/internal/testing"
	"github.com/unmango/go-make/scanner"
)

var _ = Describe("Scan", func() {
	Describe("ScanTokens", func() {
		DescribeTable("Scanner",
			Entry("target",
				"target:", []string{"target", ":"},
			),
			Entry("target with a separating space",
				"target :", []string{"target", " ", ":"},
			),
			Entry("multiple targets",
				"target target2:", []string{"target", " ", "target2", ":"},
			),
			Entry("multiple targets with a separating space",
				"target target2 :", []string{"target", " ", "target2", " ", ":"},
			),
			Entry("target with a trailing newline",
				"target:\n", []string{"target", ":", "\n"},
			),
			Entry("target with a trailing CRLF newline",
				"target:\r\n", []string{"target", ":", "\r\n"},
			),
			Entry("target with a prereq and a trailing CRLF newline",
				"target: prereq\r\n", []string{"target", ":", " ", "prereq", "\r\n"},
			),
			Entry("recipe with a trailing CRLF newline",
				"target:\r\n\trecipe\r\n",
				[]string{"target", ":", "\r\n", "\t", "recipe", "\r\n"},
			),
			Entry("carriage return without a newline",
				"target\rprereq", []string{"target", "\r", "prereq"},
			),
			Entry("carriage return at the end of the input",
				"target\r", []string{"target", "\r"},
			),
			Entry("target with a prereq",
				"target: prereq", []string{"target", ":", " ", "prereq"},
			),
			Entry("target with a prereq and trailing newline",
				"target: prereq\n", []string{"target", ":", " ", "prereq", "\n"},
			),
			Entry("target with multiple prereqs",
				"target: prereq prereq2", []string{"target", ":", " ", "prereq", " ", "prereq2"},
			),
			Entry("target with a recipe",
				"target:\n\trecipe", []string{"target", ":", "\n", "\t", "recipe"},
			),
			Entry("target with a semicolon recipe",
				"target: ; recipe",
				[]string{"target", ":", " ", ";", " ", "recipe"},
			),
			Entry("target with a semicolon recipe written against the colon",
				"target:; recipe", []string{"target", ":", ";", " ", "recipe"},
			),
			Entry("semicolon with no separating space",
				"target: ;recipe", []string{"target", ":", " ", ";recipe"},
			),
			Entry("target with a recipe and trailing newline",
				"target:\n\trecipe\n", []string{"target", ":", "\n", "\t", "recipe", "\n"},
			),
			Entry("target with multiple recipes",
				"target:\n\trecipe\n\trecipe2",
				[]string{"target", ":", "\n", "\t", "recipe", "\n", "\t", "recipe2"},
			),
			Entry("comment",
				"# comment", []string{"#", " ", "comment"},
			),
			Entry("comment with multiple words",
				"# comment word", []string{"#", " ", "comment", " ", "word"},
			),
			Entry("comment with a trailing newline",
				"# comment\n", []string{"#", " ", "comment", "\n"},
			),
			Entry("target with a comment",
				"target: # comment", []string{"target", ":", " ", "#", " ", "comment"},
			),
			Entry("directive",
				"define TEST", []string{"define", " ", "TEST"},
			),
			Entry("prefixed include directive",
				"-include foo.mk", []string{"-include", " ", "foo.mk"},
			),
			Entry("variable",
				"VAR := test", []string{"VAR", " ", ":=", " ", "test"},
			),
			Entry("variable with a trailing newline",
				"VAR := test\n", []string{"VAR", " ", ":=", " ", "test", "\n"},
			),
			Entry("recursive variable",
				"VAR = test", []string{"VAR", " ", "=", " ", "test"},
			),
			Entry("posix variable",
				"VAR ::= test", []string{"VAR", " ", "::=", " ", "test"},
			),
			Entry("immediate variable",
				"VAR :::= test", []string{"VAR", " ", ":::=", " ", "test"},
			),
			Entry("ifndef variable",
				"VAR ?= test", []string{"VAR", " ", "?=", " ", "test"},
			),
			Entry("shell variable",
				"VAR != test", []string{"VAR", " ", "!=", " ", "test"},
			),
			Entry("info function",
				"$(info thing)", []string{"$", "(", "info", " ", "thing", ")"},
			),
			Entry("subst function",
				"$(subst from,to,text)", []string{"$", "(", "subst", " ", "from", ",", "to", ",", "text", ")"},
			),
			Entry("ifeq directive",
				"ifeq (foo, bar)", []string{"ifeq", " ", "(", "foo", ",", " ", "bar", ")"},
			),
			Entry("ifeq directive with quotes",
				`ifeq 'foo' "bar"`, []string{"ifeq", " ", "'", "foo", "'", " ", `"`, "bar", `"`},
			),
			func(text string, expected []string) {
				buf := bytes.NewBufferString(text)
				s := bufio.NewScanner(buf)
				s.Split(scanner.ScanTokens)

				tokens := []string{}
				for s.Scan() {
					tokens = append(tokens, s.Text())
				}
				Expect(s.Err()).NotTo(HaveOccurred())
				Expect(tokens).To(Equal(expected))
			},
		)

		It("should not panic when data is empty and atEOF is false", func() {
			var advance int
			var token []byte
			var err error

			Expect(func() {
				advance, token, err = scanner.ScanTokens(nil, false)
			}).NotTo(Panic())

			Expect(advance).To(Equal(0))
			Expect(token).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("CRLF split across reads",
			Entry("between the carriage return and the newline",
				"ab\r\ncd\r\n", 3, []string{"ab", "\r\n", "cd", "\r\n"},
			),
			Entry("with the carriage return alone in a read",
				"a\r\nb\r\n", 2, []string{"a", "\r\n", "b", "\r\n"},
			),
			func(text string, chunk int, expected []string) {
				r := testing.NewChunkReader(text, chunk)
				s := bufio.NewScanner(r)
				s.Split(scanner.ScanTokens)

				tokens := []string{}
				for s.Scan() {
					tokens = append(tokens, s.Text())
				}
				Expect(s.Err()).NotTo(HaveOccurred())
				Expect(tokens).To(Equal(expected))
			},
		)

		DescribeTable("assignments split across reads",
			Entry("immediate variable",
				"A:::=b\n", []string{"A", ":::=", "b", "\n"},
			),
			Entry("posix variable",
				"A::=b\n", []string{"A", "::=", "b", "\n"},
			),
			Entry("simple variable",
				"A:=b\n", []string{"A", ":=", "b", "\n"},
			),
			func(text string, expected []string) {
				r := testing.NewChunkReader(text, 3)
				s := bufio.NewScanner(r)
				s.Split(scanner.ScanTokens)

				tokens := []string{}
				for s.Scan() {
					tokens = append(tokens, s.Text())
				}
				Expect(s.Err()).NotTo(HaveOccurred())
				Expect(tokens).To(Equal(expected))
			},
		)
	})
})
