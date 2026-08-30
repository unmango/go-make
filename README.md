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

|       Status       | Meaning                                                                                           |
| :----------------: | :------------------------------------------------------------------------------------------------ |
| :white_check_mark: | Modeled by a dedicated AST node and round-trips                                                   |
|     :warning:      | Round-trips, but is stored as plain text rather than a dedicated node, or is otherwise incomplete |
|      (blank)       | Not supported, see the linked issue for current behavior                                          |

| Syntax                                 | Example                                     |       Parser       |      Printer       |      Builder       | Remarks                                                                                                                                                                                                 |
| -------------------------------------- | ------------------------------------------- | :----------------: | :----------------: | :----------------: | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **general**                            |                                             |                    |                    |                    |                                                                                                                                                                                                         |
| newline escaping                       | `\trecipe text\\ncontinued on next line`    |     :warning:      |     :warning:      |                    | the backslash is plain text and each physical line becomes its own node                                                                                                                                 |
| newline separated elements             | `target:\n\ntarget2:`                       | :white_check_mark: | :white_check_mark: |                    | blank lines are recreated from stored positions                                                                                                                                                         |
| CRLF line endings                      | `target: prereq\r\n`                        |                    |                    |                    | text adjacent to `\r` is silently dropped by the scanner, [#113](https://github.com/unmango/go-make/issues/113)                                                                                         |
| no trailing newline                    | `target: prereq`                            |     :warning:      |     :warning:      |                    | the printer always terminates the final line with `\n`                                                                                                                                                  |
| trailing whitespace                    | `target: prereq  \n`                        |                    |                    |                    | trailing spaces are dropped, and a file ending in spaces scans as `UNSUPPORTED`                                                                                                                         |
| **comments**                           |                                             |                    |                    |                    |                                                                                                                                                                                                         |
| top-level comments                     | `# comment text`                            | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| comments with no leading space         | `#comment text`                             |     :warning:      |     :warning:      |                    | parses, but the printer always inserts a space after `#`, [#115](https://github.com/unmango/go-make/issues/115)                                                                                         |
| comment groups                         | `# comment text\n# more comment text`       | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| rule comments                          | `target: # comment text`                    |                    |                    |                    | parse error: `expected one of 'TEXT', '$', found 'COMMENT'`                                                                                                                                             |
| comments after a word                  | `prereq# comment text`                      |                    |                    |                    | `#` only starts a comment at the start of a token, [#112](https://github.com/unmango/go-make/issues/112)                                                                                                |
| recipe comments                        | `target:\n\trecipe # comment text\n`        | :white_check_mark: | :white_check_mark: |                    | these are not make comments and are included in the recipe text                                                                                                                                         |
| **rules**                              |                                             |                    |                    |                    |                                                                                                                                                                                                         |
| targets                                | `target:`, `target :`                       | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                                                                                                                                                         |
| multiple targets                       | `target1 target2:`                          | :white_check_mark: | :white_check_mark: | :white_check_mark: | the builder places every target at the same position, [#120](https://github.com/unmango/go-make/issues/120)                                                                                             |
| pre-requisites                         | `target: prereq`                            | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| order-only pre-requisites              | `target: \| prereq`                         | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| order-only pre-requisites, unseparated | `target: prereq\|order-only`                |                    |                    |                    | `\|` only delimits at the start of a token, [#112](https://github.com/unmango/go-make/issues/112)                                                                                                       |
| recipes                                | `\trecipe text\n`                           | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| recipe with a custom `.RECIPEPREFIX`   | `\|recipe text\n`                           |                    |                    |                    | the prefix is fixed to `TAB`, [#127](https://github.com/unmango/go-make/issues/127)                                                                                                                     |
| semicolon delimited recipes            | `target: ;recipe text\n`                    |                    |                    |                    | scanned as pre-requisite text, `target: ; recipe` is a parse error, [#127](https://github.com/unmango/go-make/issues/127)                                                                               |
| pattern rules                          | `%.o: %.c`                                  |     :warning:      |     :warning:      |                    | round-trips as ordinary target and pre-requisite text, the pattern is not modeled                                                                                                                       |
| static pattern rules                   | `a.o: %.o: %.c`                             |                    |                    |                    | parse error on the second `:`                                                                                                                                                                           |
| double-colon rules                     | `target:: prereq`                           |                    |                    |                    | parse error on the second `:`                                                                                                                                                                           |
| **variables**                          |                                             |                    |                    |                    |                                                                                                                                                                                                         |
| empty declarations                     | `VAR :=`                                    | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| simple declarations                    | `VAR := foo.c bar.c`                        | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| all assigment operators                | `VAR != foo`, `VAR ::= bar`, etc.           | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| unspaced `:`-prefixed assignments      | `VAR:=foo`, `VAR::=foo`                     | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| unspaced `=`, `?=`, and `!=`           | `VAR=foo`, `VAR?=foo`, `VAR!=foo`           |                    |                    |                    | scanned as a single `TEXT` token, parses to a `nil` entry in `File.Contents` with no error, [#112](https://github.com/unmango/go-make/issues/112) [#111](https://github.com/unmango/go-make/issues/111) |
| **variable references**                |                                             |                    |                    |                    |                                                                                                                                                                                                         |
| in targets                             | `${VAR}:`, `$(FOO) $(BAR):`                 | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                                                                                                                                                         |
| in prereqs                             | `target: ${FOO}`                            | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| in variable values                     | `A := $(B)`                                 | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| single character                       | `target: $F`                                | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| in recipes                             | `target:\n\trecipe $(VAR)\n`                |     :warning:      |     :warning:      |                    | round-trips, but is part of the flat recipe text rather than a `VarRef`                                                                                                                                 |
| adjacent to text                       | `target: prefix$(FOO)`                      |                    |                    |                    | `$` only starts a reference at the start of a token, [#112](https://github.com/unmango/go-make/issues/112)                                                                                              |
| named after a builtin function         | `$(dir)`, `$(file)`, `$(word)`              |                    |                    |                    | builtin names scan as keywords, so the parser reports `expected 'TEXT'`                                                                                                                                 |
| **directives**                         |                                             |                    |                    |                    |                                                                                                                                                                                                         |
| conditional directives                 | `ifeq`, `ifneq`, `ifdef`, `ifndef`          | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| equality directives                    | `ifeq`, `ifneq`                             | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| parentheses syntax                     | `ifeq (foo, bar)`                           | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| double quotes                          | `ifeq "foo" "bar"`                          | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| single quotes                          | `ifeq 'foo' 'bar'`                          | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| mixed syntax                           | `ifeq "foo" 'bar'`                          | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| definition directives                  | `ifdef`, `ifndef`                           | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| `else` and `else if` chains            | `ifeq (a, b)\nelse ifdef C\nendif`          | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| nested conditionals                    | `ifeq (a, b)\nifdef C\nendif\nendif`        | :white_check_mark: | :white_check_mark: |                    |                                                                                                                                                                                                         |
| definition blocks                      | `define FOO\nbar\nendef`                    |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                                                      |
| `undefine`                             | `undefine FOO`                              |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                                                      |
| include directives                     | `include foo.mk`, `-include`, `sinclude`    |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                                                      |
| scope directives                       | `export`, `unexport`, `override`, `private` |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                                                      |
| `vpath`                                | `vpath %.c src`                             |                    |                    |                    | no AST node, [#111](https://github.com/unmango/go-make/issues/111)                                                                                                                                      |
| logging directives                     | `$(info message)`                           |                    |                    |                    | parse error                                                                                                                                                                                             |
| expressions                            | `$(shell script stuff)`                     |                    |                    |                    | builtin functions are scanned as keywords and rejected by the parser                                                                                                                                    |
| many other things                      |                                             |                    |                    |                    | please open an issue if there is anything missing you'd like to see!                                                                                                                                    |

Every entry marked as unsupported above that reports no parse error appends a `nil` element to `ast.File.Contents`.
Printing such a file panics, tracked in [#111](https://github.com/unmango/go-make/issues/111).

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

Automatic variables (`$@`, `$<`, `$(@D)`, and the rest) are summarized by the manual but are not enumerated here, so they are excluded from the comparison.

### Will Not Support

Nothing, at this time

## Workflow

### Pre-Requisites

Go toolchain for the version listed in [go.mod](./go.mod)

### Building

go-make is itself built using `make`.

|         Targets | Description                                                        |
| --------------: | :----------------------------------------------------------------- |
|    default goal | Runs the `build` target                                            |
|         `build` | Runs `go build` to verify the code compiles                        |
|          `test` | Test changed packages                                              |
|      `test_all` | Test all packages                                                  |
| `sync-quickref` | Refresh the GNU Make manual fixture used by `internal/conformance` |
|         `clean` | Remove `.make` directory and coverage report                       |
|         `cover` | Collect coverage for all tests and print report                    |
|          `tidy` | Runs `go mod tidy`                                                 |
|           `dev` | Setup the [developer environment](#developer-environment)          |

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
