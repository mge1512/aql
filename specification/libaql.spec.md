# libaql

## META
Deployment:   library
Version:      0.7.0
Spec-Schema:  0.4.0
Author:       Matthias G. Eckermann <mge@mailbox.org>
License:      MIT
Verification: none
Safety-Level: QM

This specification describes libaql, a library for MCP servers that offer
one query language, AQL, as their interface to the data behind an API. AQL
(API Query Language) is a small SQL dialect. A language model reads with
SELECT and DESCRIBE, and it changes one record with INSERT, UPDATE or
DELETE.

A server for one backend binds to libaql through its DEPENDENCIES section
and adds what is particular to that backend. Every such server reads a
statement in the same way, and it answers a mistake with the same message.

This Version has the front end of the library: the tokens, the grammar,
the limits, and the messages for a statement that cannot be read. It looks
up no name. An entity or a field stays as the statement writes it.

AQL uses only keywords that SQL also has.

## TYPES

```
MaxStatementLength := 65536
// Most characters of one statement.

MaxTokens := 1000
// Most tokens of one statement.

MaxNesting := 16
// Most parentheses that are open at one place of a statement.

MaxPatternLength := 200
// Most characters of the pattern of LIKE.

MaxEcho := 40
// Most characters of one word or text of the statement that a message
// repeats.

ExcerptLength := 24
// Number of characters of the statement that a message shows from the
// place where the statement cannot be read.

// A character is a Unicode code point. Positions and lengths count
// characters. The first character of a statement has the position 1.

StatementKind := select | describe | insert | update | delete
// The read set is select and describe. The write set is insert, update
// and delete.

Keyword := SELECT | DISTINCT | FROM | AS | JOIN | INNER | LEFT | RIGHT
         | FULL | OUTER | ON | WHERE | AND | OR | NOT | IS | NULL | IN
         | CONTAINS | LIKE | BETWEEN | MATCH | TRUE | FALSE | GROUP | BY
         | HAVING | ORDER | ASC | DESC | LIMIT | OFFSET | DESCRIBE
         | INSERT | INTO | VALUES | UPDATE | SET | DELETE

TokenKind := WORD | QSTAR | TEXT | NUMBER | EQ | CMP | LPAREN | RPAREN
           | COMMA | SEMI | STAR | Keyword | END
// END is no token of a statement. It stands for the place after the last
// token.

Token := {
  kind:     TokenKind,
  text:     string,     // the characters of the token as the statement
                        // writes them; for CMP the operator, with <> as !=
  value:    Text | Number or none,
                        // the content of a TEXT, the number of a NUMBER
  position: u32         // position of the first character of the token
}

Word := string
// A name as the statement writes it: an entity, a field, an alias, or a
// field with a qualifier and a period before it (t.name).

Value   := Text | Number | Boolean | Null
Text    := string
Number  := integer or decimal number
// A Number keeps every digit that the statement writes.
Boolean := true | false
Null    := no value

Aggregate := {
  function: count | sum | avg | min | max,
  argument: Word or none,        // none means COUNT(*)
  distinct: Boolean
}
Operand := Word | Aggregate

SelectItem    := Star | QualifiedStar | OperandItem
Star          := all fields
QualifiedStar := { qualifier: Word }               // t.*
OperandItem   := { operand: Operand, label: string or none }

Condition  := Comparison | NullTest | InList | InQuery | AllOf | AnyOf
            | Negation
Comparison := {
  operand:  Operand,
  operator: "=" | "!=" | "<" | "<=" | ">" | ">=" | "contains" | "like"
          | "match",
  value:    Text | Number | Boolean
}
NullTest := { operand: Operand, negated: Boolean }  // IS NULL, IS NOT NULL
InList   := { operand: Operand, values: List<Value> }
InQuery  := { operand: Operand, query: ReadStatement }
AllOf    := { members: List<Condition> }            // AND
AnyOf    := { members: List<Condition> }            // OR
Negation := { member: Condition }                   // NOT

Source   := { entity: Word, alias: Word or none }
Join     := { source: Source, left: Word, right: Word }   // ON left = right

QuotedLabel := { label: Text }                      // ORDER BY 'label'
OrderKey    := { key: Operand | QuotedLabel, descending: Boolean }

ReadStatement := {
  items:    List<SelectItem>,
  distinct: Boolean,
  source:   Source,
  join:     Join or none,
  where:    Condition or none,
  group_by: List<Word>,
  having:   Condition or none,
  order_by: List<OrderKey>,
  limit:    u32 or none
}
DescribeStatement := { entity: Word or none }
Assignment        := { field: Word, value: Text | Number | Boolean }
InsertStatement   := {
  entity:  Word,
  columns: List<Word>,
  values:  List<Text | Number | Boolean>   // one value for each column
}
UpdateStatement := {
  entity:      Word,
  assignments: List<Assignment>,           // SET field = value
  guards:      List<Assignment>            // WHERE field = value AND ...
}
DeleteStatement := { entity: Word, guards: List<Assignment> }
Statement := ReadStatement | DescribeStatement | InsertStatement
           | UpdateStatement | DeleteStatement

Guide := {
  entities: List<string>,   // the names of the entities of the server
  example:  string          // a read statement that the server accepts
}
// What a message needs from the server to name the entities and to show a
// statement that works.

ErrorKind := syntax
// syntax: The statement cannot be read, or it is not one of the accepted
//         forms

AqlError := { kind: ErrorKind, message: string }
```

## INTERFACES

```
Aql {
  // The public interface of libaql. A server binds to the built library
  // and calls these methods. A BEHAVIOR of this specification defines
  // each of them. The interface has the Version of this specification,
  // and the rules of the library deployment template for a public
  // interface apply. An implementation writes the names of the methods,
  // of the types and of their fields in the style of its target
  // language.
  required-methods:
    ParseStatement(text, allowed_kinds, guide) -> Statement | AqlError

  implementations-required:
    production:  the library itself
    test-double: none. The methods of this Version read nothing but
                 their inputs.
}
```

## BEHAVIOR: parse-statement
Constraint: required

Reads the text of one statement. It returns the statement, or an error
whose message says what to write instead. A statement that is correct as
written is read strictly. A statement that cannot be read strictly is read
once more with an apostrophe inside a word taken as part of the text.

INPUTS:
```
text:          string
allowed_kinds: set of StatementKind
guide:         Guide
```

OUTPUTS:
```
statement: Statement | AqlError
```

PRECONDITIONS:
- allowed_kinds is the read set or the write set
- guide.entities is not empty
- guide.example is a read statement that this BEHAVIOR accepts

STEPS:
1. The text as a whole. The two checks of this step read the text and keep no copy of it. A text that fails one of them gets no other work.
   a. Valid text. If text is not valid text → syntax, message S21 with the position of the first place that is not valid.
   b. Length. If text has more than MaxStatementLength characters → syntax, message L01.
2. Strict reading. Run tokenize with lenient false and read its tokens with steps 4 to 9. If this succeeds, its result is the result.
3. Second reading. It is made when step 2 failed and the text has an apostrophe directly between a letter or digit and a letter, in the classes of tokenize. Run tokenize with lenient true and read its tokens with steps 4 to 9. If this succeeds, its result is the result. If it fails → the error of step 2.
4. Nesting. For each token, count the parentheses that are open there: the LPAREN tokens up to and including the token, less the RPAREN tokens before it. If the count is above MaxNesting at a token → syntax, message L03 with the position of the first such token.
5. Take the kind from the first token.
   a. SELECT, FROM or STAR: select. A WORD that is directly followed by LPAREN: select.
   b. DESCRIBE: describe. INSERT: insert. UPDATE: update. DELETE: delete.
   c. No token, or any other first token. With the read set → syntax, message S04. With the write set → syntax, message W02.
6. If the kind is not allowed. A kind of the write set with the read set → syntax, message S07. A kind of the read set with the write set → syntax, message W01.
7. Function names. This step is for the kind select. A WORD that is directly followed by LPAREN must be COUNT, SUM, AVG, MIN or MAX, without regard to case. The first WORD that is not → syntax, message S12.
8. Parse the tokens with the grammar below. If the tokens are not a statement of the grammar → the error of classify-failure.
   MECHANISM: a parser generator builds the parser from the grammar. The grammar is LALR(1), and the generator reports no conflict.
9. Run check-forms on the tokens and on the statement that the grammar found. If it has a finding → its error. Otherwise build the Statement by the reading rules below.

Valid text. A text is valid when it is a sequence of characters and none
of them is a surrogate code point. In some programming languages a text
can have other content, for example bytes that are not UTF-8. In such a
language a text can be not valid. The position of its first place that is
not valid is the number of the characters before that place, plus 1. In a
language where every text is a sequence of characters, check 1a always
passes.

Grammar. It is written in the notation of the Lemon parser generator. A
name in capital letters is a token kind, and every other name is a rule.
Each line is one alternative of a rule: the name of the rule, the sign
`::=`, the token kinds and rules that the alternative reads, in their
order, and a period. An alternative with nothing after `::=` reads no
token. An empty line separates two rules and has no other meaning.
`statement` is the start rule.

```
statement ::= read end.
statement ::= describe end.
statement ::= insert end.
statement ::= update end.
statement ::= delete end.

end ::=.
end ::= SEMI.

read ::= select.
read ::= from tail.
read ::= first_item more_items from tail.

select ::= SELECT items from tail.
select ::= SELECT DISTINCT items from tail.

first_item ::= STAR.
first_item ::= call label_opt.

more_items ::=.
more_items ::= more_items COMMA item.

items ::= item.
items ::= items COMMA item.

item ::= STAR.
item ::= QSTAR.
item ::= WORD label_opt.
item ::= call label_opt.

label_opt ::=.
label_opt ::= AS WORD.
label_opt ::= AS TEXT.

call ::= WORD LPAREN argument RPAREN.

argument ::= STAR.
argument ::= NUMBER.
argument ::= WORD.
argument ::= DISTINCT WORD.

from ::= FROM source.
from ::= FROM source JOIN source ON WORD EQ WORD.
from ::= FROM source INNER JOIN source ON WORD EQ WORD.

source ::= WORD.
source ::= WORD WORD.
source ::= WORD AS WORD.

tail ::= where_opt group_opt having_opt order_opt limit_opt.

where_opt ::=.
where_opt ::= WHERE condition.

group_opt ::=.
group_opt ::= GROUP BY words.

having_opt ::=.
having_opt ::= HAVING condition.

order_opt ::=.
order_opt ::= ORDER BY keys.

limit_opt ::=.
limit_opt ::= LIMIT NUMBER.

words ::= WORD.
words ::= words COMMA WORD.

keys ::= key.
keys ::= keys COMMA key.

key ::= sort_item.
key ::= sort_item ASC.
key ::= sort_item DESC.

sort_item ::= WORD.
sort_item ::= TEXT.
sort_item ::= call.

condition ::= conjunction.
condition ::= condition OR conjunction.

conjunction ::= negation.
conjunction ::= conjunction AND negation.

negation ::= primary.
negation ::= NOT negation.

primary ::= comparison.
primary ::= LPAREN condition RPAREN.

comparison ::= operand EQ value.
comparison ::= operand CMP value.
comparison ::= operand IS NULL.
comparison ::= operand IS NOT NULL.
comparison ::= operand not_opt IN LPAREN values RPAREN.
comparison ::= operand not_opt IN LPAREN select RPAREN.
comparison ::= operand not_opt CONTAINS value.
comparison ::= operand not_opt LIKE value.
comparison ::= operand not_opt BETWEEN value AND value.
comparison ::= operand not_opt MATCH TEXT.

not_opt ::=.
not_opt ::= NOT.

operand ::= WORD.
operand ::= call.

values ::= value.
values ::= values COMMA value.

value ::= TEXT.
value ::= NUMBER.
value ::= TRUE.
value ::= FALSE.
value ::= NULL.

describe ::= DESCRIBE.
describe ::= DESCRIBE WORD.

insert ::= INSERT INTO WORD LPAREN words RPAREN VALUES LPAREN values RPAREN.

update ::= UPDATE WORD SET assignments WHERE condition.

assignments ::= assignment.
assignments ::= assignments COMMA assignment.

assignment ::= WORD EQ value.

delete ::= DELETE FROM WORD WHERE condition.
```

In the rules below, a call is what the rule `call` reads, and a subquery
is the `select` in parentheses after IN.

Reading rules. They say how the Statement follows from the rules of the
grammar.

- Without SELECT the item is `*`. A statement that starts with `*` or with a call is read as if SELECT stood before it.
- An item `t.*` is a QualifiedStar with the qualifier `t`.
- The label of an item is the WORD after AS, or the content of the TEXT after AS. An item without AS has no label.
- The function of a call is its WORD, read without regard to case. `COUNT(*)` and `COUNT(1)` have no argument. DISTINCT inside a call sets distinct.
- A source has the entity and, when a second WORD follows with or without AS, that WORD as its alias.
- JOIN and INNER JOIN are the same. The two WORDs of ON are the left and the right of the Join, in the order of the statement.
- NOT binds stronger than AND, and AND binds stronger than OR. Conditions that are joined by AND are the members of one AllOf, in their order. Conditions that are joined by OR are the members of one AnyOf. A condition in parentheses is that condition, and it is one member of the AllOf or AnyOf around it. `(a OR b) OR c` is therefore an AnyOf with two members, and its first member is the AnyOf of a and b. The parentheses add nothing else.
- NOT before a condition is a Negation of it. NOT before IN, CONTAINS, LIKE, BETWEEN or MATCH is a Negation of that comparison.
- EQ gives the operator `=`, and CMP gives the operator of its text. `operand = NULL` is read as `operand IS NULL`, and `operand != NULL` as `operand IS NOT NULL`.
- `operand BETWEEN a AND b` is read as an AllOf of `operand >= a` and `operand <= b`.
- IN with values is an InList. Its values keep their order, and NULL stays one of them. IN with a subquery is an InQuery.
- CONTAINS, LIKE and MATCH give the operators contains, like and match.
- TRUE and FALSE are the Boolean values, without regard to case.
- GROUP BY gives its WORDs in their order. A key of ORDER BY is a WORD, a call, or a QuotedLabel with the content of a TEXT. It is descending with DESC only.
- LIMIT gives its number.
- DESCRIBE has the entity of its WORD, and no entity without one.
- INSERT has the WORD after INTO as its entity. Its columns are the WORDs in the first parentheses, and its values are the values in the second parentheses, in their order.
- UPDATE has the WORD after UPDATE as its entity and one Assignment for each `WORD = value` after SET. UPDATE and DELETE have one guard for each comparison of their condition, in their order. check-forms makes sure that each of them has the form `WORD = value`.

POSTCONDITIONS:
- a text that the strict reading accepts is never read by the second reading
- no token is left after the statement, except one `;`
- the result has every WORD as the statement writes it

ERRORS:
- syntax: S21 and L01 for the text as a whole; S04, S07, W01 and W02 for the kind; L03; S12; every error of tokenize, of classify-failure and of check-forms

## BEHAVIOR/INTERNAL: tokenize
Constraint: required

Splits the text of a statement into tokens. The token table below says what
a token is.

INPUTS:
```
text:    string
lenient: Boolean    // second reading, see parse-statement
```

OUTPUTS:
```
tokens: List<Token> | AqlError
```

PRECONDITIONS:
- text is valid text, as parse-statement defines it
- text has at most MaxStatementLength characters

STEPS:
1. Start at the first character of text, and repeat steps 2 to 4 until no character is left.
2. Find the row of the token table for this place. A row fits when its pattern matches a run of characters that starts at this place. The run of a row is the longest run that its pattern matches there. The row with the longest run is the row of this place. When several rows have the longest run, the first of them in the table is the row. With lenient true, the rows text_1 and open_1 have their lenient patterns. If no row fits → syntax, message S02 with the character at this place.
3. Act by the kind of the row.
   a. skip: the run gives no token.
   b. open: the run is a text that does not end → syntax, message S01 with the position of the run.
   c. bad: the run starts like a number or a date and goes on with other characters → syntax, message S13 with the run as its `<word>`.
   d. Any other kind: the run is one token of that kind. Its values follow the rules for tokens below.
4. If the statement now has more than MaxTokens tokens → syntax, message L02. Otherwise go on with the character after the run.

Token table. It has two parts: definitions and rows. A definition gives a
pattern a name. A row has the name of a definition and a kind, and the
definition with that name is the pattern of the row.

Definitions. They are written in the notation of the lexer generator re2c.
A generator for this notation takes them over as they are.

```
// classes of characters
letter     = L;
digit      = [0-9];
space      = [\t\n\v\f\r\x20\u0085] | Z;
any        = [^];
apostrophe = ['];
quote      = ["];
backslash  = [\\];

// one piece of the content of a text
part_1       = [^'\\] | apostrophe apostrophe | backslash any;
part_2       = [^"\\] | quote quote | backslash any;
part_lenient = (letter | digit) (apostrophe letter)+;

// patterns of the rows
spaces   = space+;
text_1   = apostrophe part_1* apostrophe;
text_2   = quote part_2* quote;
date     = digit{4} "-" digit{2} "-" digit{2};
number   = "-"? digit+ ("." digit+)?;
qstar    = (letter | "_") (letter | digit | "_")* ".*";
word     = (letter | "_") (letter | digit | "_" | ".")*;
compare  = "!=" | "<>" | "<=" | ">=" | "<" | ">";
equals   = "=";
lparen   = "(";
rparen   = ")";
comma    = ",";
semi     = ";";
star     = "*";
open_1   = apostrophe part_1* backslash?;
open_2   = quote part_2* backslash?;
bad_date = date (letter | digit | "_" | "-")+;
bad_num  = number (letter | digit | "_" | ".")+;

// lenient patterns of the rows text_1 and open_1
text_1_lenient = apostrophe (part_lenient | part_1)* apostrophe;
open_1_lenient = apostrophe (part_lenient | part_1)* backslash?;
```

Notation of a definition. A definition has a name, the sign `=`, a regular
expression and a semicolon. A line that starts with `//` is a comment. In
a regular expression, a name stands for its definition, and text between
double quotes stands for itself. `[ ]` is one character of a set, and
`[^ ]` is one character that is not in the set, a line break included.
`[^]` is any character. In a set, `\t`, `\n`, `\v`, `\f` and `\r` are the
tab, the line feed, the vertical tab, the form feed and the carriage
return. `\x20` is the space, `\u0085` is the character U+0085, and `\\`
is the backslash. `( )` groups, `|` separates alternatives, and `*`, `+`,
`?` and `{n}` repeat what stands before them: any number of times, one or
more times, at most once, and n times. Parts that stand one after the
other match one after the other, and the space between two parts has no
meaning. `L` and `Z` are the Unicode general categories L and Z. The lexer
generator has their definitions.

Classes of characters. `letter` is a letter of any script. `digit` is one
of 0 1 2 3 4 5 6 7 8 9. `space` is a tab, a line feed, a vertical tab, a
form feed, a carriage return, a space, the character U+0085, or a
character of the Unicode general category Z. `any` is any character, a
line break included.

Rows. Each line has the name of a row and its kind. The order of the lines
is the order of the rows in the table.

```
spaces    skip
text_1    TEXT
text_2    TEXT
date      TEXT
number    NUMBER
qstar     QSTAR
word      WORD
compare   CMP
equals    EQ
lparen    LPAREN
rparen    RPAREN
comma     COMMA
semi      SEMI
star      STAR
open_1    open
open_2    open
bad_date  bad
bad_num   bad
```

Lenient patterns. With lenient true, the row text_1 has the pattern
`text_1_lenient` and the row open_1 has the pattern `open_1_lenient`. They
add one alternative in the first place: a letter or digit, followed by one
or more groups of an apostrophe and a letter.

Two ways of matching. Some tools look for the longest run of a pattern.
Other tools take the first alternative that fits and repeat as often as
they can. Every pattern of the table gives the same run in both ways, so
an implementation can use a tool of either kind.

Rules for tokens:

- The content of a text is its run without the first and the last character. In the content, the starting character written twice is one such character. A backslash followed by `n` is a line break, and a backslash followed by `t` is a tab. A backslash followed by any other character is that character. A line break inside a text is part of its content.
- A date without quotes is a text. Its content is its run.
- The value of a NUMBER is the number that its run writes.
- A WORD whose run is a Keyword, without regard to the case of the letters A to Z, has that Keyword as its kind.
- The text of a CMP token is its run, and `<>` is read as `!=`.
- White space is needed between two tokens only where the two runs would otherwise be one run.

POSTCONDITIONS:
- with lenient false, an apostrophe ends a text unless it is doubled or follows a backslash
- the runs of the tokens and of the skipped white space, in their order, are the text

ERRORS:
- syntax: a text that is not closed (S01); a character that starts no token (S02); a run that starts like a number or a date and is neither (S13); too many tokens (L02)

## BEHAVIOR/INTERNAL: classify-failure
Constraint: required

Gives the error for tokens that are not a statement of the grammar. It
finds the first token that cannot be read, and it looks at what could stand
in its place.

INPUTS:
```
tokens: List<Token>
kind:   StatementKind    // from parse-statement, step 5
guide:  Guide
```

OUTPUTS:
```
error: AqlError
```

PRECONDITIONS:
- the tokens are not a statement of the grammar of parse-statement
- kind is the kind that parse-statement took from the first token

STEPS:
1. Find the failing token O. It is the first token for which no statement of the grammar starts with the tokens up to and including it. When every token can be read and the statement is not complete, O is END.
2. P is the token directly before O, and P2 is the token directly before P. P and P2 can be missing.
3. A token kind K can come next when some statement of the grammar starts with the kinds of the tokens before O, followed by K. END can come next when the tokens before O are a statement of the grammar.
   MECHANISM: the generated parser of parse-statement gives the answer, and no other code decides what the grammar allows. The answer comes from the tables of the parser, or from a trial. In a trial the parser reads the tokens before O and then one token of the kind K. K can come next when the parser does not stop at that token.
4. Take the table of the kind: the read table for select, the describe table for describe, and the write table for insert, update and delete. The first row that applies gives the error → syntax, with the message of that row.

In the tables, `<where>` is the place of O, and "X can come next" is the
test of step 3. A row for the message S03 gives the text `<what>`.

Read table:

| Row | Applies when | Message |
|---|---|---|
| R01 | O is OFFSET. Or P2 is LIMIT, P is NUMBER, and O is COMMA or NUMBER | S10 |
| R02 | O is LEFT, RIGHT, FULL or OUTER, and END can come next | S08 |
| R03 | O is JOIN or INNER, and a JOIN token stands before O | S09 |
| R04 | O is SELECT and P is LPAREN. Or O is LPAREN and the token after O is SELECT | S11 |
| R05 | O is WORD, TEXT can come next, and WORD cannot come next | S13 |
| R06 | P is FROM or JOIN | S03, `an entity (<entity list>)` |
| R07 | P is DISTINCT | S03, `a field after DISTINCT` |
| R08 | P is SELECT | S03, `a field, * or COUNT(*) after SELECT` |
| R09 | P is AS | S03, `a name after AS` |
| R10 | P is GROUP | S03, `BY after GROUP` |
| R11 | P is ORDER | S03, `BY after ORDER` |
| R12 | P is LIMIT | S03, `a whole number after LIMIT` |
| R13 | P is INNER | S03, `JOIN after INNER` |
| R14 | P is IS. Or P is NOT and P2 is IS | S03, `NULL after IS` |
| R15 | P is IN | S03, `'(' after IN` |
| R16 | AND can come next, and OR cannot come next | S03, `AND in BETWEEN` |
| R17 | ON can come next | S03, `ON and the join condition, in the form: ON a.field = b.field` |
| R18 | EQ can come next, and CMP cannot come next | S03, `= after the field <field>` |
| R19 | FROM can come next | S03, `FROM and an entity` |
| R20 | CONTAINS can come next, and P is WORD | S03, `a comparison after the field <field> (=, !=, <, >, CONTAINS, LIKE, IN, IS NULL)` |
| R21 | CONTAINS can come next | S03, `a comparison (=, !=, <, >, CONTAINS, LIKE, IN, IS NULL)` |
| R22 | TEXT can come next, and WORD cannot come next | S03, `a value in single quotes` |
| R23 | STAR can come next, P is LPAREN, and P2 is the word COUNT | S03, `a field or * inside COUNT( )` |
| R24 | STAR can come next, and P is LPAREN | S03, `a field inside <FUNCTION>( )` |
| R25 | STAR can come next | S03, `a field, * or COUNT(*)` |
| R26 | RPAREN can come next | S03, `')'` |
| R27 | END can come next | S03, `the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order` |
| R28 | WORD can come next | S03, `a field name` |
| R29 | Every other case | S03 without `<what>` |

In R18 and R20, `<field>` is the text of P. In R23 and R24 the word is read
without regard to case, and `<FUNCTION>` is the text of P2 in capital
letters. In R02, `<KEYWORD>` of S08 is the text of O in capital letters.
In R05, `<word>` of S13 is the text of O.

Describe table:

| Row | Applies when | Message |
|---|---|---|
| D01 | The second token is a WORD, and O comes after it | S14 |
| D02 | Every other case | S17 |

In D01, `<entity>` of S14 is the text of the second token.

Write table. A row for the message W03 gives the text `<problem>`.

| Row | Applies when | Message |
|---|---|---|
| X01 | O is WORD, TEXT can come next, and WORD cannot come next | S13 |
| X02 | O is END or SEMI, and WHERE can come next | W04 |
| X03 | The kind is update, and WHERE can come next | W10 |
| X04 | The kind is delete, and WHERE can come next | W03, `expected WHERE after the entity.` |
| X05 | The kind is insert, O is COMMA, and END can come next | W07 |
| X06 | P is INSERT | W03, `expected INTO after INSERT.` |
| X07 | P is DELETE | W03, `expected FROM after DELETE.` |
| X08 | P is INTO or UPDATE. Or P is FROM and P2 is DELETE | W03, `expected an entity at <where>.` |
| X09 | SET can come next | W03, `expected SET after the entity.` |
| X10 | The kind is insert, LPAREN can come next, and P is VALUES | W03, `expected ( after VALUES.` |
| X11 | The kind is insert, and LPAREN can come next | W03, `expected ( and the field names after the entity.` |
| X12 | VALUES can come next | W03, `expected VALUES at <where>.` |
| X13 | EQ can come next, and P is WORD | W03, `expected = after the field <field>.` |
| X14 | TEXT can come next | W03, `expected a value in single quotes at <where>.` |
| X15 | The kind is insert, RPAREN can come next, and a VALUES token stands before O | W03, `expected ) after the values at <where>.` |
| X16 | The kind is insert, and RPAREN can come next | W03, `expected ) after the field names at <where>.` |
| X17 | WORD can come next | W03, `expected a field name at <where>.` |
| X18 | END can come next | W03, `unexpected text at <where>. One statement, one record.` |
| X19 | Every other case | W03, `cannot read the statement at <where>.` |

In X13, `<field>` is the text of P. In X01, `<word>` of S13 is the text of
O. In X02, `<KEYWORD>` of W04 is UPDATE or DELETE, by the kind.

POSTCONDITIONS:
- the result is an error for every input that meets the preconditions
- the failing token does not depend on the tokens after it

ERRORS:
- syntax: S03, S08 to S11, S13, S14 and S17; W03, W04, W07 and W10

## BEHAVIOR/INTERNAL: check-forms
Constraint: required

Looks at a statement that the grammar accepts and finds the forms that AQL
does not have. The grammar reads a little more than AQL accepts, so that
each of these forms gets a message of its own.

INPUTS:
```
tokens: List<Token>
kind:   StatementKind
guide:  Guide
```

OUTPUTS:
```
finding: none | AqlError
```

PRECONDITIONS:
- the tokens are a statement of the grammar of parse-statement

STEPS:
1. Collect the findings of the table below. Each finding has a token at which it shows.
2. If there is no finding, the result is none.
3. Otherwise take the finding whose token comes first in the statement. When two findings show at one token, take the one whose row stands first in the table → syntax, with the message of that row.

| Row | Finding | Shows at | Message |
|---|---|---|---|
| C01 | A call of SUM, AVG, MIN or MAX with the argument `*` | the `*` | S03 at that token, `a field inside <FUNCTION>( )` |
| C02 | A call of COUNT with a NUMBER as its argument whose run is not `1` | the NUMBER | S03 at that token, `a field or * inside COUNT( )` |
| C03 | A call of SUM, AVG, MIN or MAX with a NUMBER as its argument | the NUMBER | S03 at that token, `a field inside <FUNCTION>( )` |
| C04 | The kind is select, and a call is the operand of a comparison in a WHERE condition | the WORD of the call | S18 |
| C05 | A subquery inside a subquery | the SELECT of the inner subquery | S19 |
| C06 | A subquery with more than one item, or with the item `*` or `t.*` | the SELECT of the subquery | S20 |
| C07 | LIMIT with a NUMBER whose run starts with `-`, has a `.`, or has more than 9 digits | the NUMBER | S03 at that token, `a whole number after LIMIT` |
| C08 | NULL as the value after a CMP other than `!=`, after CONTAINS or after LIKE, or as a value of BETWEEN | the NULL | S03 at that token, `a value in single quotes` |
| C09 | LIKE with a TEXT whose content has more than MaxPatternLength characters | the TEXT | L04 |
| C10 | The kind is insert or update, and NULL is a value of VALUES or the value of an assignment | the NULL | W11 |
| C11 | The kind is insert, and the number of columns is not the number of values | the VALUES | W08 |
| C12 | The kind is update or delete, and the WHERE condition has an OR, a NOT or an LPAREN | that token | W05 |
| C13 | The kind is update or delete, and the WHERE condition has a CMP, IS, IN, CONTAINS, LIKE, BETWEEN or MATCH, or NULL as a value | that token | W06 |

In C01 and C03, `<FUNCTION>` is the function of the call in capital
letters. In C04, `<FUNCTION>` of S18 is that function too. In C11, `<n>` of
W08 is the number of columns and `<m>` the number of values. The WHERE
condition of a subquery counts for C04 like the one of the statement. A
HAVING condition is no WHERE condition.

C02 and C07 read the run of the NUMBER as it is written, not its value.
`COUNT(01)`, `LIMIT -0`, `LIMIT 2.0` and `LIMIT 0000000001` are therefore
findings. For C12 and C13, the WHERE condition is every token after the
WHERE of the statement. The tokens of a subquery there are part of it.

POSTCONDITIONS:
- with the result none, the condition of an UPDATE or a DELETE is one comparison of the form `WORD = value`, or several of them joined by AND
- with the result none, no value of a write statement is NULL

ERRORS:
- syntax: S03, S18, S19, S20; L04; W05, W06, W08, W11

## BEHAVIOR/INTERNAL: syntax-messages
Constraint: required

The texts of the errors of this specification. A language model learns
from an error what to write next. Every message therefore shows a form or
a statement that parse-statement accepts, or it says in plain words what
to send instead.

INPUTS:
```
id:     the ID of a message in the tables below
values: the values for its placeholders
guide:  Guide
```

OUTPUTS:
```
message: string
```

PRECONDITIONS:
- none

STEPS:
1. Take the text of the ID from the tables below and replace every placeholder. The ID itself is not part of the message.
2. A placeholder that repeats a word of the statement has at most MaxEcho characters of it. For a longer word it has the first MaxEcho characters, followed by three periods. When a period of the message stands directly after such a placeholder, the three periods take its place. This rule is for `<word>`, `<field>`, `<entity>`, `<NAME>`, `<FUNCTION>` and `<KEYWORD>`. The last two are names of this specification and never have more than MaxEcho characters.
3. `<NAME>`, `<FUNCTION>` and `<KEYWORD>` are in capital letters. Only the letters a to z change, and every other character stays as the statement has it.
4. A hint that is not empty starts with one space, so that it can follow the period of the sentence before it.

The messages speak of the query tool and the execute tool. They are the two
tools of a server that uses libaql: query reads, and execute changes one
record.

Placeholders that several messages use:

| Placeholder | Text |
|---|---|
| `<example>` | guide.example |
| `<entity list>` | The names of guide.entities, in their order, joined by a comma and a space |
| `<first entity>` | The first name of guide.entities |
| `<read form>` | `SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]` |
| `<write forms>` | `INSERT INTO entity (field, ...) VALUES ('value', ...)`, `UPDATE entity SET field = 'value', ... WHERE id = 'id'` and `DELETE FROM entity WHERE id = 'id'`, joined by a space, a vertical bar and a space |
| `<where>` | `the end of the statement` when the place is END. Otherwise `position <n> (<text>)`, with the position of the token and the next ExcerptLength characters of the statement from there, without white space at their two ends |
| `<expected>` | `: expected <what>` when the row gives a `<what>`, and empty otherwise |
| `<apostrophe hint>` | Empty, unless the statement has an apostrophe directly between a letter or digit and a letter. Then: ` The text contains an apostrophe (<run>). Inside text, write an apostrophe twice: <run doubled>`. `<run>` is the run of characters without white space around the first such apostrophe, without the characters ' " % ( ) , ; at its two ends. `<run doubled>` is that run with each of its apostrophes written twice. When `<run>` has more than MaxEcho characters, the hint is empty |

Messages for the limits:

| ID | Kind | Text |
|---|---|---|
| L01 | syntax | `the statement has <n> characters, and a statement has at most <max>. Send a shorter statement.` |
| L02 | syntax | `the statement has more than <max> words, values and signs. Send a shorter statement, for example with fewer values in a list.` |
| L03 | syntax | `the parentheses at position <n> are nested more than <max> levels deep. Write the statement with fewer parentheses.` |
| L04 | syntax | `the pattern of LIKE has <n> characters, and a pattern has at most <max>. Use a shorter pattern, or CONTAINS with a part of the text.` |

In L01 to L04, `<max>` is MaxStatementLength, MaxTokens, MaxNesting and
MaxPatternLength, in this order.

Messages for a statement that cannot be read:

| ID | Kind | Text |
|---|---|---|
| S01 | syntax | `the text that starts with the <quote> at position <n> is not closed.<text hint>` |
| S02 | syntax | `unexpected character <c> at position <n>. Remove it, or put the text in single quotes: 'like this'.<apostrophe hint>` |
| S03 | syntax | `cannot read the statement at <where><expected>.<apostrophe hint> Form: <read form>. Example: <example>` |
| S04 | syntax | `a statement starts with SELECT or DESCRIBE.<apostrophe hint> Form: <read form>. Example: <example>` |
| S07 | syntax | `query only reads. Use the execute tool for INSERT, UPDATE and DELETE.` |
| S08 | syntax | `<KEYWORD> JOIN is not supported. JOIN returns the records that have a partner in the other entity. Form: SELECT fields FROM entity a JOIN entity b ON a.field = b.field` |
| S09 | syntax | `a statement has one JOIN and reads two entities. For a third entity, send a second statement, or use: field IN (SELECT field FROM entity WHERE ...)` |
| S10 | syntax | `OFFSET is not supported. Use WHERE to narrow the result and LIMIT to shorten it.` |
| S11 | syntax | `a subquery stands only after IN, in the form: field IN (SELECT field FROM entity WHERE ...). Send any other subquery as its own statement first.` |
| S12 | syntax | `unknown function <NAME>. Functions: COUNT, SUM, AVG, MIN, MAX. Example: SELECT COUNT(*) FROM <first entity>` |
| S13 | syntax | `text goes in single quotes: write '<word>' instead of <word>` |
| S14 | syntax | `DESCRIBE <entity> lists the fields of <entity> and takes nothing else. To read records: <example>` |
| S17 | syntax | `DESCRIBE takes the name of one entity: <entity list>. DESCRIBE alone lists the entities.` |
| S18 | syntax | `<FUNCTION>( ) cannot stand in WHERE. Use it after GROUP BY, in the form: SELECT field, COUNT(*) FROM entity GROUP BY field HAVING COUNT(*) > 1` |
| S19 | syntax | `a subquery cannot contain another subquery. Send the inner one as its own statement first.` |
| S20 | syntax | `a subquery selects one field, in the form: field IN (SELECT field FROM entity WHERE ...)` |
| S21 | syntax | `the statement is not valid Unicode text at position <n>. Send the statement as plain text in UTF-8.` |

The IDs are not continuous.

In S01, `<quote>` is `apostrophe` or `double quote`, and `<n>` is the
position of that character. `<text hint>` is the apostrophe hint when that
is not empty. Otherwise it is ` Text goes in single quotes: 'like this'.`
for a text that starts with an apostrophe, and ` Use single quotes for
text: 'like this'.` for a text that starts with a double quote.

In S02, `<c>` is the character and `<n>` its position. In S13 from tokenize,
`<word>` is the run of the row. In S21, `<n>` is the position of the first
place that is not valid.

In S12, `<NAME>` is the WORD before the LPAREN.

Messages of write statements:

| ID | Kind | Text |
|---|---|---|
| W01 | syntax | `execute only changes data. Use the query tool for SELECT and DESCRIBE.` |
| W02 | syntax | `a statement starts with INSERT, UPDATE or DELETE.<apostrophe hint> Forms: <write forms>` |
| W03 | syntax | `<problem><apostrophe hint> Forms: <write forms>` |
| W04 | syntax | `<KEYWORD> needs WHERE id = '...' with the id of the one record to change.<apostrophe hint> Forms: <write forms>` |
| W05 | syntax | `one statement changes one record: the WHERE must start with id = '...'. Further conditions are joined with AND and have the form field = 'value'. To change several records, find their ids with the query tool and send one statement per record.` |
| W06 | syntax | `after WHERE id = '...', further conditions must have the form AND field = 'value'. For a field without a value write: AND field = ''` |
| W07 | syntax | `one INSERT creates one record. Send one statement per record.` |
| W08 | syntax | `<n> fields but <m> values. Write one value for each field: INSERT INTO entity (field, ...) VALUES ('value', ...)` |
| W10 | syntax | `a field can only be set to a value, not to an expression` |
| W11 | syntax | `NULL cannot be written. Leave the field out.` |

POSTCONDITIONS:
- no message has more than MaxEcho characters of one word of the statement, and `<where>` has at most ExcerptLength characters of it

ERRORS:
- none

## TOOLCHAIN-CONSTRAINTS

```
RUNTIME-DEPENDENCIES:
  - external-libraries: forbidden
    reason: every server that uses libaql takes its code over. The
      library uses the standard library of the target language and
      no other library.

FRONT-END:
  - generated-parser: required
    // A parser generator builds the parser from the grammar of
    // parse-statement. The generator reads every LALR(1) grammar and
    // reports conflicts. The build fails when it reports one. The
    // parser that it writes needs the standard library of the target
    // language only. Where the parser has a limit for its stack, the
    // limit is above what a statement with MaxTokens tokens needs.
  - table-driven-tokenizer: required
    // The tokenizer comes from the token table of tokenize. A lexer
    // generator builds it from the patterns at build time. The
    // tokenizer that it writes needs the standard library of the
    // target language only. Its running time grows linearly with the
    // length of the text.
  - hand-written-scanner-or-parser: forbidden
    // No code of the implementation decides by rules of its own what a
    // token is, or whether tokens are a statement.
  - spec-parity-check: required
    // A check in the build compares the patterns of the token table and
    // the rules of the grammar in the implementation with this
    // specification. The build fails when they differ.
```

## DEPENDENCIES

none. libaql uses the standard library of the target language and no
other library (TOOLCHAIN-CONSTRAINTS: RUNTIME-DEPENDENCIES). A parser
generator and a lexer generator are tools of the build.

## PRECONDITIONS

- For ParseStatement: allowed_kinds is the read set or the write set
- For ParseStatement: guide.entities is not empty, and guide.example is a
  read statement that parse-statement accepts

## POSTCONDITIONS

- No method reads or writes a file, makes a network call or reads an
  environment variable
- Every method is deterministic: the same inputs give the same result

## INVARIANTS

- [observable] parse-statement returns a Statement or an AqlError for every text, and the kind of every error of this Version is syntax
- [observable] a text that the strict reading accepts is never read by the second reading
- [observable] text that follows a complete statement is an error, for every statement kind
- [observable] every syntax error shows a form or a statement, or says in plain words what to send instead, and every statement that a message shows as its example is accepted by parse-statement
- [observable] no message repeats more than MaxEcho characters of one word of the statement, and no message shows more than ExcerptLength characters from the place of a failure
- [observable] a statement with more than MaxStatementLength characters, with more than MaxTokens tokens or with more than MaxNesting open parentheses is refused with the message of that limit
- [observable] a text that is not valid text is refused with the message S21, and no other check comes before this one
- [observable] the failing token of a statement and the kinds that can come next follow from the grammar alone
- [implementation] the parser is built from the grammar of parse-statement by a parser generator, and the generator reports no conflict
- [implementation] the patterns and the rules in the implementation are those of the token table and of the grammar of this specification, and a check in the build compares them
- [implementation] the implementation has no scanner and no parser that is written by hand
- [implementation] the checks of step 1 of parse-statement keep no copy of the text, and a text that they refuse gets no other work

## EXAMPLES

The fixture guide has the entities tickets, projects and users, in this
order. Its example is this statement:

`SELECT id, name, status FROM tickets WHERE project = 'P-1'`

A statement and a message are shown between backticks, which are not part
of them. An accepted statement is shown with the names of its type. A
member that is not shown is none, false or an empty list. An item without
a label is shown as its operand, and a Comparison as `operand operator
value`. `COUNT(*)` stands for the Aggregate with the function count and no
argument.

### EXAMPLE: read_statement_with_every_clause (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT DISTINCT status, COUNT(*) AS n FROM tickets WHERE priority >= 2 GROUP BY status HAVING COUNT(*) > 5 ORDER BY n DESC, status LIMIT 10`
THEN:
  the result is the ReadStatement
    items:    [status, { operand: COUNT(*), label: n }]
    distinct: true
    source:   { entity: tickets }
    where:    priority >= 2
    group_by: [status]
    having:   COUNT(*) > 5
    order_by: [{ key: n, descending: true }, { key: status }]
    limit:    10

### EXAMPLE: select_is_implied (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `FROM tickets WHERE status = 'open'`
THEN:
  the result is the ReadStatement
    items:  [Star]
    source: { entity: tickets }
    where:  status = 'open'
WHEN:
  parse-statement reads, with the read set,
  `COUNT(*) FROM tickets`
THEN:
  the result is the ReadStatement
    items:  [COUNT(*)]
    source: { entity: tickets }

### EXAMPLE: join_of_two_entities (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT t.name, p.name AS project FROM tickets t JOIN projects p ON t.project = p.id WHERE p.status = 'active'`
THEN:
  the result is the ReadStatement
    items:  [t.name, { operand: p.name, label: project }]
    source: { entity: tickets, alias: t }
    join:   { source: { entity: projects, alias: p }, left: t.project, right: p.id }
    where:  p.status = 'active'

### EXAMPLE: subquery_after_in (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT name FROM tickets WHERE project IN (SELECT id FROM projects WHERE owner = 'kim')`
THEN:
  the result is the ReadStatement
    items:  [name]
    source: { entity: tickets }
    where:  InQuery { operand: project, query: Q }
  and Q is the ReadStatement
    items:  [id]
    source: { entity: projects }
    where:  owner = 'kim'

### EXAMPLE: match_condition (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE body MATCH 'budget review' AND status != 'closed'`
THEN:
  the result is the ReadStatement
    items:  [id]
    source: { entity: tickets }
    where:  AllOf [ body match 'budget review', status != 'closed' ]

### EXAMPLE: between_and_null_are_rewritten (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE due BETWEEN 2026-01-01 AND 2026-03-31 AND owner = NULL`
THEN:
  the result is the ReadStatement
    items:  [id]
    source: { entity: tickets }
    where:  AllOf [ AllOf [ due >= '2026-01-01', due <= '2026-03-31' ], NullTest { operand: owner } ]
  the two dates without quotes are texts

### EXAMPLE: not_before_in_and_before_a_condition (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE status NOT IN ('closed', 'done') AND NOT (priority = 1 OR priority = 2)`
THEN:
  the result is the ReadStatement
    items:  [id]
    source: { entity: tickets }
    where:  AllOf [ Negation { InList { operand: status, values: ['closed', 'done'] } }, Negation { AnyOf [ priority = 1, priority = 2 ] } ]

### EXAMPLE: parentheses_keep_their_group (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE (status = 'open' OR status = 'new') OR priority = 1`
THEN:
  the result is the ReadStatement
    items:  [id]
    source: { entity: tickets }
    where:  AnyOf [ AnyOf [ status = 'open', status = 'new' ], priority = 1 ]

### EXAMPLE: aggregates_with_distinct_and_label (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT COUNT(DISTINCT owner) AS owners, MAX(priority) FROM tickets`
THEN:
  the result is the ReadStatement
    items:  [{ operand: A, label: owners }, B]
    source: { entity: tickets }
  A is the Aggregate { function: count, argument: owner, distinct: true }
  B is the Aggregate { function: max, argument: priority }

### EXAMPLE: write_statements (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `INSERT INTO tickets (project, name, priority) VALUES ('P-1', 'Export to CSV', 2)`
THEN:
  the result is the InsertStatement
    entity:  tickets
    columns: [project, name, priority]
    values:  ['P-1', 'Export to CSV', 2]
WHEN:
  parse-statement reads, with the write set,
  `UPDATE tickets SET status = 'closed', owner = 'kim' WHERE id = 'T-17' AND status = 'open'`
THEN:
  the result is the UpdateStatement
    entity:      tickets
    assignments: [status = 'closed', owner = 'kim']
    guards:      [id = 'T-17', status = 'open']
WHEN:
  parse-statement reads, with the write set,
  `DELETE FROM tickets WHERE id = 'T-17';`
THEN:
  the result is the DeleteStatement
    entity: tickets
    guards: [id = 'T-17']

### EXAMPLE: describe_statements (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `DESCRIBE`
THEN:
  the result is the DescribeStatement without an entity
WHEN:
  parse-statement reads, with the read set,
  `describe Tickets;`
THEN:
  the result is the DescribeStatement { entity: Tickets }
  the word is as the statement writes it

### EXAMPLE: second_reading_takes_the_apostrophe (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE name = 'Rider's cup'`
  and the strict reading fails, and the second reading takes the apostrophe between r and s as part of the text
THEN:
  the result is the ReadStatement
    items:  [id]
    source: { entity: tickets }
    where:  a Comparison of name with the operator = and the text `Rider's cup`

### EXAMPLE: strict_reading_is_kept (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE code = 'A'OR code = 'B'`
  and the strict reading succeeds, although an apostrophe stands between the letters A and O
THEN:
  the result is the ReadStatement
    items:  [id]
    source: { entity: tickets }
    where:  AnyOf [ code = 'A', code = 'B' ]
  no second reading is made

### EXAMPLE: statement_at_the_token_limit (of: parse-statement)
GIVEN:
  the fixture guide
  a statement of 1000 tokens: `SELECT id FROM tickets WHERE `, then `NOT ` 992 times, then `id = 1`
WHEN:
  parse-statement reads this statement with the read set
THEN:
  the result is the ReadStatement
    items:  [id]
    source: { entity: tickets }
    where:  a Negation whose member is a Negation, and so on for 992 Negations. The member of the last one is `id = 1`

### EXAMPLE: statement_above_the_length_limit (of: parse-statement)
GIVEN:
  the fixture guide
  a statement of 70038 characters: `SELECT id FROM tickets WHERE name = '`, then the letter x 70000 times, then `'`
WHEN:
  parse-statement reads this statement with the read set
THEN:
  the result is an error of the kind syntax with the message
  `the statement has 70038 characters, and a statement has at most 65536. Send a shorter statement.`

### EXAMPLE: text_that_is_not_valid (of: parse-statement)
GIVEN:
  the fixture guide
  a text that is not valid, in a language where a text can have such content: `SELECT id FROM tickets WHERE name = '` in UTF-8, then one byte with the value 255, then `'`
WHEN:
  parse-statement reads this text with the read set
THEN:
  the result is an error of the kind syntax with the message
  `the statement is not valid Unicode text at position 38. Send the statement as plain text in UTF-8.`

### EXAMPLE: texts_with_quotes_inside (of: tokenize)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE name = 'Rider''s cup' AND note = "say \"hi\""`
  and tokenize reads the doubled apostrophe as one apostrophe, and a backslash with a double quote as a double quote
THEN:
  the result is a ReadStatement whose condition is an AllOf of two Comparisons
  the first compares name with the text `Rider's cup`
  the second compares note with the text `say "hi"`

### EXAMPLE: keywords_in_small_letters (of: tokenize)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `select t.* from tickets t inner join projects p on t.project = p.id where p.name <> "Apollo" order by t.name;`
  and tokenize reads the keywords without regard to case, `t.*` as one token, and `<>` as `!=`
THEN:
  the result is the ReadStatement
    items:    [QualifiedStar { qualifier: t }]
    source:   { entity: tickets, alias: t }
    join:     { source: { entity: projects, alias: p }, left: t.project, right: p.id }
    where:    p.name != 'Apollo'
    order_by: [{ key: t.name }]

### EXAMPLE: numbers_and_a_date_without_quotes (of: tokenize)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT COUNT(1) FROM tickets WHERE estimate > -2.5 AND due >= 2026-11-26`
  and tokenize reads `-2.5` as a number and `2026-11-26` as a text
THEN:
  the result is the ReadStatement
    items:  [COUNT(*)]
    source: { entity: tickets }
    where:  AllOf [ estimate > -2.5, due >= '2026-11-26' ]

### EXAMPLE: statement_above_the_token_limit (of: tokenize)
GIVEN:
  the fixture guide
  a statement that starts with `SELECT id FROM tickets WHERE name IN (`, goes on with `'a', ` 600 times, and ends with `'z')`
WHEN:
  parse-statement reads this statement with the read set, and tokenize counts more than MaxTokens tokens
THEN:
  the result is an error of the kind syntax with the message
  `the statement has more than 1000 words, values and signs. Send a shorter statement, for example with fewer values in a list.`

### EXAMPLE: like_pattern_above_the_limit (of: check-forms)
GIVEN:
  the fixture guide
  a statement: `SELECT id FROM tickets WHERE name LIKE '`, then the letter x 201 times, then `'`
WHEN:
  parse-statement reads this statement with the read set, and check-forms takes the row C09
THEN:
  the result is an error of the kind syntax with the message
  `the pattern of LIKE has 201 characters, and a pattern has at most 200. Use a shorter pattern, or CONTAINS with a part of the text.`

### EXAMPLE: text_that_is_not_closed (of: tokenize)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE name = 'open`
  and tokenize takes the row open_1
THEN:
  the result is an error of the kind syntax with the message
  `the text that starts with the apostrophe at position 37 is not closed. Text goes in single quotes: 'like this'.`

### EXAMPLE: character_without_a_row (of: tokenize)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT [name] FROM tickets`
  and tokenize finds no row for the character [
THEN:
  the result is an error of the kind syntax with the message
  `unexpected character [ at position 8. Remove it, or put the text in single quotes: 'like this'.`

### EXAMPLE: number_followed_by_letters (of: tokenize)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE estimate > 12abc`
  and tokenize takes the row bad_num, whose run 12abc is longer than the run 12 of the row number
THEN:
  the result is an error of the kind syntax with the message
  `text goes in single quotes: write '12abc' instead of 12abc`

### EXAMPLE: statement_without_a_kind (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `tickets`
THEN:
  the result is an error of the kind syntax with the message
  `a statement starts with SELECT or DESCRIBE. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: write_statement_with_the_read_set (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `UPDATE tickets SET status = 'closed' WHERE id = 'T-17'`
THEN:
  the result is an error of the kind syntax with the message
  `query only reads. Use the execute tool for INSERT, UPDATE and DELETE.`

### EXAMPLE: read_statement_with_the_write_set (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `SELECT * FROM tickets`
THEN:
  the result is an error of the kind syntax with the message
  `execute only changes data. Use the query tool for SELECT and DESCRIBE.`

### EXAMPLE: unknown_function (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE COALESCE(owner, 'x') = 'x'`
THEN:
  the result is an error of the kind syntax with the message
  `unknown function COALESCE. Functions: COUNT, SUM, AVG, MIN, MAX. Example: SELECT COUNT(*) FROM tickets`

### EXAMPLE: second_reading_fails_too (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE name = 'Rider's`
  and the strict reading fails at the word s, and the second reading finds a text that does not end
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 44 (s): expected the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order. The text contains an apostrophe (Rider's). Inside text, write an apostrophe twice: Rider''s Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: nesting_above_the_limit (of: parse-statement)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE (((((((((((((((((a = 1)))))))))))))))))`
  which has 17 open parentheses at the position 46
THEN:
  the result is an error of the kind syntax with the message
  `the parentheses at position 46 are nested more than 16 levels deep. Write the statement with fewer parentheses.`

### EXAMPLE: from_is_missing (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT name WHERE status = 'open'`
  and classify-failure finds O = WHERE, and FROM, AS, LPAREN and COMMA can come next (row R19)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 13 (WHERE status = 'open'): expected FROM and an entity. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: entity_is_missing (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM`
  and classify-failure finds O = END and P = FROM (row R06)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at the end of the statement: expected an entity (tickets, projects, users). Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: text_without_quotes (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets WHERE status = open`
  and classify-failure finds O = WORD, where a TEXT can come next and a WORD cannot (row R05)
THEN:
  the result is an error of the kind syntax with the message
  `text goes in single quotes: write 'open' instead of open`

### EXAMPLE: is_without_null (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets WHERE status IS 'open'`
  and classify-failure finds P = IS (row R14)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 39 ('open'): expected NULL after IS. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: left_join_is_refused (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets t LEFT JOIN projects p ON t.project = p.id`
  and classify-failure finds O = LEFT (row R02)
THEN:
  the result is an error of the kind syntax with the message
  `LEFT JOIN is not supported. JOIN returns the records that have a partner in the other entity. Form: SELECT fields FROM entity a JOIN entity b ON a.field = b.field`

### EXAMPLE: second_join_is_refused (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets t JOIN projects p ON t.project = p.id JOIN users u ON t.owner = u.id`
  and classify-failure finds O = JOIN after an earlier JOIN (row R03)
THEN:
  the result is an error of the kind syntax with the message
  `a statement has one JOIN and reads two entities. For a third entity, send a second statement, or use: field IN (SELECT field FROM entity WHERE ...)`

### EXAMPLE: offset_is_refused (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets LIMIT 10 OFFSET 20`
  and classify-failure finds O = OFFSET (row R01)
THEN:
  the result is an error of the kind syntax with the message
  `OFFSET is not supported. Use WHERE to narrow the result and LIMIT to shorten it.`

### EXAMPLE: limit_with_two_numbers (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets LIMIT 10, 20`
  and classify-failure finds O = COMMA after LIMIT and a NUMBER (row R01)
THEN:
  the result is an error of the kind syntax with the message
  `OFFSET is not supported. Use WHERE to narrow the result and LIMIT to shorten it.`

### EXAMPLE: subquery_as_a_value (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets WHERE priority = (SELECT MAX(priority) FROM tickets)`
  and classify-failure finds O = LPAREN before SELECT (row R04)
THEN:
  the result is an error of the kind syntax with the message
  `a subquery stands only after IN, in the form: field IN (SELECT field FROM entity WHERE ...). Send any other subquery as its own statement first.`

### EXAMPLE: clauses_in_another_order (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets ORDER BY name WHERE status = 'open'`
  and classify-failure finds O = WHERE where END can come next (row R27)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 37 (WHERE status = 'open'): expected the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: field_without_a_comparison (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets WHERE status 'open'`
  and classify-failure finds P = WORD where CONTAINS can come next (row R20)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 36 ('open'): expected a comparison after the field status (=, !=, <, >, CONTAINS, LIKE, IN, IS NULL). Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: between_without_and (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets WHERE due BETWEEN '2026-01-01' , '2026-02-01'`
  and classify-failure finds that AND can come next and OR cannot (row R16)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 54 (, '2026-02-01'): expected AND in BETWEEN. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: join_without_on (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets t JOIN projects p WHERE t.project = p.id`
  and classify-failure finds that ON can come next (row R17)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 41 (WHERE t.project = p.id): expected ON and the join condition, in the form: ON a.field = b.field. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: parenthesis_that_is_not_closed (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets WHERE (status = 'open'`
  and classify-failure finds O = END where RPAREN can come next (row R26)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at the end of the statement: expected ')'. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: count_without_an_argument (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT COUNT() FROM tickets`
  and classify-failure finds P = LPAREN after the word COUNT (row R23)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 14 () FROM tickets): expected a field or * inside COUNT( ). Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: second_statement_after_the_first (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets; SELECT * FROM projects`
  and classify-failure finds O = SELECT after the semicolon (row R27)
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 24 (SELECT * FROM projects): expected the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: describe_with_more_text (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `DESCRIBE tickets WHERE status = 'open'`
  and classify-failure takes the row D01
THEN:
  the result is an error of the kind syntax with the message
  `DESCRIBE tickets lists the fields of tickets and takes nothing else. To read records: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: describe_with_a_text (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `DESCRIBE 'tickets'`
  and classify-failure takes the row D02
THEN:
  the result is an error of the kind syntax with the message
  `DESCRIBE takes the name of one entity: tickets, projects, users. DESCRIBE alone lists the entities.`

### EXAMPLE: insert_without_into (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `INSERT tickets (name) VALUES ('x')`
  and classify-failure finds P = INSERT (row X06)
THEN:
  the result is an error of the kind syntax with the message
  `expected INTO after INSERT. Forms: INSERT INTO entity (field, ...) VALUES ('value', ...) | UPDATE entity SET field = 'value', ... WHERE id = 'id' | DELETE FROM entity WHERE id = 'id'`

### EXAMPLE: insert_without_field_names (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `INSERT INTO tickets VALUES ('x')`
  and classify-failure finds that LPAREN can come next (row X11)
THEN:
  the result is an error of the kind syntax with the message
  `expected ( and the field names after the entity. Forms: INSERT INTO entity (field, ...) VALUES ('value', ...) | UPDATE entity SET field = 'value', ... WHERE id = 'id' | DELETE FROM entity WHERE id = 'id'`

### EXAMPLE: insert_with_two_rows (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `INSERT INTO tickets (name) VALUES ('x'), ('y')`
  and classify-failure finds O = COMMA where END can come next (row X05)
THEN:
  the result is an error of the kind syntax with the message
  `one INSERT creates one record. Send one statement per record.`

### EXAMPLE: update_without_where (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `UPDATE tickets SET status = 'closed'`
  and classify-failure finds O = END where WHERE can come next (row X02)
THEN:
  the result is an error of the kind syntax with the message
  `UPDATE needs WHERE id = '...' with the id of the one record to change. Forms: INSERT INTO entity (field, ...) VALUES ('value', ...) | UPDATE entity SET field = 'value', ... WHERE id = 'id' | DELETE FROM entity WHERE id = 'id'`

### EXAMPLE: update_with_an_expression (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `UPDATE tickets SET priority = 2 * 3 WHERE id = 'T-17'`
  and classify-failure finds O = STAR where WHERE can come next (row X03)
THEN:
  the result is an error of the kind syntax with the message
  `a field can only be set to a value, not to an expression`

### EXAMPLE: delete_with_an_alias (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `DELETE FROM tickets t WHERE id = 'T-17'`
  and classify-failure finds O = WORD where WHERE can come next (row X04)
THEN:
  the result is an error of the kind syntax with the message
  `expected WHERE after the entity. Forms: INSERT INTO entity (field, ...) VALUES ('value', ...) | UPDATE entity SET field = 'value', ... WHERE id = 'id' | DELETE FROM entity WHERE id = 'id'`

### EXAMPLE: text_after_a_write_statement (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `DELETE FROM tickets WHERE id = 'T-17' LIMIT 1`
  and classify-failure finds O = LIMIT where END can come next (row X18)
THEN:
  the result is an error of the kind syntax with the message
  `unexpected text at position 39 (LIMIT 1). One statement, one record. Forms: INSERT INTO entity (field, ...) VALUES ('value', ...) | UPDATE entity SET field = 'value', ... WHERE id = 'id' | DELETE FROM entity WHERE id = 'id'`

### EXAMPLE: value_without_quotes_in_insert (of: classify-failure)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `INSERT INTO tickets (name) VALUES (open)`
  and classify-failure finds O = WORD, where a TEXT can come next and a WORD cannot (row X01)
THEN:
  the result is an error of the kind syntax with the message
  `text goes in single quotes: write 'open' instead of open`

### EXAMPLE: aggregate_in_where (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT status FROM tickets WHERE COUNT(*) > 5 GROUP BY status`
  and check-forms finds a call in the WHERE condition (row C04)
THEN:
  the result is an error of the kind syntax with the message
  `COUNT( ) cannot stand in WHERE. Use it after GROUP BY, in the form: SELECT field, COUNT(*) FROM entity GROUP BY field HAVING COUNT(*) > 1`

### EXAMPLE: subquery_inside_a_subquery (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE project IN (SELECT id FROM projects WHERE owner IN (SELECT id FROM users WHERE team = 'x'))`
  and check-forms finds the second subquery (row C05)
THEN:
  the result is an error of the kind syntax with the message
  `a subquery cannot contain another subquery. Send the inner one as its own statement first.`

### EXAMPLE: subquery_with_two_items (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT id FROM tickets WHERE project IN (SELECT id, name FROM projects)`
  and check-forms finds two items in the subquery (row C06)
THEN:
  the result is an error of the kind syntax with the message
  `a subquery selects one field, in the form: field IN (SELECT field FROM entity WHERE ...)`

### EXAMPLE: star_inside_sum (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT SUM(*) FROM tickets`
  and check-forms takes the row C01
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 12 (*) FROM tickets): expected a field inside SUM( ). Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: limit_that_is_not_whole (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets LIMIT 2.5`
  and check-forms takes the row C07
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 29 (2.5): expected a whole number after LIMIT. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: number_is_read_as_written (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets LIMIT 10.0`
  and check-forms takes the row C07, because the run of the NUMBER has a point
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 29 (10.0): expected a whole number after LIMIT. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: count_of_a_number_is_read_as_written (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT COUNT(01) FROM tickets`
  and check-forms takes the row C02, because the run of the NUMBER is not 1
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 14 (01) FROM tickets): expected a field or * inside COUNT( ). Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: null_after_an_ordering_operator (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets WHERE priority > NULL`
  and check-forms takes the row C08
THEN:
  the result is an error of the kind syntax with the message
  `cannot read the statement at position 40 (NULL): expected a value in single quotes. Form: SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]. Example: SELECT id, name, status FROM tickets WHERE project = 'P-1'`

### EXAMPLE: null_as_a_written_value (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `UPDATE tickets SET owner = NULL WHERE id = 'T-17'`
  and check-forms takes the row C10
THEN:
  the result is an error of the kind syntax with the message
  `NULL cannot be written. Leave the field out.`

### EXAMPLE: columns_and_values_differ (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `INSERT INTO tickets (name, priority) VALUES ('x')`
  and check-forms takes the row C11
THEN:
  the result is an error of the kind syntax with the message
  `2 fields but 1 values. Write one value for each field: INSERT INTO entity (field, ...) VALUES ('value', ...)`

### EXAMPLE: or_in_the_where_of_a_write_statement (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `UPDATE tickets SET status = 'closed' WHERE id = 'T-17' OR id = 'T-18'`
  and check-forms takes the row C12
THEN:
  the result is an error of the kind syntax with the message
  `one statement changes one record: the WHERE must start with id = '...'. Further conditions are joined with AND and have the form field = 'value'. To change several records, find their ids with the query tool and send one statement per record.`

### EXAMPLE: other_comparison_in_a_guard (of: check-forms)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `UPDATE tickets SET status = 'closed' WHERE id = 'T-17' AND priority > 2`
  and check-forms takes the row C13
THEN:
  the result is an error of the kind syntax with the message
  `after WHERE id = '...', further conditions must have the form AND field = 'value'. For a field without a value write: AND field = ''`

### EXAMPLE: long_word_is_cut (of: syntax-messages)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT * FROM tickets WHERE status = abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwx`
  where the word after the = has 50 characters, and syntax-messages cuts it in the message
THEN:
  the result is an error of the kind syntax with the message
  `text goes in single quotes: write 'abcdefghijklmnopqrstuvwxyzabcdefghijklmn...' instead of abcdefghijklmnopqrstuvwxyzabcdefghijklmn...`

### EXAMPLE: long_function_name_is_cut (of: syntax-messages)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrs(estimate) FROM tickets`
  where the word before the ( has 45 characters. syntax-messages cuts it, and the three periods take the place of the period after it
THEN:
  the result is an error of the kind syntax with the message
  `unknown function ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMN... Functions: COUNT, SUM, AVG, MIN, MAX. Example: SELECT COUNT(*) FROM tickets`

### EXAMPLE: function_name_with_another_letter (of: syntax-messages)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the read set,
  `SELECT zähle(id) FROM tickets`
  and syntax-messages writes only the letters a to z of the name as capital letters
THEN:
  the result is an error of the kind syntax with the message
  `unknown function ZäHLE. Functions: COUNT, SUM, AVG, MIN, MAX. Example: SELECT COUNT(*) FROM tickets`

### EXAMPLE: apostrophe_hint_in_a_message (of: syntax-messages)
GIVEN:
  the fixture guide
WHEN:
  parse-statement reads, with the write set,
  `UPDATE tickets SET name = 'Rider's cup' status = 'open' WHERE id = 'T-17'`
  and the strict reading ends in a text that does not close, and the second reading fails at the word status. The error of the strict reading is the result, and its text hint is the apostrophe hint
THEN:
  the result is an error of the kind syntax with the message
  `the text that starts with the apostrophe at position 73 is not closed. The text contains an apostrophe (Rider's). Inside text, write an apostrophe twice: Rider''s`

## DEPLOYMENT

Runtime: a linkable library. It has no executable entry point, no command
line and no service of its own. A server for one backend uses it at build
time through the linkage of the target language. The library deployment
template and the hints resolve the language and the form of the artifact.

Target languages:
  This specification is written for more than one target language. It
  names no module and no version of a tool. It names two tools, the Lemon
  parser generator and the lexer generator re2c, only for the notation of
  the grammar and of the token table. The hints of each language name the
  tools of its build.

Front end:
  The token table of tokenize and the grammar of parse-statement are the
  source of the tokenizer and of the parser. A translator takes the
  definitions of the token table and the lines of the grammar over as
  they are, and adds what a rule has to build. It does not change a
  pattern, a rule or their order.

Consumers and version alignment:
  A server binds to this library through its DEPENDENCIES section, and it
  pins the Version and the hash of this specification. The build service
  builds the library from source.

Installation:
  Signed packages from the build service, as a dependency of the servers
  at build time. A user of a server needs no package of the library.

Platform:
  Linux is the primary platform. The platform statements of the servers
  say what else is supported.
