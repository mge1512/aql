# AQL architecture

This document records the design decisions for AQL and libaql as of 2026-11-27.
The specification of libaql follows it. Where the two differ, the specification
is right and this document needs an update.

## 1. Purpose

AQL (API Query Language) is a small SQL-like language for data behind an API.
An MCP server offers it as its interface, with two tools:

- `query` reads: `SELECT` and `DESCRIBE`
- `execute` changes: `INSERT`, `UPDATE` and `DELETE`

Two statements as an example, for a backend with tickets:

```
SELECT id, name, status FROM tickets WHERE status = 'open' ORDER BY name LIMIT 20
UPDATE tickets SET status = 'closed' WHERE id = 'T-17'
```

Reasons for this form:

- The tool definitions stay small and do not grow with the number of entities.
  A model finds entities and fields with `DESCRIBE`.
- Models write SQL without extra instructions. This includes small local
  models.
- The server filters, selects, groups and counts, so results stay small.
- Reading and changing are separate tools. A host can allow the first and ask
  before the second.

## 2. Names

| Name | Meaning |
|---|---|
| AQL | the language |
| libaql | the library with the engine and the MCP shell |
| aql | this repository. In code it is also the name of the library: the Go package and the Rust crate. |
| `mcp-server-<backend>-aql` | a server for one backend, for example `mcp-server-aha-aql` |

Other products use the acronym AQL for other languages. Text that a model
reads therefore names AQL together with "SQL-like" and a `SELECT` example.

## 3. Layers

```
MCP host
   |  tools: query, execute
shell     libaql        tools, modes, configuration, log line for each change
engine    libaql        tokens, grammar, rules, evaluation, writes, messages, limits
   |  adapter contract
adapter   per backend   plan, fetch, convert, write
catalog   per backend   entities, fields, references, scope rules
   |
backend API
```

There are three kinds of content, and each is produced in its own way:

| Content | Written in | Becomes code by |
|---|---|---|
| Engine and shell | the specification of libaql | translation, once per target language |
| Catalog | tables in the specification of a server | a tool that reads the tables (planned) |
| Adapter | behaviors in the specification of a server | translation, per backend |

The engine contains no MCP code, so it can be tested alone.

## 4. Decisions

### 4.1 An engine of our own, as a library

libaql has its own engine and no embedded SQL database. It is small, it brings
no database engine with it, and we decide the wording of every message. As a
library it is translated and reviewed once for each target language, and every
server in that language uses the same code.

### 4.2 One language for reading and changing

`SELECT`, `DESCRIBE`, `INSERT`, `UPDATE` and `DELETE` belong to one language
with one front end. The two tools accept different statement kinds.

### 4.3 Scope of the language, version 1

The starting point is the language of aha-aql 0.24. Version 1 adds:

- `JOIN` of two entities along a reference that the catalog declares
- `field IN (SELECT ...)`, not correlated, one level deep. `NOT` negates it,
  as for an `IN` list.
- `HAVING`
- `field MATCH 'text'`, a search condition that the backend answers. It is
  allowed only on a field that the catalog marks as a search field. This is
  the form of full text search in SQLite.

Everything else stays out of version 1. Examples are a statement for actions
that are not records, arithmetic in the select list, and a join of more than
two entities.

AQL uses only keywords that SQL also has, in the standard or in a widely used
dialect. A new construct takes the SQL word and, where one exists, the SQL
form.

A rule of aha-aql that turns out to fit one backend only moves to the adapter.

### 4.4 A declarative front end

The specification has two formal blocks:

- a token table in the notation of the lexer generator re2c. It has a block of
  definitions with one pattern for each row, and the list of the rows. The
  longest match wins. The first row wins a tie.
- a grammar in the notation of the Lemon parser generator, without actions.

A binding generates the parser from the grammar and the tokenizer from the
token table. No scanning or parsing logic is written by hand. A check compares
the patterns and rules in the code with the two blocks.

Every pattern gives the same match with a tool that looks for the longest
match and with a tool that takes the first alternative.

### 4.5 A grammar without conflicts

The grammar is LALR(1) and has no conflict. The generator checks this in every
build. A build with a reported conflict fails.

### 4.6 Common mistakes

- Text with an apostrophe inside a word gets a second attempt with a lenient
  text pattern.
- The grammar has the valid statements only. When it cannot read a
  statement, a table chooses the message. The table looks at the first token
  that cannot be read, at the two tokens before it, and at the token kinds
  that the grammar allows at that place.
- The grammar reads a few forms that AQL does not have, and a check after the
  parser refuses them. Each of these forms gets its own message in this way.
- Messages about names, types and labels come from the code after the parser.
- A message says what to write instead and shows a statement that works.

### 4.7 Evaluation without a planner

The engine fetches each entity of a statement once through the adapter, inside
a budget. It joins in memory and evaluates every condition itself. This is
enough because a budget of pages limits every list.

### 4.8 The adapter contract

An adapter answers a small set of calls. It describes the catalog, plans the
requests for an entity, fetches and converts records, and writes one record.
These rules apply to every adapter:

1. A plan returns a superset: every record that matches the conditions it
   claims. The filter of the backend saves time. The result comes from the
   evaluation in the engine. The exception is a `MATCH` condition, which the
   backend alone answers.
2. A statement is answered completely or refused with a scope message.
3. The engine counts requests and pages against a budget. An adapter cannot
   exceed it.
4. An adapter receives typed values and never statement text. It builds
   requests through encoders.
5. One write statement changes one record, found by id. The engine checks the
   guards on a record that it read immediately before.
6. No state is kept between statements, except caches that the specification
   names.
7. Record content is data. It leaves the server as JSON text and is never
   interpreted.

### 4.9 Limits

The engine has a limit for each of these:

- the length of a statement
- the number of tokens
- the nesting depth
- the length of a `LIKE` pattern
- the length of text that a message repeats

The limit for tokens also limits an `IN` list. `LIKE` is evaluated with
constant extra memory. A statement above one of the first four limits is
refused with a message. Text that a message repeats is cut at its limit.

### 4.10 Bindings

A binding is the library in one target language. The specification is written
for more than one. The first bindings are Go and Rust. Each binding has its
own hints, and the specification names none of its tools.

Rules for every binding:

- The parser comes from lemongrass (see 4.11) and the tokenizer from re2c (see
  4.12). All bindings use the same pinned version of each tool.
- A build with a reported conflict in the grammar fails.
- The generated files are part of the source. The package build generates them
  again and compares.
- At run time a binding uses the standard library of its language and no
  other library.

Place and names:

- The bindings are in this repository, beside the specification. The Go module
  is in `go/aql/` and the Rust crate is in `rust/aql/`. Each of the two
  directories has the complete result of one translation.
- In code the library has the name `aql`. This is the name of the Go package
  and of the Rust crate. libaql stays the name of the component in the
  specification.
- The import path of the Go package is `github.com/mge1512/aql/go/aql`. Go
  needs the directory of a module at the start of its release tags, so a tag
  of the Go module starts with `go/aql/`.
- No build takes the Rust crate from crates.io by its name alone. A server
  names the source of the crate and builds from vendored sources. From its
  first release on, the crate is also published on crates.io, so that nobody
  else can take the name.

Versions:

- A binding has the Version of the specification that it was translated from.
  The rule is the same for every target language.
- A binding has at most one release for each Version.
- A new translation without a change of the specification needs a new PATCH
  Version of the specification. Examples are new hints and a new version of a
  tool. The changelog entry gives the reason.
- A release tag has the directory of the binding, the letter `v` and the
  Version: `go/aql/v0.6.0` for the Go module and `rust/aql/v0.6.0` for the
  Rust crate. Go needs this form, and the Rust crate takes the same one.

### 4.11 One parser generator for all bindings

The parser generator is [lemongrass](https://github.com/mge1512/lemongrass).
It is the Lemon parser generator of SQLite with one addition: a template can
describe the syntax of its target language. lemongrass has templates for C, Go
and Rust. The bindings pin its release 0.1.1. An earlier plan had
goyacc for the Go binding and no generator for the Rust binding.

Reasons:

- All bindings use one generator and one notation. The grammar block of the
  specification is written in this notation, and lemongrass reads the block as
  it is.
- lemongrass prints the rules of a grammar file without the labels and the
  actions. The rules of a binding can therefore be compared with the block as
  text.
- The parser that it writes needs no library. The Go parser imports nothing.
  The Rust parser uses the standard library and has no `unsafe` code.
- lemongrass is in the public domain, like Lemon.
- The generator is one C file and needs the C library only. A build service
  can build it from source and offer it as a signed package.
- It ends with an error status when the grammar has a conflict. It writes the
  parser file in this case too, so the build checks the status.

What follows for a binding:

- Each binding has its own grammar file. Its rules are the lines of the grammar
  block in the specification, in the same order. The actions are code of the
  target language.
- The check of 4.4 has two parts for the grammar. lemongrass prints the rules
  that have an action first. The first part therefore sorts this list and the
  lines of the block, and then compares them. The second part reads the order
  of the rules from the grammar file and compares it with the block.
- The parser stack has a limit. It is 100 entries when the grammar file gives
  none. A statement can need one entry for each token, for example with `NOT`
  many times in a row. With 100 entries the parser refuses a correct statement
  that has 150 `NOT` in a row. A binding therefore turns the limit of the
  parser off. The library counts the tokens before the parser runs, so the
  stack stays at about MaxTokens entries or below.
- To find the token kinds that can come next (see 4.6), a binding makes a
  trial. A new parser reads the tokens before the failing token and then one
  token of the kind in question.
- The templates for Go and Rust have no error symbol and no destructors. AQL
  needs no error symbol, because the library gives the parser no token after
  the first one that the parser refuses. Go and Rust free the values
  themselves.
- The build calls lemongrass without line directives and with the bare name
  of the grammar file. The parser file then has no path of the build in it,
  and a second run writes the same bytes.
- After a token that the parser refuses, it can still report an accepted
  statement for the rest of the input. A binding therefore stops at the
  first refused token.

### 4.12 One lexer generator for all bindings

The lexer generator is [re2c](https://re2c.org). It reads rules with regular
expressions and writes a tokenizer as code of the target language. An earlier
plan had the regular expression package of the standard library for Go and an
open choice for Rust.

Reasons:

- re2c writes tokenizers for Go and for Rust. Both bindings take theirs from
  the same definitions, so the two tokenizers find the same tokens.
- The tokenizer that it writes needs no library. The Rust tokenizer has no
  `unsafe` code.
- re2c takes the longest match, and the first rule wins a tie. The token table
  has the same rule.
- Both bindings take the classes for letters and spaces from the same file of
  re2c. They therefore agree on every character.
- re2c is in the public domain, like Lemon.
- Debian and Ubuntu have re2c as a package.

A test on 2026-11-27 with re2c 3.1 compared three tokenizers. Two came from
re2c, one for Go and one for Rust. The third applied the token table with the
regular expression package of Go. The inputs were the 132 texts between
backticks in the examples of the specification, 27 other statements and 80000
random texts. Each input was read with the strict and with the lenient
patterns. The three tokenizers gave the same tokens for every input. At that
time the token table had a notation of its own. A second run built the Go and
the Rust tokenizer from the two blocks of specification 0.6.0, with no change
by hand. It gave the same tokens again. A Go tokenizer from re2c 4.6 gave
them too.

What follows for a binding:

- The definitions of the token table are a block of the specification. Both
  bindings include this block as it is. Each binding then has one rule for
  each row, in the order of the rows, with the action in its language.
- The check of 4.4 compares the included definitions with the block of the
  specification, as text. It also compares the names and the order of the
  rules with the list of the rows.
- re2c reads the input as UTF-8.
- The build calls re2c without the date of the call and without line
  directives. A second run then writes the same bytes.
- The first line of the generated file names the version of re2c. The
  version also decides the classes of characters: from re2c 4.5 on they are
  those of Unicode 17.0.0. The hints of a binding name the pinned version.
- The generated file is large: 1.3 MB to 1.5 MB for the Go binding, with the
  strict and the lenient tokenizer. The class of letters covers all scripts,
  and this needs many states. A compiler needs about 30 seconds for the file.

### 4.13 Place in PCD

- libaql is one component with `Deployment: library`.
- A server has `Deployment: mcp-server` and binds to libaql through
  DEPENDENCIES as a pcd-artifact, with pinned version and specification hash.
- The token table and the grammar are fenced blocks in the specification. PCD
  needs no schema change for them.

## 5. Order of work

1. Skeleton and front end of the specification: tokens, grammar, syntax
   messages, limits
2. Static rules: names, types, labels, scope and join rules
3. Evaluation, with `JOIN`, `IN (SELECT ...)` and `HAVING`
4. Write statements
5. Adapter contract and a reference backend for the examples
6. Shell
7. Hints, translation and review for each binding: Go first, then Rust
8. aha-aql moves to libaql. Its recorded probe statements are the regression
   check.

A first translation of the front end to Go comes before step 2. It proves the
tools of 4.11 and 4.12 on the specification as it is, before more of the
specification is built on them.

## 6. Open points

- The second backend. A candidate is the mail index of mu. It would add a
  point to the contract: `ORDER BY` with `LIMIT` in the backend. That is
  allowed only when the backend applies all conditions exactly.
- The places in a condition where `MATCH` is allowed, for example under `OR`
  or `NOT`.
- A left join in version 1, in addition to the inner join.
- A package of lemongrass for the build of a binding.
- The Unicode version. Both bindings take the classes for letters and spaces
  from the same file of re2c, so the pinned version of re2c decides it. The
  specification names no Unicode version.
- A text that is not valid Unicode. Some languages allow such a text in a
  string, other languages do not. The specification does not say yet what the
  library does with it. A tokenizer from re2c takes a byte that is not valid
  UTF-8 for no character. Outside a text no row fits then. Inside a text the
  text does not end.
- The notation for catalog tables. It could extend the TYPE tables of PCD.
- Lint rules in PCD for the token table and the grammar block.
- A statement for actions that are not records.
