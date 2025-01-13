// Package token defines constants representing the lexical tokens of a Makefile.
// It is based heavily on the [go/token package]
//
// [go/token package]: https://pkg.go.dev/go/token
package token

import (
	"strconv"
)

// Token defines the set of lexical tokens of a Makefile.
// [Quick Reference]
//
// [Quick Reference]: https://www.gnu.org/software/make/manual/html_node/Quick-Reference.html
type Token int

const (
	UNSUPPORTED Token = -1
	ILLEGAL     Token = iota
	EOF
	COMMENT // #comment text

	literal_beg
	IDENT // some_name
	literal_end

	operator_beg
	// Operators and delimiters
	LPAREN  // (
	LBRACE  // {
	RPAREN  // )
	RBRACE  // }
	DOLLAR  // $
	COLON   // :
	COMMA   // ,
	NEWLINE // \n
	TAB     // \t

	RECURSIVE_ASSIGN // =
	SIMPLE_ASSIGN    // :=
	POSIX_ASSIGN     // ::=
	IMMEDIATE_ASSIGN // :::=
	IFNDEF_ASSIGN    // ?=
	SHELL_ASSIGN     // !=
	operator_end

	directive_beg
	// Directives
	DEFINE       // define
	ENDEF        // endef
	UNDEFINE     // undefine
	IFDEF        // ifdef
	IFNDEF       // ifndef
	IFEQ         // ifeq
	IFNEQ        // ifneq
	ELSE         // else
	ENDIF        // endif
	INCLUDE      // include
	DASH_INCLUDE // -include
	SINCLUDE     // sinclude
	OVERRIDE     // override
	EXPORT       // export
	UNEXPORT     // unexport
	PRIVATE      // private
	VPATH        // vpath
	directive_end

	function_beg
	// Built-in functions
	SUBST      // $(subst from,to.text)
	PATSUBST   // $(patsubst pattern,replacement,text)
	STRIP      // $(strip string)
	FINDSTRING // $(findstring find,text)
	FILTER     // $(filter patern...,text)
	FILTER_OUT // $(filter-out patern...,text)
	SORT       // $(sort list)
	WORD       // $(word n,text)
	WORDS      // $(words text)
	WORDLIST   // $(wordlist s,e,text)
	FIRSTWORD  // $(firstword names...)
	LASTWORD   // $(lastword names...)
	DIR        // $(dir names...)
	NOTDIR     // $(notdir names...)
	SUFFIX     // $(suffix names...)
	BASENAME   // $(basename names...)
	ADDSUFFIX  // $(addsuffix suffix,names...)
	ADDPREFIX  // $(addprefix prefix,names...)
	JOIN       // $(join list1,list2)
	WILDCARD   // $(wildcard pattern...)
	REALPATH   // $(realpath names...)
	ABSPATH    // $(abspath names...)
	ERROR      // $(error text...)
	WARNING    // $(warning text...)
	SHELL      // $(shell command)
	ORIGIN     // $(origin variable)
	FLAVOR     // $(flavor variable)
	LET        // $(let var [var ...],words,text)
	FOREACH    // $(foreach var,words,text)
	IF         // $(if condition,then-part[,else-part])
	OR         // $(or condition1[,condition2[,condition3…]])
	AND        // $(and condition1[,condition2[,condition3…]])
	INTCMP     // $(intcmp lhs,rhs[,lt-part[,eq-part[,gt-part]]])
	CALL       // $(call var,param,...)
	EVAL       // $(eval text)
	FILE       // $(file op filename,text)
	VALUE      // $(value var)
	function_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",
	IDENT:   "IDENT",

	LPAREN:  "(",
	LBRACE:  "{",
	RPAREN:  ")",
	RBRACE:  "}",
	DOLLAR:  "$",
	COLON:   ":",
	COMMA:   ",",
	NEWLINE: "\n",
	TAB:     "\t",

	RECURSIVE_ASSIGN: "=",
	SIMPLE_ASSIGN:    ":=",
	POSIX_ASSIGN:     "::=",
	IMMEDIATE_ASSIGN: ":::=",
	IFNDEF_ASSIGN:    "?=",
	SHELL_ASSIGN:     "!=",

	DEFINE:       "define",
	ENDEF:        "endef",
	UNDEFINE:     "undefine",
	IFDEF:        "ifdef",
	IFNDEF:       "ifndef",
	IFEQ:         "ifeq",
	IFNEQ:        "ifneq",
	ELSE:         "else",
	ENDIF:        "endif",
	INCLUDE:      "include",
	DASH_INCLUDE: "-include",
	SINCLUDE:     "sinclude",
	OVERRIDE:     "override",
	EXPORT:       "export",
	UNEXPORT:     "unexport",
	PRIVATE:      "private",
	VPATH:        "vpath",

	SUBST:      "subst",
	PATSUBST:   "patsubst",
	STRIP:      "strip",
	FINDSTRING: "findstring",
	FILTER:     "filter",
	FILTER_OUT: "filter-out",
	SORT:       "sort",
	WORD:       "word",
	WORDS:      "words",
	WORDLIST:   "wordlist",
	FIRSTWORD:  "firstword",
	LASTWORD:   "lastword",
	DIR:        "dir",
	NOTDIR:     "notdir",
	SUFFIX:     "suffix",
	BASENAME:   "basename",
	ADDSUFFIX:  "addsuffix",
	ADDPREFIX:  "addprefix",
	JOIN:       "join",
	WILDCARD:   "wildcard",
	REALPATH:   "realpath",
	ABSPATH:    "abspath",
	ERROR:      "error",
	WARNING:    "warning",
	SHELL:      "shell",
	ORIGIN:     "origin",
	FLAVOR:     "flavor",
	LET:        "let",
	FOREACH:    "foreach",
	IF:         "if",
	OR:         "or",
	AND:        "and",
	INTCMP:     "intcmp",
	CALL:       "call",
	EVAL:       "eval",
	FILE:       "file",
	VALUE:      "value",
}

// String returns the string corresponding to the token tok.
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}

	return s
}

var (
	directives map[string]Token
	functions  map[string]Token
	keywords   map[string]Token
)

func init() {
	directives = make(map[string]Token, directive_end-(directive_beg+1))
	for i := directive_beg + 1; i < directive_end; i++ {
		directives[tokens[i]] = i
	}

	functions = make(map[string]Token, function_end-(function_beg+1))
	for i := function_beg + 1; i < function_end; i++ {
		functions[tokens[i]] = i
	}

	keywords = make(map[string]Token, len(directives)+len(functions))
	for k, v := range directives {
		keywords[k] = v
	}
	for k, v := range functions {
		keywords[k] = v
	}
}

func Lookup(ident string) Token {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}

// IsLiteral returns true for tokens corresponding to identifiers.
func (tok Token) IsLiteral() bool {
	return literal_beg < tok && tok < literal_end
}

// IsOperator returns true for tokens corresponding to operators.
func (tok Token) IsOperator() bool {
	return operator_beg < tok && tok < operator_end
}

// IsDirective returns true for tokens corresponding to directives.
func (tok Token) IsDirective() bool {
	return directive_beg < tok && tok < directive_end
}

// IsBuiltinFunction returns true for tokens corresponding to built-in functions.
func (tok Token) IsBuiltinFunction() bool {
	return function_beg < tok && tok < function_end
}

// IsKeyword reports whether name is a keyword, that is, whether
// name is either a directive or a built-in function.
func IsKeyword(name string) bool {
	return IsDirective(name) || IsBuiltinFunction(name)
}

// IsDirective reports whether name is a directive.
func IsDirective(name string) bool {
	_, ok := directives[name]
	return ok
}

// IsBuiltinFunction reports whether name is a built-in function.
func IsBuiltinFunction(name string) bool {
	_, ok := functions[name]
	return ok
}

// IsIdentifier reports whether name is a valid identifier.
// Keywords are not identifiers.
func IsIdentifier(name string) bool {
	if name == "" || IsKeyword(name) {
		return false
	}
	switch name {
	case "(", ")", "{", "}", "$", ":", ",", "\n", "\t":
		fallthrough
	case "=", ":=", "::=", ":::=", "?=", "!=":
		return false
	}

	return true
}
