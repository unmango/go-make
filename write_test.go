package make_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/internal/testing"
)

var _ = Describe("Write", func() {
	It("should write a line", func() {
		buf := &bytes.Buffer{}
		w := make.NewWriter(buf)

		n, err := w.WriteLine()

		Expect(err).NotTo(HaveOccurred())
		Expect(buf.String()).To(Equal("\n"))
		Expect(n).To(Equal(1))
	})

	It("should write a target", func() {
		buf := &bytes.Buffer{}
		w := make.NewWriter(buf)

		n, err := w.WriteTarget("target")

		Expect(err).NotTo(HaveOccurred())
		Expect(buf.String()).To(Equal("target:"))
		Expect(n).To(Equal(7))
	})

	It("should write multiple targets", func() {
		buf := &bytes.Buffer{}
		w := make.NewWriter(buf)

		n, err := w.WriteTargets([]string{"target", "target2"})

		Expect(err).NotTo(HaveOccurred())
		Expect(buf.String()).To(Equal("target target2:"))
		Expect(n).To(Equal(15))
	})

	Describe("WriteRule", func() {
		DescribeTable("Rules",
			Entry(nil,
				make.Rule{Target: []string{"target"}},
				"target:\n",
			),
			Entry(nil,
				make.Rule{Target: []string{"target", "target2"}},
				"target target2:\n",
			),
			Entry(nil,
				make.Rule{
					Target:  []string{"target"},
					PreReqs: []string{"prereq"},
				},
				"target: prereq\n",
			),
			Entry(nil,
				make.Rule{
					Target:  []string{"target"},
					PreReqs: []string{"prereq"},
					Recipe:  []string{"curl https://example.com"},
				},
				"target: prereq\n\tcurl https://example.com\n",
			),
			Entry(nil,
				make.Rule{
					Target: []string{"target"},
					Recipe: []string{"curl https://example.com"},
				},
				"target:\n\tcurl https://example.com\n",
			),
			func(r make.Rule, expected string) {
				buf := &bytes.Buffer{}
				w := make.NewWriter(buf)

				n, err := w.WriteRule(r)

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal(expected))
				Expect(n).To(Equal(len(expected)))
			},
		)

		It("should write multiple rules", func() {
			buf := &bytes.Buffer{}
			w := make.NewWriter(buf)

			_, err := w.WriteRule(make.Rule{Target: []string{"target"}})
			Expect(err).NotTo(HaveOccurred())
			_, err = w.WriteRule(make.Rule{Target: []string{"target2"}})
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target:\ntarget2:\n"))
		})

		DescribeTable("should error when rule has no targets",
			Entry("empty rule", make.Rule{}),
			Entry("with prereqs", make.Rule{PreReqs: []string{"foo"}}),
			Entry("with recipes", make.Rule{Recipe: []string{"foo"}}),
			func(rule make.Rule) {
				buf := &bytes.Buffer{}
				w := make.NewWriter(buf)

				_, err := w.WriteRule(rule)

				Expect(err).To(MatchError("no targets"))
			},
		)

		It("should return target errors", func() {
			writer := testing.NewErrAfterWriter(1)
			w := make.NewWriter(writer)

			_, err := w.WriteRule(make.Rule{
				Target: []string{"foo"},
			})

			Expect(err).To(MatchError("write err: 1"))
		})

		It("should return prereq errors", func() {
			writer := testing.NewErrAfterWriter(2)
			w := make.NewWriter(writer)

			_, err := w.WriteRule(make.Rule{
				Target:  []string{"foo"},
				PreReqs: []string{"bar"},
			})

			Expect(err).To(MatchError("write err: 2"))
			Expect(writer.Buf.String()).To(Equal("foo:"))
		})

		It("should return newline errors", func() {
			writer := testing.NewErrAfterWriter(3)
			w := make.NewWriter(writer)

			_, err := w.WriteRule(make.Rule{
				Target:  []string{"foo"},
				PreReqs: []string{"bar"},
			})

			Expect(err).To(MatchError("write err: 3"))
			Expect(writer.Buf.String()).To(Equal("foo: bar"))
		})

		It("should return recipe errors", func() {
			writer := testing.NewErrAfterWriter(4)
			w := make.NewWriter(writer)

			_, err := w.WriteRule(make.Rule{
				Target:  []string{"foo"},
				PreReqs: []string{"bar"},
				Recipe:  []string{"baz"},
			})

			Expect(err).To(MatchError("write err: 4"))
			Expect(writer.Buf.String()).To(Equal("foo: bar\n"))
		})

		It("should return trailing newline errors", func() {
			writer := testing.NewErrAfterWriter(5)
			w := make.NewWriter(writer)

			_, err := w.WriteRule(make.Rule{
				Target:  []string{"foo"},
				PreReqs: []string{"bar"},
				Recipe:  []string{"baz"},
			})

			Expect(err).To(MatchError("write err: 5"))
			Expect(writer.Buf.String()).To(Equal("foo: bar\n\tbaz"))
		})
	})

	It("should write a Makefile", func() {
		buf := &bytes.Buffer{}
		w := make.NewWriter(buf)

		_, err := w.WriteMakefile(make.Makefile{
			Rules: []make.Rule{{
				Target: []string{"target"},
			}},
		})

		Expect(err).NotTo(HaveOccurred())
	})

	It("should return errors found when writing a Makefile", func() {
		w := make.NewWriter(testing.ErrWriter("io error"))

		_, err := w.WriteMakefile(make.Makefile{
			Rules: []make.Rule{{
				Target: []string{"target"},
			}},
		})

		Expect(err).To(MatchError("io error"))
	})

	It("should return errors found when writing PreReqs", func() {
		w := make.NewWriter(testing.ErrWriter("io error"))

		_, err := w.WritePreReqs([]string{"blah"})

		Expect(err).To(MatchError("io error"))
	})

	It("should return errors found when writing Recipes", func() {
		w := make.NewWriter(testing.ErrWriter("io error"))

		_, err := w.WriteRecipes([]string{"blah"})

		Expect(err).To(MatchError("io error"))
	})
})
