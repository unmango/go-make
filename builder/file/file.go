package file

import (
	"github.com/unmango/go-make/builder/build"
)

func WithRule(e func(build.Expr), rs ...func(build.Rule)) func(build.File) {
	return func(f build.File) {
		f.Rule(e, rs...)
	}
}
