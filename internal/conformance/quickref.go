// Package conformance compares the make syntax enumerated by this module
// against the summaries published in the GNU Make manual.
//
// The manual is not fetched at test time. [Parse] extracts the summaries from
// the manual's HTML, and cmd/syncquickref writes the result to
// testdata/quickref.json, which the tests read.
package conformance

import (
	"fmt"
	"html"
	"regexp"
	"slices"
	"strings"
)

// Manual pages summarizing the syntax make understands.
const (
	QuickRefURL       = "https://www.gnu.org/software/make/manual/html_node/Quick-Reference.html"
	SpecialTargetsURL = "https://www.gnu.org/software/make/manual/html_node/Special-Targets.html"
)

// QuickRef holds the names make documents, keyed the way this module
// enumerates them. Automatic variables are omitted because no package here
// enumerates them.
type QuickRef struct {
	Sources          []string `json:"sources"`
	Directives       []string `json:"directives"`
	Functions        []string `json:"functions"`
	SpecialVariables []string `json:"specialVariables"`
	SpecialTargets   []string `json:"specialTargets"`
}

var (
	// <dt> entries carry one syntax summary each, i.e.
	// <dt><span><code>define <var>variable</var></code></span></dt>
	dtPattern     = regexp.MustCompile(`(?s)<dt>.*?<code>(.*?)</code>`)
	tagPattern    = regexp.MustCompile(`<[^>]+>`)
	targetPattern = regexp.MustCompile(`<code>(\.[A-Z_]+)</code>`)
	funcPattern   = regexp.MustCompile(`^\$\(([a-z][a-z-]*)[\s,)]`)
	varPattern    = regexp.MustCompile(`^\.?[A-Z][A-Z_]*$`)
	directivePfx  = regexp.MustCompile(`^-?[a-z]+`)
)

// Parse extracts the summarized names from the Quick Reference and Special
// Built-in Target Names pages.
func Parse(quickRef, specialTargets []byte) (*QuickRef, error) {
	ref := &QuickRef{Sources: []string{QuickRefURL, SpecialTargetsURL}}

	for _, m := range dtPattern.FindAllStringSubmatch(string(quickRef), -1) {
		syntax := text(m[1])
		switch {
		case syntax == "":
		case funcPattern.MatchString(syntax):
			ref.Functions = append(ref.Functions, funcPattern.FindStringSubmatch(syntax)[1])
		case strings.HasPrefix(syntax, "$"):
			// an automatic variable, i.e. $@ or $(@D)
		case varPattern.MatchString(syntax):
			ref.SpecialVariables = append(ref.SpecialVariables, syntax)
		case directivePfx.MatchString(syntax):
			ref.Directives = append(ref.Directives, directivePfx.FindString(syntax))
		}
	}

	for _, m := range targetPattern.FindAllStringSubmatch(string(specialTargets), -1) {
		ref.SpecialTargets = append(ref.SpecialTargets, m[1])
	}

	for _, l := range []*[]string{&ref.Directives, &ref.Functions, &ref.SpecialVariables, &ref.SpecialTargets} {
		slices.Sort(*l)
		*l = slices.Compact(*l)
		if len(*l) == 0 {
			return nil, fmt.Errorf("no names extracted, the manual's markup may have changed")
		}
	}

	return ref, nil
}

// text strips markup from a <code> body and collapses its whitespace.
func text(s string) string {
	return strings.Join(strings.Fields(html.UnescapeString(tagPattern.ReplaceAllString(s, ""))), " ")
}
