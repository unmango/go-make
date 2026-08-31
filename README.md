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

| Syntax                            | Example                                         |       Parser       |      Printer       |      Builder       | Remarks                                                                                                                       |
| --------------------------------- | ----------------------------------------------- | :----------------: | :----------------: | :----------------: | ----------------------------------------------------------------------------------------------------------------------------- |
| **general**                       |                                                 |                    |                    |                    |                                                                                                                               |
| empty file                        |                                                 | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| blank lines between elements      | `target:\n\n\ntarget2:`                         | :white_check_mark: | :white_check_mark: |                    | blank lines are recreated from stored positions                                                                               |
| leading blank lines               | `\n\ntarget:`                                   | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| file of blank lines only          | `\n`                                            |                    |                    |                    | prints as an empty file                                                                                                       |
| no trailing newline               | `target: prereq`                                |     :warning:      |     :warning:      |                    | the printer always terminates the final line with `\n`                                                                        |
| trailing whitespace               | `target: prereq  \n`                            |                    |                    |                    | dropped by the scanner, so it is absent from printed output, with or without a trailing newline                               |
| CRLF line endings                 | `target: prereq\r\n`                            | :white_check_mark: | :white_check_mark: |                    | the line ending is recorded once per file, a file that mixes endings is normalized to its dominant ending                     |
| escaped `#`                       | `target: foo\#bar`                              | :white_check_mark: | :white_check_mark: |                    | kept verbatim, no unescaping is performed                                                                                     |
| line continuation, recipe         | `\techo x \\\n\t\ty`                            |     :warning:      |     :warning:      |                    | each physical line becomes its own `Recipe`                                                                                   |
| line continuation, prerequisites  | `target: b \\\n\tc`                             |     :warning:      |     :warning:      |                    | the `\\` is a prerequisite and the next line becomes a recipe                                                                 |
| line continuation, variable value | `VAR := b \\\n\tc`                              |     :warning:      |     :warning:      |                    | the continued line becomes a `BadObj`, [#139](https://github.com/unmango/go-make/issues/139)                                  |
| **comments**                      |                                                 |                    |                    |                    |                                                                                                                               |
| top-level comments                | `# comment text`                                | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| comment groups                    | `# comment text\n# more comment text`           | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| groups separated by a blank line  | `# a\n\n# b`                                    | :white_check_mark: | :white_check_mark: |                    | each group is a separate `CommentGroup`                                                                                       |
| comments with no leading space    | `#comment text`                                 | :white_check_mark: | :white_check_mark: |                    | `Comment.Text` keeps the text after `#` verbatim, so the original spacing round-trips                                         |
| recipe comments                   | `target:\n\trecipe # comment text`              | :white_check_mark: | :white_check_mark: |                    | these are not make comments and are included in the recipe text                                                               |
| rule comments                     | `target: # comment text`                        |                    |                    |                    | parse error: `expected one of 'TEXT', '$', found 'COMMENT'`                                                                   |
| comments after a word             | `prereq# comment text`                          |                    |                    |                    | `#` only starts a comment at the start of a token, [#112](https://github.com/unmango/go-make/issues/112)                      |
| **rules**                         |                                                 |                    |                    |                    |                                                                                                                               |
| targets                           | `target:`, `target :`                           | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                                                                               |
| multiple targets                  | `target1 target2:`                              | :white_check_mark: | :white_check_mark: | :white_check_mark: | the builder places every target at the same position, [#120](https://github.com/unmango/go-make/issues/120)                   |
| pre-requisites                    | `target: prereq`                                | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| order-only pre-requisites         | `target: \| prereq`                             | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| order-only, unseparated           | `target: prereq\|order-only`                    | :white_check_mark: | :white_check_mark: |                    | only the first `\|` separates, a later one is text of the prerequisite holding it                                             |
| pattern rules                     | `%.o: %.c`                                      | :white_check_mark: | :white_check_mark: |                    | `%` is ordinary text, the pattern itself is not modeled                                                                       |
| suffix rules                      | `.c.o:`                                         | :white_check_mark: | :white_check_mark: |                    | ordinary target text                                                                                                          |
| special targets                   | `.PHONY: target`                                | :white_check_mark: | :white_check_mark: |                    | ordinary target text, see `ast/target` for the known names                                                                    |
| wildcard pre-requisites           | `target: *.c`                                   | :white_check_mark: | :white_check_mark: |                    | ordinary prerequisite text                                                                                                    |
| static pattern rules              | `a.o: %.o: %.c`                                 |                    |                    |                    | parse error on the second `:`                                                                                                 |
| double-colon rules                | `target:: prereq`                               |                    |                    |                    | parse error on the second `:`                                                                                                 |
| grouped targets                   | `a b &: prereq`                                 |     :warning:      |     :warning:      |                    | round-trips, but `&` is parsed as an additional target                                                                        |
| escaped spaces in targets         | `a\\ b:`                                        |     :warning:      |     :warning:      |                    | round-trips, but yields two targets rather than one                                                                           |
| target-specific variables         | `target: VAR = value`                           |                    |                    |                    | parse error on `=`                                                                                                            |
| pattern-specific variables        | `%.o: VAR = value`                              |                    |                    |                    | parse error on `=`                                                                                                            |
| **recipes**                       |                                                 |                    |                    |                    |                                                                                                                               |
| recipes                           | `target:\n\trecipe text`                        | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| multiple recipe lines             | `target:\n\tone\n\ttwo`                         | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| recipe prefix characters          | `\t@echo`, `\t-rm x`, `\t+$(MAKE)`              | :white_check_mark: | :white_check_mark: |                    | part of the recipe text                                                                                                       |
| escaped `$$` in recipes           | `\techo $$HOME`                                 | :white_check_mark: | :white_check_mark: |                    | recipe bodies are captured verbatim                                                                                           |
| automatic variables in recipes    | `\tcp $< $@`                                    |     :warning:      |     :warning:      |                    | verbatim text rather than `VarRef` nodes                                                                                      |
| variable references in recipes    | `\techo $(VAR)`                                 |     :warning:      |     :warning:      |                    | verbatim text rather than `VarRef` nodes                                                                                      |
| blank line inside a recipe block  | `target:\n\tone\n\n\ttwo`                       |     :warning:      |     :warning:      |                    | ends the rule, following lines become `BadObj` nodes, [#139](https://github.com/unmango/go-make/issues/139)                   |
| semicolon delimited recipes       | `target: ; recipe`                              | :white_check_mark: | :white_check_mark: |                    | a semicolon ends the prerequisite list wherever it is written, so `target: ;recipe` and `target: prereq;recipe` parse as well |
| custom `.RECIPEPREFIX`            | `.RECIPEPREFIX = >\n>recipe`                    | :white_check_mark: | :white_check_mark: |                    | the first character of the value introduces a recipe, an expanded value leaves the prefix unchanged                           |
| `.RECIPEPREFIX` of `;`            | `.RECIPEPREFIX = ;\n;recipe`                    | :white_check_mark: | :white_check_mark: |                    | recorded as a `TEXT` prefix like any other custom one, so it stays apart from the `;` of a `target: ; recipe`                 |
| **variables**                     |                                                 |                    |                    |                    |                                                                                                                               |
| assignment operators              | `=`, `:=`, `::=`, `:::=`, `?=`, `!=`            | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| append assignment                 | `VAR += x`                                      | :white_check_mark: | :white_check_mark: |                    | the unspaced `VAR+=x` parses as well, see the row below                                                                       |
| empty declarations                | `VAR :=`                                        | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| multi-word values                 | `VAR := foo.c bar.c`                            | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| unspaced `:`-prefixed operators   | `VAR:=x`, `VAR::=x`, `VAR:::=x`                 | :white_check_mark: | :white_check_mark: |                    | `:` delimits, so these scan correctly                                                                                         |
| unspaced `=`, `?=`, `!=`, `+=`    | `VAR=x`, `VAR?=x`, `VAR!=x`, `VAR+=x`           | :white_check_mark: | :white_check_mark: |                    | `=` and `!` delimit, so the name, the operator, and the value are separate tokens                                             |
| operators in a value              | `X = a = b`, `CFLAGS=-DFOO=1`                   | :white_check_mark: | :white_check_mark: |                    | make reads the first operator on a line, so a later `=`, `!`, or `:` is text of the value                                     |
| assignment with no name           | `=x`                                            |                    |                    |                    | parse error: `variable name is empty`, which is what make reports as well                                                     |
| definition blocks                 | `define FOO\nbody\nendef`                       |     :warning:      |     :warning:      |                    | no dedicated AST node, every line is stored as a `BadObj`                                                                     |
| `undefine`                        | `undefine VAR`                                  |     :warning:      |     :warning:      |                    | no dedicated AST node, stored as a `BadObj`                                                                                   |
| `override`                        | `override VAR := x`                             |     :warning:      |     :warning:      |                    | no dedicated AST node, stored as a `BadObj`                                                                                   |
| `export` and `unexport`           | `export VAR`, `export VAR := x`                 |     :warning:      |     :warning:      |                    | no dedicated AST node, stored as a `BadObj`                                                                                   |
| `private`                         | `private VAR := x`                              |     :warning:      |     :warning:      |                    | no dedicated AST node, stored as a `BadObj`                                                                                   |
| **variable references**           |                                                 |                    |                    |                    |                                                                                                                               |
| parentheses and braces            | `$(VAR)`, `${VAR}`                              | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                                                                               |
| single character                  | `$V`                                            | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| in targets                        | `${VAR}:`, `$(FOO) $(BAR):`                     | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                                                                               |
| in pre-requisites                 | `target: ${FOO}`                                | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| in variable values                | `A := $(B)`                                     | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| in conditions                     | `ifeq ($(V),x)`                                 | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| escaped `$$`                      | `target: $$V`                                   |     :warning:      |     :warning:      |                    | round-trips, but is stored as an `ast.Text` holding `$$` rather than a dedicated node                                         |
| adjacent to text                  | `target: prefix$(FOO)`, `a$(F)b`                | :white_check_mark: | :white_check_mark: |                    | `$` delimits, so the text and the reference are one `JuxtaposedExpr` rather than one `TEXT` token                             |
| a `$` that opens nothing          | `X := a$`, `X := $ b`                           | :white_check_mark: | :white_check_mark: |                    | a `$` with no expansion after it is an `ast.Text` holding the character                                                       |
| joined expressions                | `target: $(FOO)bar`, `$$(notdir x)`             | :white_check_mark: | :white_check_mark: |                    | expressions written with no blank between them are one `JuxtaposedExpr`                                                       |
| substitution references           | `$(VAR:.c=.o)`                                  |                    |                    |                    | parse error on `:`                                                                                                            |
| computed names                    | `$($(VAR))`                                     |                    |                    |                    | parse error on the inner `$`                                                                                                  |
| names matching a builtin          | `$(dir)`, `$(file)`, `$(word)`                  |     :warning:      |     :warning:      |                    | round-trips, but is a `FuncCall` with no arguments, which is how make reads it                                                |
| **functions**                     |                                                 |                    |                    |                    |                                                                                                                               |
| text and filename functions       | `$(subst a,b,$(S))`, `$(wildcard *.c)`          | :white_check_mark: | :white_check_mark: | :white_check_mark: | a `FuncCall`, arguments split on the commas make splits on                                                                    |
| `$(shell ...)`                    | `A := $(shell echo x)`                          | :white_check_mark: | :white_check_mark: | :white_check_mark: | the command is argument text, it is not parsed as shell syntax                                                                |
| `$(call ...)` and `$(eval ...)`   | `A := $(call f,x)`                              | :white_check_mark: | :white_check_mark: | :white_check_mark: | the called name is the first argument, not a `VarRef`                                                                         |
| control functions                 | `$(foreach v,$(L),$(v))`, `$(if $(V),x,y)`      | :white_check_mark: | :white_check_mark: | :white_check_mark: | a comma inside a nested call belongs to the argument holding it                                                               |
| logging functions                 | `$(info msg)`, `$(warning msg)`, `$(error msg)` | :white_check_mark: | :white_check_mark: | :white_check_mark: | `info` is absent from the manual's list, so it is a call to a name make does not know                                         |
| **conditionals**                  |                                                 |                    |                    |                    |                                                                                                                               |
| equality directives               | `ifeq`, `ifneq`                                 | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| definition directives             | `ifdef`, `ifndef`                               | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| parentheses syntax                | `ifeq (foo, bar)`                               | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| double quotes                     | `ifeq "foo" "bar"`                              | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| single quotes                     | `ifeq 'foo' 'bar'`                              | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| mixed quotes                      | `ifeq "foo" 'bar'`                              | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| empty argument                    | `ifeq ($(V),)`, `ifeq "" ""`                    | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| `else`                            | `ifdef V\na:\nelse\nb:\nendif`                  | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| `else if` chains                  | `else ifdef W`, `else ifeq (a, b)`              | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| nested conditionals               | `ifdef V\nifdef W\nendif\nendif`                | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| wrapping variables or rules       | `ifdef V\nA := x\nendif`                        | :white_check_mark: | :white_check_mark: |                    |                                                                                                                               |
| inside a recipe block             | `target:\nifdef V\n\techo x\nendif`             | :white_check_mark: | :white_check_mark: |                    | the conditional joins the recipe list of the rule it follows                                                                  |
| **other directives**              |                                                 |                    |                    |                    |                                                                                                                               |
| include directives                | `include a.mk`, `-include`, `sinclude`          |     :warning:      |     :warning:      |                    | no dedicated AST node, stored as a `BadObj`                                                                                   |
| `vpath` directive                 | `vpath %.c src`                                 |     :warning:      |     :warning:      |                    | no dedicated AST node, stored as a `BadObj`                                                                                   |
| `VPATH` variable                  | `VPATH := src`                                  | :white_check_mark: | :white_check_mark: |                    | an ordinary variable assignment                                                                                               |
| `load` directive                  | `load a.so`                                     |     :warning:      |     :warning:      |                    | not tokenized, stored as a `BadObj`                                                                                           |
| many other things                 |                                                 |                    |                    |                    | please open an issue if there is anything missing you'd like to see!                                                          |

Constructs marked unsupported either report a parse error or print differently than they were written.
Syntax the parser has no node for is stored verbatim as an `ast.BadObj` and printed back unchanged, so the line survives a round trip.

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
Targets will obtain dependencies automatically as needed.

Binaries are stored in a `.gitignore`d `bin/` directory at the root of the repository.
An example `.envrc` file for [direnv](https://github.com/direnv/direnv) is provided in [hack/example.envrc](./hack/example.envrc) to add `./bin` to your `PATH` automatically.
To use it, run `make .envrc` or `make dev`.
This will copy `hack/example.envrc` to `.envrc` at the root of the repository.

## References

GNU Make Quick Reference: <https://www.gnu.org/software/make/manual/html_node/Quick-Reference.html>
