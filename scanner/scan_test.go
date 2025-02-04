package scanner_test

import (
	"bufio"
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
	})
})
