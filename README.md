# Go Make

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/unmango/go-make/ci.yml)
![GitHub branch check runs](https://img.shields.io/github/check-runs/unmango/go-make/main)
![Libraries.io dependency status for GitHub repo](https://img.shields.io/librariesio/github/unmango/go-make)
![Codecov](https://img.shields.io/codecov/c/github/unmango/go-make)
![GitHub Release](https://img.shields.io/github/v/release/unmango/go-make)
![GitHub Release Date](https://img.shields.io/github/release-date/unmango/go-make)

Makefile parsing and utilities in Go

## Usage

### Reading

The `make.Parser` is the primary way to read Makefiles.

```go
f := os.Open("Makefile")
p := make.NewParser(f, nil)

m, err := p.ParseFile()

fmt.Println(m.Rules)
```

The more primitive `make.Scanner` and `make.ScanTokens` used by `make.Parser` can be used individually.

Using `make.ScanTokens` with a `bufio.Scanner`

```go
f := os.Open("Makefile")
s := bufio.NewScanner(f)
s.Split(make.ScanTokens)

for s.Scan() {
  s.Bytes() // The current token byte slice i.e. []byte(":=")
  s.Text() // The current token as a string i.e. ":="
}
```

Using `make.Scanner`

```go
f := os.Open("Makefile")
s := make.NewScanner(f, nil)

for pos, tok, lit := s.Scan(); tok != token.EOF; {
  fmt.Println(pos) // The position of tok
  fmt.Println(tok) // The current token.Token i.e. token.SIMPLE_ASSIGN
  fmt.Println(lit) // Literal tokens as a string i.e. "identifier"
}

if err := s.Err(); err != nil {
  fmt.Println(err)
}
```

### Writing

Use `make.Fprint` to write ast nodes.

> **Note**
> The AST in this project is a made-up, package-specific representation for Makefiles. It is not an official GNU Make or POSIX AST.

```go
var file *ast.File

n, err := make.Fprint(os.Stdout, file)
```

The `make.Writer` can be used to incrementally write make syntax to an `io.Writer`.

```go
buf := &bytes.Buffer{}
w := make.NewWriter(buf)

n, err := w.WriteRule(&ast.Rule{})
```

### Builder

The `builder` package contains utilities for building AST nodes.

🚧 This API is not stable yet 🚧

```go
f := builder.NewFile(1,
  file.WithRule(expr.Text("target1"),
    rule.WithVarRefTarget("FOO")
  ),
)

make.Fprint(os.Stdout, f)
// target1 ${FOO}:\n
```

## Features

### Syntax Support

Makefile syntax that is guaranteed to round-trip (parse and print without modification) is listed in [./testdata/roundtrip](./testdata/roundtrip/).
Additional syntax is supported and may round-trip successfully, but no guarentees are provided until it is listed under `./testdata/roundtrip`.

The table below covers the syntax in the GNU Make [Quick Reference](https://www.gnu.org/software/make/manual/html_node/Quick-Reference.html).
Every status is verified by parsing the example, printing it, and comparing the result against the input, rather than by reading the code.

|       Status       | Meaning                                                                                                 |
| :----------------: | :------------------------------------------------------------------------------------------------------ |
| :white_check_mark: | Modeled by a dedicated AST node and round-trips                                                         |
|     :warning:      | Round-trips, but is stored as plain text or produces a structure that does not match make's own reading |
|      (blank)       | Not supported, see the remark for current behavior                                                      |

| Syntax                            | Example                                         |       Parser       |      Printer       |      Builder       | Remarks                                                                                                                                                           |
| --------------------------------- | ----------------------------------------------- | :----------------: | :----------------: | :----------------: | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **general**                       |                                                 |                    |                    |                    |                                                                                                                                                                   |
| empty file                        |                                                 | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| blank lines between elements      | `target:\n\n\ntarget2:`                         | :white_check_mark: | :white_check_mark: |                    | blank lines are recreated from stored positions                                                                                                                   |
| leading blank lines               | `\n\ntarget:`                                   | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| file of blank lines only          | `\n`                                            |                    |                    |                    | prints as an empty file                                                                                                                                           |
| no trailing newline               | `target: prereq`                                |     :warning:      |     :warning:      |                    | the printer always terminates the final line with `\n`                                                                                                            |
| trailing whitespace               | `target: prereq  \n`                            |                    |                    |                    | dropped, and a file ending in spaces scans as `UNSUPPORTED`                                                                                                       |
| CRLF line endings                 | `target: prereq\r\n`                            |                    |                    |                    | text adjacent to `\r` is silently dropped by the scanner, [#113](https://github.com/unmango/go-make/issues/113)                                                   |
| escaped `#`                       | `target: foo\#bar`                              | :white_check_mark: | :white_check_mark: |                    | kept verbatim, no unescaping is performed                                                                                                                         |
| line continuation, recipe         | `\techo x \\\n\t\ty`                            |     :warning:      |     :warning:      |                    | each physical line becomes its own `Recipe`                                                                                                                       |
| line continuation, prerequisites  | `target: b \\\n\tc`                             |     :warning:      |     :warning:      |                    | the `\\` is a prerequisite and the next line becomes a recipe                                                                                                     |
| line continuation, variable value | `VAR := b \\\n\tc`                              |                    |                    |                    | parses to a `nil` entry, [#139](https://github.com/unmango/go-make/issues/139) [#111](https://github.com/unmango/go-make/issues/111)                              |
| **comments**                      |                                                 |                    |                    |                    |                                                                                                                                                                   |
| top-level comments                | `# comment text`                                | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| comment groups                    | `# comment text\n# more comment text`           | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| groups separated by a blank line  | `# a\n\n# b`                                    | :white_check_mark: | :white_check_mark: |                    | each group is a separate `CommentGroup`                                                                                                                           |
| comments with no leading space    | `#comment text`                                 |     :warning:      |     :warning:      |                    | parses, but the printer always inserts a space after `#`, [#115](https://github.com/unmango/go-make/issues/115)                                                   |
| recipe comments                   | `target:\n\trecipe # comment text`              | :white_check_mark: | :white_check_mark: |                    | these are not make comments and are included in the recipe text                                                                                                   |
| rule comments                     | `target: # comment text`                        |                    |                    |                    | parse error: `expected one of 'TEXT', '$', found 'COMMENT'`                                                                                                       |
| comments after a word             | `prereq# comment text`                          |                    |                    |                    | `#` only starts a comment at the start of a token, [#112](https://github.com/unmango/go-make/issues/112)                                                          |
| **rules**                         |                                                 |                    |                    |                    |                                                                                                                                                                   |
| targets                           | `target:`, `target :`                           | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                                                                                                                   |
| multiple targets                  | `target1 target2:`                              | :white_check_mark: | :white_check_mark: | :white_check_mark: | the builder places every target at the same position, [#120](https://github.com/unmango/go-make/issues/120)                                                       |
| pre-requisites                    | `target: prereq`                                | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| order-only pre-requisites         | `target: \| prereq`                             | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| order-only, unseparated           | `target: prereq\|order-only`                    |                    |                    |                    | `\|` only delimits at the start of a token, [#112](https://github.com/unmango/go-make/issues/112)                                                                 |
| pattern rules                     | `%.o: %.c`                                      | :white_check_mark: | :white_check_mark: |                    | `%` is ordinary text, the pattern itself is not modeled                                                                                                           |
| suffix rules                      | `.c.o:`                                         | :white_check_mark: | :white_check_mark: |                    | ordinary target text                                                                                                                                              |
| special targets                   | `.PHONY: target`                                | :white_check_mark: | :white_check_mark: |                    | ordinary target text, see `ast/target` for the known names                                                                                                        |
| wildcard pre-requisites           | `target: *.c`                                   | :white_check_mark: | :white_check_mark: |                    | ordinary prerequisite text                                                                                                                                        |
| static pattern rules              | `a.o: %.o: %.c`                                 |                    |                    |                    | parse error on the second `:`                                                                                                                                     |
| double-colon rules                | `target:: prereq`                               |                    |                    |                    | parse error on the second `:`                                                                                                                                     |
| grouped targets                   | `a b &: prereq`                                 |     :warning:      |     :warning:      |                    | round-trips, but `&` is parsed as an additional target                                                                                                            |
| escaped spaces in targets         | `a\\ b:`                                        |     :warning:      |     :warning:      |                    | round-trips, but yields two targets rather than one                                                                                                               |
| target-specific variables         | `target: VAR = value`                           |                    |                    |                    | parse error on `=`                                                                                                                                                |
| pattern-specific variables        | `%.o: VAR = value`                              |                    |                    |                    | parse error on `=`                                                                                                                                                |
| **recipes**                       |                                                 |                    |                    |                    |                                                                                                                                                                   |
| recipes                           | `target:\n\trecipe text`                        | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| multiple recipe lines             | `target:\n\tone\n\ttwo`                         | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| recipe prefix characters          | `\t@echo`, `\t-rm x`, `\t+$(MAKE)`              | :white_check_mark: | :white_check_mark: |                    | part of the recipe text                                                                                                                                           |
| escaped `$$` in recipes           | `\techo $$HOME`                                 | :white_check_mark: | :white_check_mark: |                    | recipe bodies are captured verbatim                                                                                                                               |
| automatic variables in recipes    | `\tcp $< $@`                                    |     :warning:      |     :warning:      |                    | verbatim text rather than `VarRef` nodes                                                                                                                          |
| variable references in recipes    | `\techo $(VAR)`                                 |     :warning:      |     :warning:      |                    | verbatim text rather than `VarRef` nodes                                                                                                                          |
| blank line inside a recipe block  | `target:\n\tone\n\n\ttwo`                       |                    |                    |                    | ends the rule and parses to a `nil` entry, [#139](https://github.com/unmango/go-make/issues/139) [#111](https://github.com/unmango/go-make/issues/111)            |
| semicolon delimited recipes       | `target: ;recipe text`                          |                    |                    |                    | scanned as pre-requisite text, and `target: ; recipe` is a parse error, [#127](https://github.com/unmango/go-make/issues/127)                                     |
| custom `.RECIPEPREFIX`            | `.RECIPEPREFIX = >\n>recipe`                    |                    |                    |                    | the prefix is fixed to `TAB`, [#127](https://github.com/unmango/go-make/issues/127)                                                                               |
| **variables**                     |                                                 |                    |                    |                    |                                                                                                                                                                   |
| assignment operators              | `=`, `:=`, `::=`, `:::=`, `?=`, `!=`            | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| append assignment                 | `VAR += x`                                      |                    |                    |                    | `+=` is not tokenized at all, [#137](https://github.com/unmango/go-make/issues/137)                                                                               |
| empty declarations                | `VAR :=`                                        | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| multi-word values                 | `VAR := foo.c bar.c`                            | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| unspaced `:`-prefixed operators   | `VAR:=x`, `VAR::=x`, `VAR:::=x`                 | :white_check_mark: | :white_check_mark: |                    | `:` delimits, so these scan correctly                                                                                                                             |
| unspaced `=`, `?=`, `!=`, `+=`    | `VAR=x`, `VAR?=x`, `VAR!=x`                     |                    |                    |                    | scanned as one `TEXT` token, parses to a `nil` entry, [#112](https://github.com/unmango/go-make/issues/112) [#111](https://github.com/unmango/go-make/issues/111) |
| definition blocks                 | `define FOO\nbody\nendef`                       |                    |                    |                    | no AST node, parses to a `nil` entry, [#111](https://github.com/unmango/go-make/issues/111)                                                                       |
| `undefine`                        | `undefine VAR`                                  |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                |
| `override`                        | `override VAR := x`                             |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                |
| `export` and `unexport`           | `export VAR`, `export VAR := x`                 |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                |
| `private`                         | `private VAR := x`                              |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                |
| **variable references**           |                                                 |                    |                    |                    |                                                                                                                                                                   |
| parentheses and braces            | `$(VAR)`, `${VAR}`                              | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                                                                                                                   |
| single character                  | `$V`                                            | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| in targets                        | `${VAR}:`, `$(FOO) $(BAR):`                     | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                                                                                                                   |
| in pre-requisites                 | `target: ${FOO}`                                | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| in variable values                | `A := $(B)`                                     | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| in conditions                     | `ifeq ($(V),x)`                                 | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| escaped `$$`                      | `target: $$V`                                   |                    |                    |                    | parsed as a reference and printed as `$_$V`, [#138](https://github.com/unmango/go-make/issues/138)                                                                |
| adjacent to text                  | `target: prefix$(FOO)`                          |                    |                    |                    | `$` only starts a reference at the start of a token, [#112](https://github.com/unmango/go-make/issues/112)                                                        |
| substitution references           | `$(VAR:.c=.o)`                                  |                    |                    |                    | parse error on `:`                                                                                                                                                |
| computed names                    | `$($(VAR))`                                     |                    |                    |                    | parse error on the inner `$`                                                                                                                                      |
| names matching a builtin          | `$(dir)`, `$(file)`, `$(word)`                  |                    |                    |                    | builtin names scan as keywords, so the parser reports `expected 'TEXT'`                                                                                           |
| **functions**                     |                                                 |                    |                    |                    |                                                                                                                                                                   |
| text and filename functions       | `$(subst a,b,$(S))`, `$(wildcard *.c)`          |                    |                    |                    | `token` enumerates every builtin, but the parser rejects them                                                                                                     |
| `$(shell ...)`                    | `A := $(shell echo x)`                          |                    |                    |                    |                                                                                                                                                                   |
| `$(call ...)` and `$(eval ...)`   | `A := $(call f,x)`                              |                    |                    |                    |                                                                                                                                                                   |
| control functions                 | `$(foreach v,$(L),$(v))`, `$(if $(V),x,y)`      |                    |                    |                    |                                                                                                                                                                   |
| logging functions                 | `$(info msg)`, `$(warning msg)`, `$(error msg)` |                    |                    |                    |                                                                                                                                                                   |
| **conditionals**                  |                                                 |                    |                    |                    |                                                                                                                                                                   |
| equality directives               | `ifeq`, `ifneq`                                 | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| definition directives             | `ifdef`, `ifndef`                               | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| parentheses syntax                | `ifeq (foo, bar)`                               | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| double quotes                     | `ifeq "foo" "bar"`                              | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| single quotes                     | `ifeq 'foo' 'bar'`                              | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| mixed quotes                      | `ifeq "foo" 'bar'`                              | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| `else`                            | `ifdef V\na:\nelse\nb:\nendif`                  | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| `else if` chains                  | `else ifdef W`, `else ifeq (a, b)`              | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| nested conditionals               | `ifdef V\nifdef W\nendif\nendif`                | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| wrapping variables or rules       | `ifdef V\nA := x\nendif`                        | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                   |
| inside a recipe block             | `target:\nifdef V\n\techo x\nendif`             |                    |                    |                    | the recipe `TAB` is consumed as whitespace, [#139](https://github.com/unmango/go-make/issues/139) [#111](https://github.com/unmango/go-make/issues/111)           |
| **other directives**              |                                                 |                    |                    |                    |                                                                                                                                                                   |
| include directives                | `include a.mk`, `-include`, `sinclude`          |                    |                    |                    | no AST node, parses to a `nil` entry, [#111](https://github.com/unmango/go-make/issues/111)                                                                       |
| `vpath` directive                 | `vpath %.c src`                                 |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                |
| `VPATH` variable                  | `VPATH := src`                                  | :white_check_mark: | :white_check_mark: |                    | an ordinary variable assignment                                                                                                                                   |
| `load` directive                  | `load a.so`                                     |                    |                    |                    | not tokenized, parses to a `nil` entry, [#111](https://github.com/unmango/go-make/issues/111)                                                                     |
| many other things                 |                                                 |                    |                    |                    | please open an issue if there is anything missing you'd like to see!                                                                                              |

Constructs marked unsupported either report a parse error or append a `nil` element to `ast.File.Contents`.
A `nil` element is returned with no error and panics when printed, tracked in [#111](https://github.com/unmango/go-make/issues/111).

### Reference Coverage

The names make understands are enumerated by [token](./token/token.go), [ast/target](./ast/target/target.go), and [ast/variable](./ast/variable/variable.go).
`internal/conformance` compares each enumeration against the summaries published in the GNU Make manual, so syntax added to make cannot go unnoticed.

| Enumeration        | go-make | Manual | Source                                                                                                   |
| ------------------ | ------: | -----: | -------------------------------------------------------------------------------------------------------- |
| directives         |      17 |     17 | [Quick Reference](https://www.gnu.org/software/make/manual/html_node/Quick-Reference.html)               |
| built-in functions |      37 |     37 | [Quick Reference](https://www.gnu.org/software/make/manual/html_node/Quick-Reference.html)               |
| special variables  |      14 |     14 | [Quick Reference](https://www.gnu.org/software/make/manual/html_node/Quick-Reference.html)               |
| special targets    |      17 |     17 | [Special Built-in Target Names](https://www.gnu.org/software/make/manual/html_node/Special-Targets.html) |

The comparison runs against `internal/conformance/testdata/quickref.json`, a fixture extracted from those two pages, so the suite needs no network access.
Run `make sync-quickref` to refresh it.
The manual pages are pinned by content hash in [flake.nix](./flake.nix) and fetched with `fetchurl`, so a refresh either reproduces the same input or fails with a hash mismatch reporting that the manual changed.
A weekly workflow runs the refresh and fails if the fixture is out of date, so drift in the manual surfaces without anyone rerunning it by hand.
It is not part of the per-commit checks, because it reaches gnu.org and an outage there is not a reason to block a merge.

Automatic variables (`$@`, `$<`, `$(@D)`, and the rest) are summarized by the manual but are not enumerated here, so they are excluded from the comparison.

### Will Not Support

Nothing, at this time

## Workflow

### Pre-Requisites

Go toolchain for the version listed in [go.mod](./go.mod)

### Building

go-make is itself built using `make`.

|         Targets | Description                                                                              |
| --------------: | :--------------------------------------------------------------------------------------- |
|    default goal | Runs the `build` target                                                                  |
|         `build` | Runs `go build` to verify the code compiles                                              |
|          `test` | Test changed packages                                                                    |
|      `test_all` | Test all packages                                                                        |
| `sync-quickref` | Refresh the pinned GNU Make manual fixture used by `internal/conformance` (requires nix) |
|         `clean` | Remove `.make` directory and coverage report                                             |
|         `cover` | Collect coverage for all tests and print report                                          |
|          `tidy` | Runs `go mod tidy`                                                                       |
|           `dev` | Setup the [developer environment](#developer-environment)                                |

### Developer Environment

Apart from the Go toolchain, the only main dependency is the `ginkgo` cli to run tests.
This repo also uses [devctl](https://github.com/unmango/devctl) but its use is optional.
Targets will obtain dependencies automatically as needed.

Binaries are stored in a `.gitignore`d `bin/` directory at the root of the repository.
An example `.envrc` file for [direnv](https://github.com/direnv/direnv) is provided in [hack/example.envrc](./hack/example.envrc) to add `./bin` to your `PATH` automatically.
To use it, run `make .envrc` or `make dev`.
This will copy `hack/example.envrc` to `.envrc` at the root of the repository.

## References

GNU Make Quick Reference: <https://www.gnu.org/software/make/manual/html_node/Quick-Reference.html>
