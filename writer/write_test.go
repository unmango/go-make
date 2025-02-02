package writer_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/internal/testing"
	"github.com/unmango/go-make/token"
	"github.com/unmango/go-make/writer"
)

var _ = Describe("Write", func() {

	Describe("WriteRule", func() {
		DescribeTable("Rules",
			Entry("target",
				&ast.Rule{Targets: []ast.Expr{&ast.Text{Value: "target"}}},
				"target:\n",
			),
			Entry("multiple targets",
				&ast.Rule{Targets: []ast.Expr{
					&ast.Text{Value: "target", ValuePos: token.Pos(1)},
					&ast.Text{Value: "target2", ValuePos: token.Pos(8)},
				}},
				"target target2:\n",
			),
			Entry("target with prereq",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
					PreReqs: []ast.Expr{&ast.Text{
						Value:    "prereq",
						ValuePos: token.Pos(9),
					}},
				},
				"target: prereq\n",
			),
			Entry("target, prereq, and recipe",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
					PreReqs: []ast.Expr{&ast.Text{
						Value:    "prereq",
						ValuePos: token.Pos(9),
					}},
					Recipes: []*ast.Recipe{{
						Prefix:    token.TAB,
						PrefixPos: token.Pos(16),
						Text: ast.Text{
							Value:    "curl https://example.com",
							ValuePos: token.Pos(17),
						},
					}},
				},
				"target: prereq\n\tcurl https://example.com\n",
			),
			Entry("target with recipe",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
					Recipes: []*ast.Recipe{{
						Prefix:    token.TAB,
						PrefixPos: token.Pos(9),
						Text: ast.Text{
							Value:    "curl https://example.com",
							ValuePos: token.Pos(10),
						},
					}},
				},
				"target:\n\tcurl https://example.com\n",
			),
			func(r *ast.Rule, expected string) {
				buf := &bytes.Buffer{}
				w := writer.New(buf)

				n, err := writer.WriteRule(w, r)

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal(expected))
				Expect(n).To(Equal(len(expected)))
			},
		)

		It("should write multiple rules", func() {
			buf := &bytes.Buffer{}
			w := writer.New(buf)

			_, err := writer.WriteRule(w, &ast.Rule{Targets: []ast.Expr{
				&ast.Text{Value: "target"},
			}})
			Expect(err).NotTo(HaveOccurred())
			_, err = writer.WriteRule(w, &ast.Rule{Targets: []ast.Expr{
				&ast.Text{Value: "target2"},
			}})
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target:\ntarget2:\n"))
		})

		It("should ignore nil", func() {
			w := writer.New(&bytes.Buffer{})

			Expect(writer.WriteRule(w, nil)).To(Equal(0))
		})

		It("should return write errors", func() {
			w := writer.New(testing.NewErrAfterWriter(1))

			_, err := writer.WriteRule(w, &ast.Rule{
				Targets: []ast.Expr{&ast.Text{Value: "foo"}},
				PreReqs: []ast.Expr{&ast.Text{Value: "bar"}},
				Recipes: []*ast.Recipe{{
					Prefix: token.TAB,
					Text:   ast.Text{Value: "baz"},
				}},
			})

			Expect(err).To(MatchError("write err: 1"))
		})
	})

	Describe("WriteFile", func() {
		It("should write a Makefile", func() {
			buf := &bytes.Buffer{}
			w := writer.New(buf)

			_, err := writer.WriteFile(w, &ast.File{
				Contents: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
				}},
			})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should ignore nil", func() {
			w := writer.New(&bytes.Buffer{})

			Expect(writer.WriteFile(w, nil)).To(Equal(0))
		})

		It("should return errors found when writing a Makefile", func() {
			w := writer.New(testing.ErrWriter("io error"))

			_, err := writer.WriteFile(w, &ast.File{
				Contents: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
				}},
			})

			Expect(err).To(MatchError("io error"))
		})
	})

	Describe("Obj", func() {
		It("should ignore nil", func() {
			w := writer.New(&bytes.Buffer{})

			Expect(writer.Obj(w, nil)).To(Equal(0))
		})

		It("should write a variable", func() {
			w := writer.New(&bytes.Buffer{})

			n, err := writer.Obj(w, &ast.Variable{
				Name:  &ast.Text{Value: "TEST", ValuePos: token.Pos(1)},
				Op:    token.SIMPLE_ASSIGN,
				OpPos: token.Pos(6),
				Value: []ast.Expr{&ast.Text{
					Value:    "value",
					ValuePos: token.Pos(9),
				}},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(14))
		})
	})

	Describe("WriteVar", func() {
		It("should ignore nil variables", func() {
			w := writer.New(&bytes.Buffer{})

			Expect(writer.WriteVar(w, nil)).To(Equal(0))
		})

		When("Value is empty", func() {
			It("should write a variable", func() {
				buf := &bytes.Buffer{}
				w := writer.New(buf)

				n, err := writer.WriteVar(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST"},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(5),
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST:=\n"))
				Expect(n).To(Equal(7))
			})

			It("should write a space-separated variable", func() {
				buf := &bytes.Buffer{}
				w := writer.New(buf)

				n, err := writer.WriteVar(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST", ValuePos: token.Pos(1)},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(6),
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST :=\n"))
				Expect(n).To(Equal(8))
			})
		})

		When("Value is defined", func() {
			It("should write a variable", func() {
				buf := &bytes.Buffer{}
				w := writer.New(buf)

				n, err := writer.WriteVar(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST"},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(5),
					Value: []ast.Expr{&ast.Text{
						Value:    "value",
						ValuePos: token.Pos(6),
					}},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST:=value\n"))
				Expect(n).To(Equal(12))
			})

			It("should write a space-separated variable", func() {
				buf := &bytes.Buffer{}
				w := writer.New(buf)

				n, err := writer.WriteVar(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST"},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(6),
					Value: []ast.Expr{&ast.Text{
						Value:    "value",
						ValuePos: token.Pos(9),
					}},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST := value\n"))
				Expect(n).To(Equal(14))
			})

			It("should return write errors", func() {
				w := writer.New(
					testing.NewErrAfterWriter(1),
				)

				_, err := writer.WriteVar(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST"},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(6),
					Value: []ast.Expr{&ast.Text{
						Value:    "value",
						ValuePos: token.Pos(9),
					}},
				})

				Expect(err).To(MatchError(ContainSubstring("write err:")))
			})
		})
	})
})
