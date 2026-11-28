# libaql - Go realization hints

Companion to libaql.spec.md for a Go translation. Two library hints apply:
`library.go.lemongrass.hints.md` for the parser generator and
`library.go.re2c.hints.md` for the lexer generator. This file has only what
is specific to libaql.

Language: Go
Module:   github.com/mge1512/aql/go/aql
Package:  aql

Every statement about a tool in this file was run with lemongrass 0.1.1,
re2c 4.6 and Go 1.24.7, on the specification at Version 0.7.0.

## 1. Identity, place and version

- The module path is `github.com/mge1512/aql/go/aql`, and the package has
  the name `aql`. The component keeps the name libaql in the specification.
- The module is in the directory `go/aql/` of the repository that has the
  specification. That directory is the root of the module and has the
  complete result of the translation.
- `go.mod` has the line `go 1.24`. The library needs nothing newer.
- The library uses the standard library only. `go.mod` has no `require`
  line. There is no `go.sum`, no `vendor` directory and no vendor tarball.
- The version of the binding is the Version of the specification. The file
  `VERSION` has that Version. The maintainer sets the release tag, which is
  `go/aql/v` with the Version.

## 2. Files beside the Go sources

These files are part of the module, in addition to the files that the
library template names:

| File | Content | Comes from |
|---|---|---|
| `tokens.re` | The definitions of the token table | The definitions block, copied byte for byte |
| `lexer.re` | The two tokenizers, as input for re2c | The translator |
| `lexer.go` | The two tokenizers in Go | re2c, from `lexer.re` and `tokens.re` |
| `grammar.y` | The grammar with its actions, as input for lemongrass | The translator |
| `grammar.go` | The parser in Go | lemongrass, from `grammar.y` |
| `check-spec.sh` | The check of section 8 | Section 8, copied as it stands |

- `lexer.go` and `grammar.go` are committed with the other sources. Nobody
  edits them, and nobody runs `gofmt` on them.
- The head comment with the hash of the specification stands at the top of
  `lexer.re`, and in the `%include` block of `grammar.y` before the package
  clause. The two tools copy it into the generated files.
- The translator writes one Go file for each BEHAVIOR of the
  specification. One more file has the types of the public interface.

## 3. Makefile

The Makefile has these lines. It also has what the library template asks
for beyond them, for example the target `dist`. The recipe lines start with
a tab.

```make
# Tools of the build and their pinned versions.
LEMONGRASS          ?= lemongrass
LEMONGRASS_TEMPLATE ?= /usr/share/lemongrass/lempar.go
LEMONGRASS_VERSION   = 0.1.1
RE2C                ?= re2c
RE2C_VERSION         = 4.6
SPEC                ?= ../../specification/libaql.spec.md
GO                  ?= go

LEMONGRASS_FLAGS = -q -l -T$(LEMONGRASS_TEMPLATE)
RE2C_FLAGS       = --lang go -8 -i --no-generation-date

.PHONY: all generate check-tools check-generated check-spec build test clean

all: build

# The tools must have the pinned versions.
check-tools:
	@test "$$($(LEMONGRASS) -x | sed -n 's/^lemongrass version //p')" = "$(LEMONGRASS_VERSION)" || \
	  { echo "check-tools: lemongrass $(LEMONGRASS_VERSION) is needed" >&2; exit 1; }
	@test "$$($(RE2C) --version)" = "re2c $(RE2C_VERSION)" || \
	  { echo "check-tools: re2c $(RE2C_VERSION) is needed" >&2; exit 1; }

# Writes grammar.go and lexer.go again. Run it after a change of grammar.y,
# lexer.re or tokens.re, and commit the result.
generate: check-tools
	$(LEMONGRASS) $(LEMONGRASS_FLAGS) grammar.y
	$(RE2C) $(RE2C_FLAGS) lexer.re -o lexer.go
	chmod 644 lexer.go

# The generated files in this directory must be what the tools write now.
check-generated: check-tools
	@tmp=$$(mktemp -d) && \
	$(LEMONGRASS) $(LEMONGRASS_FLAGS) -d$$tmp grammar.y && \
	$(RE2C) $(RE2C_FLAGS) lexer.re -o $$tmp/lexer.go && \
	cmp grammar.go $$tmp/grammar.go && cmp lexer.go $$tmp/lexer.go; \
	status=$$?; rm -rf $$tmp; exit $$status

# The token table and the grammar must be those of the specification.
check-spec:
	@LEMONGRASS="$(LEMONGRASS)" sh check-spec.sh $(SPEC)

build: check-generated check-spec
	$(GO) build ./...

test: build
	$(GO) test ./...

clean:
	$(GO) clean ./...
```

- `build` depends on `check-generated` and `check-spec`. A package build
  therefore fails when a generated file is old or a tool has another
  version. It also fails when the token table or the grammar differ from
  the specification.
- `generate` is for the translator and the maintainer. A package build
  never calls it.
- re2c writes its output file with the mode 0600. `generate` therefore
  sets the mode 0644, so that a tarball of the working tree has a file
  that every user can read.
- The command `make install PREFIX=/usr` of lemongrass installs the Go
  template as `/usr/share/lemongrass/lempar.go`. A package with another
  prefix has another path. `LEMONGRASS_TEMPLATE` is the one place that
  names the path.
- The release tarball of `dist` keeps the layout of the repository:
  `specification/libaql.spec.md` and `go/aql/`. The value of `SPEC` then
  works in a package build without a change.

## 4. tokenize: the tokenizer comes from re2c

- `tokens.re` is the definitions block of the token table. Copy it byte for
  byte, with its comment lines.
- `lexer.re` has two functions, in this order: one for the strict reading
  and one for the lenient reading. The first block includes `tokens.re`.
- Each block has one rule for each row of the rows list, in the order of
  the list. The name of the rule is the name of the row. In the second
  block the rows text_1 and open_1 have the names `text_1_lenient` and
  `open_1_lenient`.
- Write each of these rules on one line: the name, spaces, and the code in
  braces. `check-spec.sh` reads this form:

      text_1           { return rowTEXT, cur }

- After the rows each block has three more rules: `$` for the end of the
  text, `any` for a character that starts no token, and `*`.
- A rule returns the kind of its row and the place after the run. The Go
  code around the tokenizers does not decide what a token is.
- No row fits when the rule `any` or the rule `*` wins.
- A keyword is found with the letters A to Z only. Do not use
  `strings.ToUpper` or `strings.EqualFold`, because they change other
  letters too. With the letter U+017F in place of its `s`, the word
  `select` would still be the keyword SELECT. Change the letters `a` to `z`
  in code of your own.
- Positions and lengths count characters. Add
  `utf8.RuneCountInString(run)` to the position after each run, and use
  `utf8.RuneCountInString(text)` for the length of the statement.
- The content of a TEXT comes from its run by the rules for tokens. Code
  that reads a run to find its value is allowed.
- A NUMBER keeps every digit. Keep its run as text. Do not use `float64`.
- A Go string can have bytes that are not valid UTF-8. Check 1a of
  parse-statement is `utf8.ValidString(text)`, and it comes before every
  other use of the text. For the position in the message S21, walk the
  text with `utf8.DecodeRuneInString`. The first result `utf8.RuneError`
  with the width 1 is the place that is not valid. The position is the
  number of the characters before it, plus 1.
- The two checks of step 1 keep no copy of the text. Use
  `utf8.ValidString` and `utf8.RuneCountInString` for them, and make no
  `[]rune` of the text before both have passed.
- The tokenizers therefore never get a byte that is not valid UTF-8. The
  rule `*` stays in `lexer.re`. Without it re2c warns that the tokenizer
  has no rule for such bytes.

## 5. parse-statement: the parser comes from lemongrass

- `grammar.y` has the lines of the grammar block in their order. Write
  each rule on one line up to its period. Add labels and actions, and
  change nothing else. `check-spec.sh` reads this form.
- Declare the five keywords that no rule uses, so that they have a token
  number:

      %token LEFT RIGHT FULL OUTER OFFSET.

- Use `%stack_size 0`. tokenize refuses a statement with more than
  MaxTokens tokens before the parser runs, so the stack cannot grow beyond
  about MaxTokens entries. The example statement_at_the_token_limit needs
  1000 entries. With the default of 100 the parser fails for correct
  statements.
- Use the index of a token in the token list as the value of a token:
  `%token_type {int}`. An action keeps such indices in the node that it
  builds, and it reads no token. The code that builds the Statement reads
  the tokens after the parse. The value of `%extra_argument` has the
  result of the parse and no token list.
- A trial of classify-failure is safe with this form. Its last token has
  an index outside the token list, and no action reads it.
- In Go, `allowed_kinds` is a `[]StatementKind`. The functions `ReadSet()`
  and `WriteSet()` return the two sets. Steps 5c and 6 tell the read set
  by its members select and describe.
- The kind END is the token number 0.
- Steps 4 to 7 look at the tokens before the parser runs.
- Stop at the first token that the parser refuses. The parser can report
  an accepted statement after a refused token.
- One parser from lemongrass fits into one package. The package `aql` has
  this one.

## 6. classify-failure: which kinds can come next

- The failing token O is the token at which the parser ran the code of
  `%syntax_error`. It is END when that happened in the call with the token
  number 0.
- "X can come next" is a trial with a new parser. The function
  `canComeNext` in `library.go.lemongrass.hints.md` shows it. The parser
  of the trial gets the tokens before O and then one token of the kind in
  question.
- No table of the parser is read by hand, and no other code decides what
  the grammar allows.
- The kinds LEFT, RIGHT, FULL, OUTER and OFFSET are in no rule of the
  grammar. They have token numbers only because the grammar file declares
  them with a `%token` line.

## 7. syntax-messages: characters, not bytes

- MaxEcho and ExcerptLength count characters. `s[:n]` cuts bytes and can
  cut a character in two. Cut with `[]rune(s)[:n]` or with a loop over the
  characters.
- `<where>` and every position in a message count characters from 1.
- Capital letters in a message are made like the keywords of section 4:
  only the letters `a` to `z` change.

## 8. The check against the specification

`check-spec.sh` is the check that the specification asks for under
`spec-parity-check`. Write the file as it stands here:

```sh
#!/bin/sh
# Compares the token table and the grammar in this directory with the
# specification.
#
# Usage: sh check-spec.sh SPECIFICATION
#
# The script reads tokens.re, lexer.re and grammar.y. It needs the program
# lemongrass (variable LEMONGRASS). It ends with status 1 when a part
# differs.

set -u
[ $# -eq 1 ] && [ -r "$1" ] || { echo "usage: sh check-spec.sh SPECIFICATION" >&2; exit 2; }
spec=$1
lemongrass=${LEMONGRASS:-lemongrass}
tmp=$(mktemp -d) || exit 2
trap 'rm -rf "$tmp"' EXIT
failed=0
differ() { echo "check-spec: $1" >&2; failed=1; }

# block START: the first fenced block after the line that starts with START
block() {
  awk -v start="$1" 'BEGIN { mark = sprintf("%c%c%c", 96, 96, 96) }
    index($0, start) == 1 { found = 1 }
    found && index($0, mark) == 1 { fence++; next }
    found && fence == 1 { print }
    found && fence == 2 { exit }' "$spec"
}

# 1. The definitions of the token table: the same text.
block "Definitions. They are written" >"$tmp/definitions"
[ -s "$tmp/definitions" ] || differ "no definitions block in $spec"
cmp -s "$tmp/definitions" tokens.re ||
  differ "tokens.re is not the definitions block of the specification"

# 2. The rows: each tokenizer has one rule for each row, in the order of
#    the table. The second tokenizer has the lenient patterns.
block "Rows. Each line has" | awk '{ print $1 }' >"$tmp/rows"
[ -s "$tmp/rows" ] || differ "no rows block in $spec"
sed -e 's/^text_1$/text_1_lenient/' -e 's/^open_1$/open_1_lenient/' \
  "$tmp/rows" >"$tmp/rows-lenient"
# rules N: the names of the rules in the re2c block number N of lexer.re.
# The rule for any character that comes after the rows is left out.
rules() {
  awk -v want="$1" '/\/\*!re2c/ { n++; inside = 1; next }
    inside && /\*\// { inside = 0 }
    inside && n == want && $1 ~ /^[a-z_][a-z0-9_]*$/ && $2 == "{" && $1 != "any" { print $1 }' lexer.re
}
rules 1 >"$tmp/rules-strict"
rules 2 >"$tmp/rules-lenient"
cmp -s "$tmp/rows" "$tmp/rules-strict" ||
  differ "the rules of the first tokenizer in lexer.re are not the rows of the specification"
cmp -s "$tmp/rows-lenient" "$tmp/rules-lenient" ||
  differ "the rules of the second tokenizer in lexer.re are not the rows with the lenient patterns"

# 3. The grammar.
block "Grammar. It is written" | grep -v '^$' >"$tmp/grammar"
[ -s "$tmp/grammar" ] || differ "no grammar block in $spec"
# 3a. The rules as the generator reads them. It prints the rules that have
#     an action first, so the two lists are sorted before the comparison.
"$lemongrass" -g grammar.y >"$tmp/reprint" 2>"$tmp/reprint.err" ||
  differ "lemongrass -g grammar.y failed: $(head -1 "$tmp/reprint.err")"
grep '::=' "$tmp/reprint" | sort >"$tmp/rules-read"
sort "$tmp/grammar" >"$tmp/grammar-sorted"
cmp -s "$tmp/grammar-sorted" "$tmp/rules-read" ||
  differ "the rules that lemongrass reads from grammar.y are not the rules of the specification"
# 3b. The order of the rules in grammar.y. A rule starts at the beginning
#     of a line and ends at the first period. Labels are left out.
awk '/^[a-z_][a-z0-9_]*(\([A-Za-z0-9_]+\))? ::=/ {
    rule = substr($0, 1, index($0, "."))
    gsub(/\([A-Za-z0-9_]+\)/, "", rule)
    print rule }' grammar.y >"$tmp/rules-order"
cmp -s "$tmp/grammar" "$tmp/rules-order" ||
  differ "the rules of grammar.y are not in the order of the specification"

[ $failed -eq 0 ] && echo "check-spec: the token table and the grammar are those of the specification"
exit $failed
```

The script was run against planted differences: two rules in another
order, one changed rule, one changed definition, and two rows in another
order. It found each of them.

## 9. Tests

- The EXAMPLES are the tests. Some of them describe a long statement by a
  part that repeats. Build these with `strings.Repeat`.
- The example text_that_is_not_valid needs a byte that is not valid UTF-8.
  Write it as `\xff` in a Go string.
- `make test` runs the checks of section 3 first and then `go test ./...`.

## 10. Packaging

- The package build needs Go, lemongrass 0.1.1 and re2c 4.6. The versions
  are exact, because the build compares the generated files.
- The build needs no network and downloads nothing.
- The package build calls `make build` and `make test` in `go/aql/`.
- The directory `debian/` is in `go/aql/`, with the rest of the result.
  Its `rules` file works from the top of the `dist` tarball, so its paths
  start with `go/aql/`.

## 11. What not to build

- No scanner and no parser written by hand, and no use of the `regexp`
  package to find tokens.
- No module from outside the standard library.
- No `go:generate` line. The Makefile is the one place that calls the
  tools.
- No change by hand in `tokens.re`, `lexer.go` or `grammar.go`.
- No file for source control or for an editor, such as `.gitignore`.
