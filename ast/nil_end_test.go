package ast_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

var _ = Describe("nil AST list entries", func() {
	It("should skip nil file contents when finding the end", func() {
		file := &ast.File{FileEnd: 42, Contents: []ast.Obj{&ast.BadObj{To: 7}, nil}}
		Expect(file.End()).To(Equal(token.Pos(7)))
	})

	It("should skip nil variable values when finding the end", func() {
		variable := &ast.Variable{OpPos: 5, Op: token.Token(0), Value: []ast.Expr{&ast.Text{ValuePos: 10}, nil}}
		Expect(variable.End()).To(Equal(token.Pos(10)))
	})
})
