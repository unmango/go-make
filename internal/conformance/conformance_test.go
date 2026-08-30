package conformance_test

import (
	_ "embed"
	"encoding/json"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast/target"
	"github.com/unmango/go-make/ast/variable"
	"github.com/unmango/go-make/internal/conformance"
	"github.com/unmango/go-make/token"
)

//go:embed testdata/quickref.json
var fixture []byte

// ref holds the names the GNU Make manual summarizes. Refresh it with
// `make sync-quickref`.
var ref = func() conformance.QuickRef {
	r := conformance.QuickRef{}
	if err := json.Unmarshal(fixture, &r); err != nil {
		panic(err)
	}
	return r
}()

// enumerated collects the names matching pred from the token constants.
func enumerated(pred func(token.Token) bool) (names []string) {
	for i := range 256 {
		if tok := token.Token(i); pred(tok) {
			names = append(names, tok.String())
		}
	}
	slices.Sort(names)
	return
}

func entries(names []string) []TableEntry {
	e := make([]TableEntry, len(names))
	for i, n := range names {
		e[i] = Entry(n, n)
	}
	return e
}

var _ = Describe("Quick Reference", func() {
	Describe("directives", func() {
		DescribeTable("should be recognized", entries(ref.Directives),
			func(name string) {
				Expect(token.IsDirective(name)).To(BeTrue())
			},
		)

		It("should not recognize directives the manual does not list", func() {
			Expect(enumerated(token.Token.IsDirective)).To(ConsistOf(ref.Directives))
		})
	})

	Describe("built-in functions", func() {
		DescribeTable("should be recognized", entries(ref.Functions),
			func(name string) {
				Expect(token.IsBuiltinFunction(name)).To(BeTrue())
			},
		)

		It("should not recognize functions the manual does not list", func() {
			Expect(enumerated(token.Token.IsBuiltinFunction)).To(ConsistOf(ref.Functions))
		})
	})

	Describe("special variables", func() {
		DescribeTable("should be enumerated", entries(ref.SpecialVariables),
			func(name string) {
				Expect(variable.Special).To(ContainElement(name))
			},
		)

		It("should not enumerate variables the manual does not list", func() {
			Expect(variable.Special).To(ConsistOf(ref.SpecialVariables))
		})
	})

	Describe("special targets", func() {
		DescribeTable("should be enumerated", entries(ref.SpecialTargets),
			func(name string) {
				Expect(target.Builtin).To(ContainElement(name))
			},
		)

		It("should not enumerate targets the manual does not list", func() {
			Expect(target.Builtin).To(ConsistOf(ref.SpecialTargets))
		})
	})

	Describe("Parse", func() {
		const quickRef = `
			<dl compact="compact">
			<dt><span><code>define <var>variable</var></code></span></dt>
			<dt><span><code>define <var>variable</var> :=</code></span></dt>
			<dt><span><code>-include <var>file</var></code></span></dt>
			<dt><span><code>$(subst <var>from</var>,<var>to</var>,<var>text</var>)</code></span></dt>
			<dt><span><code>$(words <var>text</var>)</code></span></dt>
			<dt><span><code>$@</code></span></dt>
			<dt><span><code>$(@D)</code></span></dt>
			<dt><span><code>MAKEFLAGS</code></span></dt>
			<dt><span><code>.LIBPATTERNS</code></span></dt>
			</dl>`
		const specialTargets = `<dt><code>.PHONY</code></dt><dt><code>.WAIT</code></dt>`

		It("should extract each kind of name", func() {
			r, err := conformance.Parse([]byte(quickRef), []byte(specialTargets))

			Expect(err).NotTo(HaveOccurred())
			Expect(r.Directives).To(Equal([]string{"-include", "define"}))
			Expect(r.Functions).To(Equal([]string{"subst", "words"}))
			Expect(r.SpecialVariables).To(Equal([]string{".LIBPATTERNS", "MAKEFLAGS"}))
			Expect(r.SpecialTargets).To(Equal([]string{".PHONY", ".WAIT"}))
		})

		It("should record where the names came from", func() {
			r, err := conformance.Parse([]byte(quickRef), []byte(specialTargets))

			Expect(err).NotTo(HaveOccurred())
			Expect(r.Sources).To(ConsistOf(conformance.QuickRefURL, conformance.SpecialTargetsURL))
		})

		It("should error when the markup yields no names", func() {
			_, err := conformance.Parse([]byte("<p>no definition lists here</p>"), []byte(specialTargets))

			Expect(err).To(MatchError(ContainSubstring("no names extracted")))
		})
	})
})
