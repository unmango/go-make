package builder_test

// import (
// 	"bytes"
// 	"testing/quick"

// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"

// 	"github.com/unmango/go-make"
// 	"github.com/unmango/go-make/ast"
// 	"github.com/unmango/go-make/builder"
// 	"github.com/unmango/go-make/builder/build"
// 	"github.com/unmango/go-make/builder/expr"
// 	"github.com/unmango/go-make/builder/file"
// 	"github.com/unmango/go-make/builder/rule"
// 	"github.com/unmango/go-make/token"
// )

// var _ = Describe("Nodes", func() {
// 	Describe("NewFile", func() {
// 		It("should set the file start", func() {
// 			err := quick.Check(func(n int) bool {
// 				f := builder.NewFile(token.Pos(n), builder.Noop)

// 				return f.FileStart == token.Pos(n)
// 			}, nil)

// 			Expect(err).NotTo(HaveOccurred())
// 		})

// 		It("should build a rule", func() {
// 			f := builder.NewFile(token.Pos(1),
// 				file.Rule(rule.TextTarget("target")),
// 			)

// 			Expect(f).To(Equal(&ast.File{
// 				FileStart: token.Pos(1),
// 				FileEnd:   token.Pos(8),
// 				Contents: []ast.Obj{&ast.Rule{
// 					Targets: []ast.Expr{&ast.Text{
// 						Value:    "target",
// 						ValuePos: token.Pos(1),
// 					}},
// 					Colon: token.Pos(7),
// 				}},
// 			}))
// 			ExpectFprintToEqual(f, "target:\n")
// 		})

// 		It("should build a rule with multiple targets", func() {
// 			f := builder.NewFile(token.Pos(1),
// 				file.Rule(rule.TextTarget("target"),
// 					rule.TextTarget("target2"),
// 				),
// 			)

// 			Expect(f).To(Equal(&ast.File{
// 				FileStart: token.Pos(1),
// 				FileEnd:   token.Pos(16),
// 				Contents: []ast.Obj{&ast.Rule{
// 					Targets: []ast.Expr{
// 						&ast.Text{Value: "target", ValuePos: token.Pos(1)},
// 						&ast.Text{Value: "target2", ValuePos: token.Pos(8)},
// 					},
// 					Colon: token.Pos(15),
// 				}},
// 			}))
// 			ExpectFprintToEqual(f, "target target2:\n")
// 		})

// 		It("should build a rule with a target expression", func() {
// 			f := builder.NewFile(token.Pos(1),
// 				file.Rule(rule.VarRefTarget("FOO")),
// 			)

// 			Expect(f).To(Equal(&ast.File{
// 				FileStart: token.Pos(1),
// 				FileEnd:   token.Pos(8),
// 				Contents: []ast.Obj{&ast.Rule{
// 					Targets: []ast.Expr{&ast.VarRef{
// 						Dollar: token.Pos(1),
// 						Open:   token.LBRACE,
// 						Name:   "FOO",
// 						Close:  token.RBRACE,
// 					}},
// 					Colon: token.Pos(7),
// 				}},
// 			}))
// 			ExpectFprintToEqual(f, "${FOO}:\n")
// 		})

// 		It("should build a rule with a second target expression", func() {
// 			f := builder.NewFile(token.Pos(1),
// 				file.Rule(rule.TextTarget("target"),
// 					rule.VarRefTarget("FOO"),
// 				),
// 			)

// 			Expect(f).To(Equal(&ast.File{
// 				FileStart: token.Pos(1),
// 				FileEnd:   token.Pos(15),
// 				Contents: []ast.Obj{&ast.Rule{
// 					Targets: []ast.Expr{
// 						&ast.Text{Value: "target", ValuePos: token.Pos(1)},
// 						&ast.VarRef{
// 							Dollar: token.Pos(8),
// 							Open:   token.LBRACE,
// 							Name:   "FOO",
// 							Close:  token.RBRACE,
// 						},
// 					},
// 					Colon: token.Pos(14),
// 				}},
// 			}))
// 			ExpectFprintToEqual(f, "target ${FOO}:\n")
// 		})

// 		It("should insert a rule into an empty file", func() {
// 			f := builder.NewFile(token.Pos(1), func(f build.File) {
// 				f.InsertRule(0, rule.TextTarget("target"))
// 			})

// 			Expect(f).To(Equal(&ast.File{
// 				FileStart: token.Pos(1),
// 				FileEnd:   token.Pos(8),
// 				Contents: []ast.Obj{&ast.Rule{
// 					Targets: []ast.Expr{
// 						&ast.Text{Value: "target", ValuePos: token.Pos(1)},
// 					},
// 					Colon: token.Pos(7),
// 				}},
// 			}))
// 			ExpectFprintToEqual(f, "target:\n")
// 		})

// 		It("should insert a rule before an existing rule", func() {
// 			f := builder.NewFile(token.Pos(1), func(f build.File) {
// 				f.AddRule(rule.TextTarget("target"))
// 				f.InsertRule(0, rule.TextTarget("target2"))
// 			})

// 			Expect(f).To(Equal(&ast.File{
// 				FileStart: token.Pos(1),
// 				FileEnd:   token.Pos(16),
// 				Contents: []ast.Obj{
// 					&ast.Rule{
// 						Targets: []ast.Expr{
// 							&ast.Text{Value: "target2", ValuePos: token.Pos(1)},
// 						},
// 						Colon: token.Pos(8),
// 					},
// 					&ast.Rule{
// 						Targets: []ast.Expr{
// 							&ast.Text{Value: "target", ValuePos: token.Pos(10)},
// 						},
// 						Colon: token.Pos(15),
// 					},
// 				},
// 			}))
// 			ExpectFprintToEqual(f, "target2:\ntarget:\n")
// 		})
// 	})

// 	Describe("NewRule", func() {
// 		It("should build a new Rule", func() {
// 			r := builder.NewRule(token.Pos(1),
// 				rule.TextTarget("target"),
// 			)

// 			Expect(r).To(Equal(&ast.Rule{
// 				Targets: []ast.Expr{&ast.Text{
// 					Value:    "target",
// 					ValuePos: token.Pos(1),
// 				}},
// 				Colon: token.Pos(7),
// 			}))
// 			ExpectFprintToEqual(r, "target:\n")
// 		})

// 		It("should build a Rule with multiple targets", func() {
// 			r := builder.NewRule(token.Pos(1),
// 				rule.TextTarget("target"),
// 				rule.VarRefTarget("FOO"),
// 			)

// 			Expect(r).To(Equal(&ast.Rule{
// 				Targets: []ast.Expr{
// 					&ast.Text{
// 						Value:    "target",
// 						ValuePos: token.Pos(1),
// 					},
// 					&ast.VarRef{
// 						Dollar: token.Pos(8),
// 						Open:   token.LBRACE,
// 						Name:   "FOO",
// 						Close:  token.RBRACE,
// 					},
// 				},
// 				Colon: token.Pos(14),
// 			}))
// 			ExpectFprintToEqual(r, "target ${FOO}:\n")
// 		})
// 	})

// 	Describe("NewExpr", func() {
// 		It("should build a Text expression", func() {
// 			e := builder.NewExpr(token.Pos(1), expr.Text("test"))

// 			Expect(e).To(Equal(&ast.Text{
// 				Value:    "test",
// 				ValuePos: token.Pos(1),
// 			}))
// 			ExpectFprintToEqual(e, "test")
// 		})

// 		It("should build a VarRef expression", func() {
// 			e := builder.NewExpr(token.Pos(1), expr.VarRef("FOO"))

// 			Expect(e).To(Equal(&ast.VarRef{
// 				Dollar: token.Pos(1),
// 				Open:   token.LBRACE,
// 				Name:   "FOO",
// 				Close:  token.RBRACE,
// 			}))
// 			ExpectFprintToEqual(e, "${FOO}")
// 		})
// 	})
// })

// func ExpectFprintToEqual(x any, text string) {
// 	GinkgoHelper()
// 	buf := &bytes.Buffer{}
// 	Expect(make.Fprint(buf, x)).To(BeNumerically(">", 0))
// 	Expect(buf.String()).To(Equal(text))
// }
