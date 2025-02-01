# Go Make

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

### Supported Features

Makefile syntax that is guaranteed to round-trip (parse and print without modification) is listed in [./testdata/roundtrip](./testdata/roundtrip/).
Additional syntax is supported and may round-trip successfully, but no guarentees are provided until it is listed under `./testdata/roundtrip`.

- comments
  - [ ] top-level comments i.e. `# comment text`
  - [ ] rule comments i.e. `target: # comment text`
  - [x] recipe comments i.e. `target:\n\trecipe # comment text\n`
    - these are not make comments and are included in the recipe text
- rules
  - [x] targets i.e. `target:`, `target :`
  - [x] multiple targets i.e. `target1 target2:`
  - [x] pre-requisites i.e. `target: prereq`
  - [x] order-only pre-requisites i.e. `target: | prereq`
  - [x] recipes i.e. `\trecipe text\n`
  - [ ] recipe with a custom `.RECIPEPREFIX`
  - [ ] semimcolon delimited recipes i.e. `target: ;recipe text\n`
- variables
  - [x] empty declarations i.e. `VARIABLE :=`
  - [x] simple declarations i.e. `VARIALBE := foo.c bar.c`
  - [x] all assigment operators i.e. `VARIABLE != foo`, `VARIABLE ::= bar`, etc.
  - variable references i.e. `${VARIABLE}`
    - [x] in targets i.e. `${VARIABLE}:`, `$(FOO) $(BAR):`
    - [ ] in prereqs i.e. `target: ${FOO}`
    - [ ] in recipes i.e. `target:\n\trecipe $(VAR)\n`
- directives
  - [ ] top-level directives i.e. `ifeq`, `define`, etc.
  - [ ] logging directives i.e. `$(info message)`
  - [ ] expressions i.e. `$(shell script stuff)`

#### Will Not Support

Nothing, at this time
