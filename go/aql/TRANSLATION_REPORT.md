# TRANSLATION_REPORT.md: libaql (Go)

Spec-SHA256: 2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
Spec-SHA256 (host): 2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
Included-Specs: none (META declares no `Includes:`)

| Path | SHA256 |
|------|--------|

LLM-Name: claude
Mode: translator
Date: 2026-11-28

`ROLE.md` holds `llm-name: <model-name>`, which is a placeholder. The
translator used `claude` for the directory `independent_tests/claude/` and
did not change `ROLE.md`.

## Translation Inputs (provenance)

- Spec-SHA256 (merged): 2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
- Spec-SHA256 (host): 2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
- Decisions-Hints-SHA256: specification/libaql.go.hints.md 4678ff3ac07f0f65f562fa7df5ae72d214a2b6a31bff4487f5c4351b7e6f45ae
- Milestones-Hints-SHA256: none
- Template-SHA256: pcd-translate/templates/library.template.md 3a70fd013352700230b61731021c40369e3a4da45953fd95ef2fd7028f4ce0b7 (Version 0.1.0)
- Prompt-SHA256: pcd-translate/prompts/prompt.md 7b54336daad95c91506b06ac419232067052a8ba893ee1ffa8d9486fc297607f
- Mode-SHA256: pcd-translate/modes/translator.md b61beef3f426e16dcc13dfbc5e448bda73b8cd4d943f320eac58bd455f1976ba
- Library-Hints-SHA256: .pcd/hints/library.go.lemongrass.hints.md e8931fb55149f9bf534a95be3b909ce4d66a48784d5822b946aa1cca45455dd8
- Library-Hints-SHA256: .pcd/hints/library.go.re2c.hints.md 07343f0cb281a6593079ed20b819c3290e490d44a9c0eb04e882da317f4efacd
- Style-Hints-SHA256: none
- Role-SHA256: specification/ROLE.md 167c5f73c7dc5327aaf76de7b26a6b5cf3fde39a4361d6755382137083355260
- Upgrade-Brief-SHA256: none (no `UPGRADE.md`; full translation into an empty `go/aql/`)
- Directive-SHA256: none
- Slice-Spec-Hash: spec:c200af55518fdfe36de34be5a2dfdcc822a16e31410c131bbd8390a6d7897c0f (`pcd-slice version`)
- Slice-Source-SHA256: source-sha256=2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
- Slice-Hints-SHA256: hints-sha256=4678ff3ac07f0f65f562fa7df5ae72d214a2b6a31bff4487f5c4351b7e6f45ae
- Slice-Manifest-SHA256: specification/spec.d/MANIFEST.tsv ffec15519939a3fb6cb432f6798f112f0b789fdd8423f114475f83200bcbe840
- Skill-SHA256: pcd-translate/SKILL.md 27b6947e8c205dba713156978a09d346d3b8d44013cb0ed02e5e291b81c2fd9e
- Skill-Manifest-SHA256: 71026a4e6a4154d27a2319176dd619b70d655ac92e85e25682768786c21378cf
- Toolchain: pcd-lint 0.4.0 cli (checked by `check-tools.sh`; not run separately)
- Toolchain: pcd-slice 0.1.1 cli
- Guard-Override: none

Guard: `check-tools.sh` exited 0. `pcd-slice check` was clean before the
first step and again after the last step, so the sources did not move
during the session. The translator read `spec.d/preamble.md` and the five
behavior bundles. It did not open `libaql.spec.md`. The build target
`check-spec` reads the specification as a tool, as the hints prescribe,
and the `tokens.re` definitions were extracted from the tokenize bundle.

## Tests-First-Compliance

yes. The seven files in `independent_tests/claude/` were written before
any implementation file. The structural check (7 test files, 0
implementation `.go` files) ran before `go.mod`, `tokens.re`, `lexer.re`
and the Go sources were written. `gofmt -w` aligned the spacing of two
test files before the first test run; no assertion changed.

## Continuity-Check

not applicable: there is no test-author input. This is a single-LLM run.

## Resolution

- Target language: Go. The library template declares no default; the
  decisions hints `libaql.go.hints.md` declare `Language: Go`. No preset
  overrides it.
- Module identity: `github.com/mge1512/aql/go/aql`, package `aql`, from
  source 2 (the language hints file). The spec META has no `Module:`
  field, and there is no prior manifest. There is no conflict.
- Delivery mode: filesystem. The files are written to `go/aql/` in the
  repository. Nothing was committed; the maintainer reviews and commits.
- Active MILESTONE: none. The specification has no MILESTONE section, so
  the full specification was translated.
- Constraint fields: every BEHAVIOR is `Constraint: required`. No
  behavior is `supported` or `forbidden`.
- INTERFACES: `Aql` with `ParseStatement`. Test double: none, as the
  specification declares ("The methods of this Version read nothing but
  their inputs").
- TYPE-BINDINGS / GENERATED-FILE-BINDINGS: the template has neither
  section. Types follow the Go mapping listed under Public API Surface.
- COMPONENT entries: the specification has no DELIVERABLES section.

## How STEPS ordering was applied

- tokenize (`tokenize.go`): steps 1 to 4 in order, in one loop. Step 2
  is `lexStrict` / `lexLenient` from re2c. The Go code acts only on the
  row kind it gets back (step 3a to 3d) and checks MaxTokens after each
  token (step 4). Keywords are found with the letters A to Z only
  (`capitals`). Positions add `utf8.RuneCountInString(run)`.
- parse-statement (`parse_statement.go`): step 1 (L01), step 2 (strict
  reading), step 3 (second reading only after a failed strict reading
  and only with an inner apostrophe; on failure the error of step 2).
  Inside `readStatement`: step 4 nesting, step 5 kind, step 6 allowed
  kinds, step 7 function names (select only), step 8 the lemongrass
  parser (`runParser`; on failure classify-failure), step 9 check-forms
  and then the reading rules (`builder`).
- check-forms (`check_forms.go`): step 1 collects the findings of all
  rows C01 to C13, each with its token index. Step 2 returns none. Step
  3 takes the smallest token index and, at the same token, the smallest
  row.
- classify-failure (`classify_failure.go`): step 1 is the token at which
  the parser ran `%syntax_error` (from `runParser`), or END. Step 2 is
  P and P2. Step 3 is `canComeNext`, a trial with a new parser, with a
  cache per failure. Step 4 is the three tables, row by row, in order.
- syntax-messages (`syntax_messages.go`): one function per message ID.
  `echo` cuts to MaxEcho characters plus three periods (step 2),
  `capitals` makes the capital letters (step 3), and the hint
  placeholders start with one space (step 4).
- MECHANISM of parse-statement step 8: `grammar.y` → lemongrass 0.1.1;
  the generator exits 0, so there is no conflict. MECHANISM of
  classify-failure step 3: trials with the generated parser only; no
  table is read by hand.

## Deliverables

| Deliverable | Files |
|---|---|
| source | `types.go`, `tokenize.go`, `parse_statement.go`, `check_forms.go`, `classify_failure.go`, `syntax_messages.go`; generator inputs `tokens.re`, `lexer.re`, `grammar.y`; generated `lexer.go` (re2c 4.6), `grammar.go` (lemongrass 0.1.1); `go.mod`; `VERSION` |
| spec parity | `check-spec.sh` (hints section 8, copied as it stands) |
| build | `Makefile` (hints section 3 verbatim, plus `SPEC_SHA256`, `VERSION`, `NAME`, `doc`, `dist`) |
| api-docs | Go doc comments; `make doc` runs `go doc -all .` |
| docs | `README.md` |
| license | `LICENSE` (SPDX identifier, URL, copyright line; no license text) |
| RPM | `libaql.spec` (subpackage `libaql-devel`, Go sources) |
| DEB | `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright` |
| tests | `independent_tests/claude/*_test.go` (7 files) |
| report | `TRANSLATION_REPORT.md` |

Spec-hash carriers: a header comment in every hand-written Go file and
test file, in `lexer.re` (which re2c copies into `lexer.go`), and in the
`%include` block of `grammar.y` (which lemongrass copies into
`grammar.go`). The hash is also in the `Makefile` variable `SPEC_SHA256`,
the `# pcd-spec-sha256:` comment of `libaql.spec`, the
`X-PCD-Spec-SHA256:` field of `debian/control`, `debian/rules`,
`debian/copyright`, `README.md` and `LICENSE`. There is no binary.

## Public API Surface

Module `github.com/mge1512/aql/go/aql`, package `aql`:

```
const MaxStatementLength = 65536
const MaxTokens = 1000
const MaxNesting = 16
const MaxPatternLength = 200
const MaxEcho = 40
const ExcerptLength = 24

func ParseStatement(text string, allowedKinds []StatementKind, guide Guide) (Statement, error)   // error is *AQLError
func ReadSet() []StatementKind
func WriteSet() []StatementKind

type StatementKind string
const KindSelect, KindDescribe, KindInsert, KindUpdate, KindDelete StatementKind   // "select" ... "delete"
type Word string
type Value interface{ isValue() }                 // Text | Number | Boolean | Null
type Text string
type Number string
type Boolean bool
type Null struct{}
type AggregateFunction string
const FunctionCount, FunctionSum, FunctionAvg, FunctionMin, FunctionMax AggregateFunction   // "count" ... "max"
type Aggregate struct{ Function AggregateFunction; Argument *Word; Distinct bool }
type Operand interface{ isOperand() }             // Word | Aggregate
type SelectItem interface{ isSelectItem() }       // Star | QualifiedStar | OperandItem
type Star struct{}
type QualifiedStar struct{ Qualifier Word }
type OperandItem struct{ Operand Operand; Label *string }
type Operator string
const OpEq, OpNe, OpLt, OpLe, OpGt, OpGe, OpContains, OpLike, OpMatch Operator   // "=" "!=" "<" "<=" ">" ">=" "contains" "like" "match"
type Condition interface{ isCondition() }         // Comparison | NullTest | InList | InQuery | AllOf | AnyOf | Negation
type Comparison struct{ Operand Operand; Operator Operator; Value Value }
type NullTest struct{ Operand Operand; Negated bool }
type InList struct{ Operand Operand; Values []Value }
type InQuery struct{ Operand Operand; Query ReadStatement }
type AllOf struct{ Members []Condition }
type AnyOf struct{ Members []Condition }
type Negation struct{ Member Condition }
type Source struct{ Entity Word; Alias *Word }
type Join struct{ Source Source; Left Word; Right Word }
type QuotedLabel struct{ Label Text }
type SortKey interface{ isSortKey() }             // Word | Aggregate | QuotedLabel
type OrderKey struct{ Key SortKey; Descending bool }
type Statement interface{ Kind() StatementKind }
type ReadStatement struct{ Items []SelectItem; Distinct bool; Source Source; Join *Join; Where Condition; GroupBy []Word; Having Condition; OrderBy []OrderKey; Limit *uint32 }
type DescribeStatement struct{ Entity *Word }
type Assignment struct{ Field Word; Value Value }
type InsertStatement struct{ Entity Word; Columns []Word; Values []Value }
type UpdateStatement struct{ Entity Word; Assignments []Assignment; Guards []Assignment }
type DeleteStatement struct{ Entity Word; Guards []Assignment }
func (ReadStatement) Kind() StatementKind
func (DescribeStatement) Kind() StatementKind
func (InsertStatement) Kind() StatementKind
func (UpdateStatement) Kind() StatementKind
func (DeleteStatement) Kind() StatementKind
type Guide struct{ Entities []string; Example string }
type ErrorKind string
const ErrorSyntax ErrorKind = "syntax"
type AQLError struct{ Kind ErrorKind; Message string }
func (e *AQLError) Error() string
```

Exported by the generated `grammar.go` (lemongrass template), but not
part of the interface: `YYACTIONTYPE`, `YYCODETYPE`, `YYMINORTYPE`, and
the constants `YYERRORSYMBOL`, `YYFALLBACK`, `YYNOCODE`, `YYNRULE`,
`YYNRULE_WITH_ACTION`, `YYNSTATE`, `YYNTOKEN`, `YYSTACKDEPTH`,
`YYWILDCARD`, `YY_ACCEPT_ACTION`, `YY_ACTTAB_COUNT`, `YY_ERROR_ACTION`,
`YY_MAX_DSTRCTR`, `YY_MAX_REDUCE`, `YY_MAX_SHIFT`, `YY_MAX_SHIFTREDUCE`,
`YY_MIN_DSTRCTR`, `YY_MIN_REDUCE`, `YY_MIN_SHIFTREDUCE`, `YY_NO_ACTION`,
`YY_REDUCE_COUNT`, `YY_REDUCE_MAX`, `YY_REDUCE_MIN`, `YY_SHIFT_COUNT`,
`YY_SHIFT_MAX`, `YY_SHIFT_MIN`. The hints forbid editing `grammar.go`. See
ambiguity 12.

The spec types `Token`, `TokenKind` and `Keyword` belong to internal
behaviors and are not exported (`token` in `tokenize.go`; the token
numbers are the `tok*` constants of the generated parser).

## Specification ambiguities encountered

1. tokenize bundle, hints section 4: a Go string with bytes that are not
   valid UTF-8. The specification does not say what happens. As the
   hints say, no check was added. Outside a text, the re2c rule `*` wins
   at such a byte, and the result is S02. The `<c>` of S02 is then U+FFFD
   (the decode result of the byte), because the byte alone would make the
   message invalid UTF-8. Inside a text, open_1/open_2 wins, and the
   result is S01.
2. syntax-messages, step 1 and step 2, message S12 `unknown function
   <NAME>. Functions: ...`: a cut `<NAME>` ends with `...`, and the
   period of the template follows it, which gives `FFF....`. The
   implementation does this literally. If three periods are wanted, the
   specification must say that the cut replaces the period that follows.
3. INVARIANTS, "every syntax error shows a form or a statement, or says in
   plain words what to send instead". S02 (`unexpected character <c> at
   position <n>.`, without a hint) and W08 (`<n> fields but <m> values`)
   do neither. The test `TestMessagesSayWhatToSend` leaves these two
   out. Either the invariant or these two texts needs a change.
4. check-forms C07, "negative, not whole, or more than 9 digits": the
   check reads the run literally. A leading `-` is negative (so is `-0`),
   any `.` is not whole (so is `2.0`), and every digit of the run counts
   (so does `0000000001`). A mathematical reading would accept `-0`,
   `2.0` and leading zeros.
5. check-forms C02, "other than COUNT(1)": the check compares the run with
   `1`. `COUNT(01)` and `COUNT(1.0)` are refused.
6. parse-statement reading rules, "Conditions that are joined by OR are
   the members of one AnyOf ... the parentheses add nothing". In this
   implementation, a chain in parentheses stays one member:
   `(a OR b) OR c` gives `AnyOf[AnyOf[a, b], c]`, and `a OR b OR c` gives
   `AnyOf[a, b, c]`. The other reading would flatten the first form too.
7. syntax-messages step 3, "in capital letters": only the letters a to z
   are changed, the same rule as keyword detection (hints section 4).
   Other letters keep their case.
8. syntax-messages step 2 lists `<word>`, `<field>`, `<entity>` and
   `<NAME>` for the MaxEcho cut. `<FUNCTION>` (S18, R24, C01, C03) and
   `<KEYWORD>` (S08) are cut as well, so that the INVARIANT on MaxEcho
   holds for every message. In this Version those values are short
   anyway.
9. INTERFACES, `allowed_kinds: set of StatementKind`: in Go this is
   `[]StatementKind`, with `ReadSet()` and `WriteSet()`. Steps 5c and 6
   tell the read set by membership of select or describe. The
   PRECONDITION makes this the same as comparing the sets.
10. check-forms C12/C13 "the WHERE condition": the implementation scans
    every token after the first WHERE of an UPDATE or DELETE. The tokens
    of a subquery after IN count too, but C13 at that IN comes first in
    any case.
11. Packaging: the hints put the complete result, `debian/` included,
    into `go/aql/`. The `dist` tarball has `specification/` and `go/aql/`
    at its top. `dpkg-buildpackage` expects `debian/` at the top of the
    source tree, so `debian/rules` uses paths `go/aql/...` that are
    relative to that top. The maintainer has to place `debian/` at the
    top of the unpacked tarball, or the hints have to say where it goes.
12. Public API surface (template PUBLIC-API-SURFACE): the lemongrass Go
    template exports about 30 `YY*` names into package `aql`. They cannot
    be removed without editing `grammar.go`, which the hints forbid.
    Proposal for `library.go.lemongrass.hints.md`, or for the lemongrass
    template: write these names in small letters, or state that they are
    outside the stable interface.

## Rules that could not be implemented exactly as written

- Hints section 5: "The value of `%extra_argument` has the token list,
  and an action reads the text and the value of a token there." The
  actions here keep token indices only and read no token. The tokens are
  read after the parse, when the statement is built. This also makes the
  trials of classify-failure safe: the trial token has an index outside
  the list, and no action reads it. `parseState` therefore has no token
  list.
- RPM and DEB: the package names (`libaql-devel`,
  `golang-github-mge1512-aql-dev`), the build-requirement names
  `lemongrass` and `re2c` with exact versions, and the install paths
  (`%{_datadir}/go/src/...` for RPM, `/usr/share/gocode/src/...` for
  DEB) follow common conventions. They were not verified on OBS. The
  maintainer has to check them against the Go packaging macros of the
  target distribution.

## Compile gate

pass. `make build && make test` ran in a fresh copy of
`specification/libaql.spec.md` and `go/aql/`:

- `check-tools`: lemongrass 0.1.1 and re2c 4.6 found
- `check-generated`: `grammar.go` and `lexer.go` are equal to fresh
  output of the tools
- `check-spec`: "the token table and the grammar are those of the
  specification"
- `go build ./...`: ok (local Go 1.27.4; `go.mod` says `go 1.24`)
- `go test ./...`: ok

`lemongrass` exit status 0 means that the grammar has no conflict. `go
vet ./...` is clean. `gofmt -l` is clean on every hand-written file. `make
dist` produced a tarball of 34 entries, which was then deleted. `make
doc` works. After the run, `git status` shows only the new `go/`
directory (and `ROLE.md` / `spec.d/`, which were present before the run).

## Test results: translator suite

`independent_tests/claude/`: 123 tests, 123 pass, 0 fail, 0 skip. No test
needs a live service.

tokenize bundle: TestTextsWithQuotesInside, TestKeywordsInSmallLetters,
TestNumbersAndADateWithoutQuotes, TestStatementAboveTheTokenLimit,
TestTextThatIsNotClosed, TestCharacterWithoutARow,
TestNumberFollowedByLetters, TestDoubleQuotedTextThatIsNotClosed,
TestDateFollowedByLetters, TestTokenLimitBoundary, TestTokenLimitPlusOne,
TestKeywordOnlyWithLettersAToZ, TestBackslashEscapes,
TestWhiteSpaceAndLineBreak, TestNumberKeepsEveryDigit,
TestPositionsCountCharacters, TestCompareOperators: pass.

parse-statement bundle: TestReadStatementWithEveryClause,
TestSelectIsImplied, TestJoinOfTwoEntities, TestSubqueryAfterIn,
TestMatchCondition, TestBetweenAndNullAreRewritten,
TestNotBeforeInAndBeforeACondition, TestAggregatesWithDistinctAndLabel,
TestWriteStatements, TestDescribeStatements,
TestSecondReadingTakesTheApostrophe, TestStrictReadingIsKept,
TestStatementAtTheTokenLimit, TestStatementAboveTheLengthLimit,
TestStatementWithoutAKind, TestWriteStatementWithTheReadSet,
TestReadStatementWithTheWriteSet, TestUnknownFunction,
TestSecondReadingFailsToo, TestNestingAboveTheLimit,
TestNestingAtTheLimit, TestLengthLimitBoundary,
TestTextAfterACompleteStatement, TestEmptyStatement,
TestFunctionNamesWithoutCase, TestLabelsAndKeys, TestNullsAndBooleans,
TestNotBeforeOperatorsAndPrecedence,
TestSecondReadingWithDigitBeforeApostrophe, TestUnknownFunctionInOrderBy,
TestGuideExampleIsAccepted, TestKindsOfTheSets: pass.

check-forms bundle: TestLikePatternAboveTheLimit, TestAggregateInWhere,
TestSubqueryInsideASubquery, TestSubqueryWithTwoItems, TestStarInsideSum,
TestLimitThatIsNotWhole, TestNullAfterAnOrderingOperator,
TestNullAsAWrittenValue, TestColumnsAndValuesDiffer,
TestOrInTheWhereOfAWriteStatement, TestOtherComparisonInAGuard,
TestLikePatternAtTheLimit, TestNumbersInsideCalls, TestLimitBoundaries,
TestNullAfterOtherOperators, TestCallInWhereOfASubquery,
TestSubqueryWithAStar, TestFirstFindingWins, TestNullInInsert,
TestMoreValuesThanColumns, TestGuardsOfDelete, TestGuardsWithOtherValues:
pass.

syntax-messages bundle: TestLongWordIsCut, TestApostropheHintInAMessage,
TestWordAtMaxEchoIsNotCut, TestLongUnknownFunctionIsCut,
TestExcerptIsCutAndTrimmed, TestApostropheHintAfterOtherMessages,
TestLongRunGivesNoHint, TestDoubledApostropheGivesNoHint: pass.

classify-failure bundle: TestFromIsMissing, TestEntityIsMissing,
TestTextWithoutQuotes, TestIsWithoutNull, TestLeftJoinIsRefused,
TestSecondJoinIsRefused, TestOffsetIsRefused, TestLimitWithTwoNumbers,
TestSubqueryAsAValue, TestClausesInAnotherOrder,
TestFieldWithoutAComparison, TestBetweenWithoutAnd, TestJoinWithoutOn,
TestParenthesisThatIsNotClosed, TestCountWithoutAnArgument,
TestSecondStatementAfterTheFirst, TestDescribeWithMoreText,
TestDescribeWithAText, TestInsertWithoutInto, TestInsertWithoutFieldNames,
TestInsertWithTwoRows, TestUpdateWithoutWhere, TestUpdateWithAnExpression,
TestDeleteWithAnAlias, TestTextAfterAWriteStatement,
TestValueWithoutQuotesInInsert, TestLimitWithTwoNumbersWithoutComma,
TestOuterJoinInSmallLetters, TestSecondInnerJoin,
TestSubqueryInParentheses, TestRowsForTheTokenBefore, TestJoinRows,
TestComparisonAfterACall, TestValueAfterAnOperator,
TestSumWithoutAnArgument, TestFieldNameAfterGroupBy,
TestDescribeWithANumber, TestWriteTableRows: pass.

INVARIANTS: TestEveryTextGivesAStatementOrAnError, TestMessagesAreShort,
TestMessagesSayWhatToSend (S02 and W08 left out, see ambiguity 3),
TestStrictReadingIsNeverReplaced, TestFailingTokenDoesNotDependOnTheRest,
TestWordsKeepTheirCase: pass.

The INVARIANT tests run over a corpus of more than 1000 texts: the example
statements, every prefix of them, and mutations, with both sets. A
separate throwaway probe, which is not a deliverable and was deleted, ran
400,000 random token sequences with invalid UTF-8 and NUL. It found no
panic and no result without exactly one of statement and error.

## Test results: test-author suite

not applicable: no test-author input.

## Test Refinements

| Test | Result before | Action | Rationale |
|---|---|---|---|
| TestLongUnknownFunctionIsCut | failed | test edited | The test expected `FFF... Functions`. syntax-messages step 1 replaces `<NAME>` in the text `unknown function <NAME>. Functions: ...`, and step 2 makes `<NAME>` the first MaxEcho characters followed by three periods. The period of the template follows, so the message is `FFF.... Functions`. See ambiguity 2. |
| all other 122 tests | passed | none | (none) |

Before any test run, `gofmt -w` aligned the spacing in `show_test.go`
and `tokenize_test.go`. No assertion changed, so these are not
refinements.

## Per-example confidence

Tests-First-Compliance is `yes`, and every named test passes without a
live service. The prompt's definition of High also asks for a
test-author suite. That suite is absent, so the confidence is Medium for
every EXAMPLE.

| EXAMPLE | Confidence | Verification method | Unverified claims |
|---|---|---|---|
| texts_with_quotes_inside | Medium | TestTextsWithQuotesInside | no independent suite |
| keywords_in_small_letters | Medium | TestKeywordsInSmallLetters | no independent suite |
| numbers_and_a_date_without_quotes | Medium | TestNumbersAndADateWithoutQuotes | no independent suite |
| statement_above_the_token_limit | Medium | TestStatementAboveTheTokenLimit | no independent suite |
| text_that_is_not_closed | Medium | TestTextThatIsNotClosed | no independent suite |
| character_without_a_row | Medium | TestCharacterWithoutARow | no independent suite |
| number_followed_by_letters | Medium | TestNumberFollowedByLetters | no independent suite |
| read_statement_with_every_clause | Medium | TestReadStatementWithEveryClause | no independent suite |
| select_is_implied | Medium | TestSelectIsImplied | no independent suite |
| join_of_two_entities | Medium | TestJoinOfTwoEntities | no independent suite |
| subquery_after_in | Medium | TestSubqueryAfterIn | no independent suite |
| match_condition | Medium | TestMatchCondition | no independent suite |
| between_and_null_are_rewritten | Medium | TestBetweenAndNullAreRewritten | no independent suite |
| not_before_in_and_before_a_condition | Medium | TestNotBeforeInAndBeforeACondition | no independent suite |
| aggregates_with_distinct_and_label | Medium | TestAggregatesWithDistinctAndLabel | no independent suite |
| write_statements | Medium | TestWriteStatements | no independent suite |
| describe_statements | Medium | TestDescribeStatements | no independent suite |
| second_reading_takes_the_apostrophe | Medium | TestSecondReadingTakesTheApostrophe | no independent suite |
| strict_reading_is_kept | Medium | TestStrictReadingIsKept | "no second reading is made" is shown by the result only, not by watching the reading |
| statement_at_the_token_limit | Medium | TestStatementAtTheTokenLimit | no independent suite |
| statement_above_the_length_limit | Medium | TestStatementAboveTheLengthLimit | no independent suite |
| statement_without_a_kind | Medium | TestStatementWithoutAKind | no independent suite |
| write_statement_with_the_read_set | Medium | TestWriteStatementWithTheReadSet | no independent suite |
| read_statement_with_the_write_set | Medium | TestReadStatementWithTheWriteSet | no independent suite |
| unknown_function | Medium | TestUnknownFunction | no independent suite |
| second_reading_fails_too | Medium | TestSecondReadingFailsToo | no independent suite |
| nesting_above_the_limit | Medium | TestNestingAboveTheLimit | no independent suite |
| like_pattern_above_the_limit | Medium | TestLikePatternAboveTheLimit | no independent suite |
| aggregate_in_where | Medium | TestAggregateInWhere | no independent suite |
| subquery_inside_a_subquery | Medium | TestSubqueryInsideASubquery | no independent suite |
| subquery_with_two_items | Medium | TestSubqueryWithTwoItems | no independent suite |
| star_inside_sum | Medium | TestStarInsideSum | no independent suite |
| limit_that_is_not_whole | Medium | TestLimitThatIsNotWhole | no independent suite |
| null_after_an_ordering_operator | Medium | TestNullAfterAnOrderingOperator | no independent suite |
| null_as_a_written_value | Medium | TestNullAsAWrittenValue | no independent suite |
| columns_and_values_differ | Medium | TestColumnsAndValuesDiffer | no independent suite |
| or_in_the_where_of_a_write_statement | Medium | TestOrInTheWhereOfAWriteStatement | no independent suite |
| other_comparison_in_a_guard | Medium | TestOtherComparisonInAGuard | no independent suite |
| long_word_is_cut | Medium | TestLongWordIsCut | no independent suite |
| apostrophe_hint_in_a_message | Medium | TestApostropheHintInAMessage | no independent suite |
| from_is_missing | Medium | TestFromIsMissing | no independent suite |
| entity_is_missing | Medium | TestEntityIsMissing | no independent suite |
| text_without_quotes | Medium | TestTextWithoutQuotes | no independent suite |
| is_without_null | Medium | TestIsWithoutNull | no independent suite |
| left_join_is_refused | Medium | TestLeftJoinIsRefused | no independent suite |
| second_join_is_refused | Medium | TestSecondJoinIsRefused | no independent suite |
| offset_is_refused | Medium | TestOffsetIsRefused | no independent suite |
| limit_with_two_numbers | Medium | TestLimitWithTwoNumbers | no independent suite |
| subquery_as_a_value | Medium | TestSubqueryAsAValue | no independent suite |
| clauses_in_another_order | Medium | TestClausesInAnotherOrder | no independent suite |
| field_without_a_comparison | Medium | TestFieldWithoutAComparison | no independent suite |
| between_without_and | Medium | TestBetweenWithoutAnd | no independent suite |
| join_without_on | Medium | TestJoinWithoutOn | no independent suite |
| parenthesis_that_is_not_closed | Medium | TestParenthesisThatIsNotClosed | no independent suite |
| count_without_an_argument | Medium | TestCountWithoutAnArgument | no independent suite |
| second_statement_after_the_first | Medium | TestSecondStatementAfterTheFirst | no independent suite |
| describe_with_more_text | Medium | TestDescribeWithMoreText | no independent suite |
| describe_with_a_text | Medium | TestDescribeWithAText | no independent suite |
| insert_without_into | Medium | TestInsertWithoutInto | no independent suite |
| insert_without_field_names | Medium | TestInsertWithoutFieldNames | no independent suite |
| insert_with_two_rows | Medium | TestInsertWithTwoRows | no independent suite |
| update_without_where | Medium | TestUpdateWithoutWhere | no independent suite |
| update_with_an_expression | Medium | TestUpdateWithAnExpression | no independent suite |
| delete_with_an_alias | Medium | TestDeleteWithAnAlias | no independent suite |
| text_after_a_write_statement | Medium | TestTextAfterAWriteStatement | no independent suite |
| value_without_quotes_in_insert | Medium | TestValueWithoutQuotesInInsert | no independent suite |

Claims that no test verifies: the POSTCONDITION "no method reads or
writes a file, makes a network call or reads an environment variable" is
shown only by code review. The package imports only `strconv`,
`strings`, `unicode` and `unicode/utf8`, and the generated files import
nothing. The [implementation] INVARIANTS are checked by the build, not by
tests: `check-spec` (parity), `check-generated` (no hand edits), and the
lemongrass exit status (no conflict).
