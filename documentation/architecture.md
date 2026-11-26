# AQL architecture

This document records the design decisions for AQL and libaql as of 2026-11-26.
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
| aql | this repository |
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

libaql has its own engine and no embedded SQL database. It is small, it adds
no runtime dependency, and we decide the wording of every message. As a
library it is translated and reviewed once, and every server uses the same
code.

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

- a token table with one regular expression per token. The longest match wins.
  The first rule wins a tie.
- a grammar in yacc notation without actions.

A binding generates the parser from the grammar and drives the token table with
a regular expression engine. No scanning or parsing logic is written by hand.
A check compares the patterns and rules in the code with the two blocks.

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

### 4.10 Binding for Go

- The parser comes from goyacc of `golang.org/x/tools`, at a pinned version.
  The versions 0.46.0 to 0.53.0 contain the same generator.
- The token table runs on the regular expression package of the standard
  library, which works in linear time.
- The generated parser file is part of the source. The package build generates
  it again and compares.
- goyacc reports a conflict and still ends with status 0. The build therefore
  checks the report.
- The first line of the generated file repeats the arguments of the goyacc
  call. The build therefore calls goyacc with fixed arguments.

The specification names none of these tools. They belong to the hints for Go.

### 4.11 Place in PCD

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
7. Hints for Go, translation, review
8. aha-aql moves to libaql. Its recorded probe statements are the regression
   check.

## 6. Open points

- The second backend. A candidate is the mail index of mu. It would add a
  point to the contract: `ORDER BY` with `LIMIT` in the backend. That is
  allowed only when the backend applies all conditions exactly.
- The places in a condition where `MATCH` is allowed, for example under `OR`
  or `NOT`.
- A left join in version 1, in addition to the inner join.
- The notation for catalog tables. It could extend the TYPE tables of PCD.
- Lint rules in PCD for the token table and the grammar block.
- A statement for actions that are not records.
