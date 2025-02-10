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

| Syntax                               | Example                                  |       Parser       |      Printer       |      Builder       | Remarks                                                              |
| ------------------------------------ | ---------------------------------------- | :----------------: | :----------------: | :----------------: | -------------------------------------------------------------------- |
| newline escaping                     | `\trecipe text\\ncontinued on next line` |                    |                    |                    |                                                                      |
| newline separated elements           | `target:\n\ntarget2:`                    |                    |                    |                    |                                                                      |
| **comments**                         |                                          |                    |                    |                    |                                                                      |
| top-level comments                   | `# comment text`                         | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| comment groups                       | `# comment text\n# more comment text`    | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| rule comments                        | `target: # comment text`                 |                    |                    |                    |                                                                      |
| recipe comments                      | `target:\n\trecipe # comment text\n`     | :white_check_mark: | :white_check_mark: |                    | these are not make comments and are included in the recipe text      |
| **rules**                            |                                          |                    |                    |                    |                                                                      |
| targets                              | `target:`, `target :`                    | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                      |
| multiple targets                     | `target1 target2:`                       | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                      |
| pre-requisites                       | `target: prereq`                         | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| order-only pre-requisites            | `target: \| prereq`                      | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| recipes                              | `\trecipe text\n`                        | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| recipe with a custom `.RECIPEPREFIX` | `\|recipe text\n`                        |                    |                    |                    |                                                                      |
| semimcolon delimited recipes         | `target: ;recipe text\n`                 |                    |                    |                    |                                                                      |
| **variables**                        |                                          |                    |                    |                    |                                                                      |
| empty declarations                   | `VAR :=`                                 | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| simple declarations                  | `VAR := foo.c bar.c`                     | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| all assigment operators              | `VAR != foo`, `VAR ::= bar`, etc.        | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| **variable references**              |                                          |                    |                    |                    |                                                                      |
| in targets                           | `${VAR}:`, `$(FOO) $(BAR):`              | :white_check_mark: | :white_check_mark: | :white_check_mark: |                                                                      |
| in prereqs                           | `target: ${FOO}`                         | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| in recipes                           | `target:\n\trecipe $(VAR)\n`             |                    |                    |                    |                                                                      |
| **directives**                       |                                          |                    |                    |                    |                                                                      |
| top-level directives                 | `ifeq`, `define`, etc.                   |                    |                    |                    |                                                                      |
| conditional directives               | `ifeq`, `ifneq`, `ifdef`, `ifndef`       | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| equality directives                  | `ifeq`, `ifneq`                          | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| parentheses syntax                   | `ifeq (foo, bar)`                        | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| double quotes                        | `ifeq "foo" "bar"`                       | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| single quotes                        | `ifeq 'foo' 'bar'`                       | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| mixed syntax                         | `ifeq "foo" 'bar'`                       | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| definition directives                | `ifdef`, `ifndef`                        | :white_check_mark: | :white_check_mark: |                    |                                                                      |
| logging directives                   | `$(info message)`                        |                    |                    |                    |                                                                      |
| expressions                          | `$(shell script stuff)`                  |                    |                    |                    |                                                                      |
| many other things                    |                                          |                    |                    |                    | please open an issue if there is anything missing you'd like to see! |

### Will Not Support

Nothing, at this time

## Workflow

### Pre-Requisites

Go toolchain for the version listed in [go.mod](./go.mod)

### Building

go-make is itself built using `make`.

|      Targets | Description                                               |
| -----------: | :-------------------------------------------------------- |
| default goal | Runs the `build` target                                   |
|      `build` | Runs `go build` to verify the code compiles               |
|       `test` | Test changed packages                                     |
|   `test_all` | Test all packages                                         |
|      `clean` | Remove `.make` directory and coverage report              |
|      `cover` | Collect coverage for all tests and print report           |
|       `tidy` | Runs `go mod tidy`                                        |
|        `dev` | Setup the [developer environment](#developer-environment) |

### Developer Environment

Apart from the Go toolchain, the only main dependency is the `ginkgo` cli to run tests.
This repo also uses [devctl](https://github.com/unmango/devctl) but its use is optional.
Targets will obtain dependencies automatically as needed.

Binaries are stored in a `.gitignore`d `bin/` directory at the root of the repository.
An example `.envrc` file for [direnv](https://github.com/direnv/direnv) is provided in [hack/example.envrc](./hack/example.envrc) to add `./bin` to your `PATH` automatically.
To use it, run `make .envrc` or `make dev`.
This will copy `hack/example.envrc` to `.envrc` at the root of the repository.
