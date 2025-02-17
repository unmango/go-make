package file

import (
	"github.com/unmango/go-make/builder/build"
)

func Rule(fn func(build.Rule), fs ...func(build.Rule)) func(build.File) {
	return func(f build.File) {
		f.AddRule(fn, fs...)
	}
}
