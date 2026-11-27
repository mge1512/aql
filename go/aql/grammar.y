// The grammar of the BEHAVIOR parse-statement, as input for the parser
// generator lemongrass. The rules are the lines of the grammar block of
// the specification, in their order, each on one line up to its period.
// The labels and the actions are added. An action keeps the indices of
// the tokens in the token list only; it reads no token.

%name aqlParse
%token_prefix tok
%token_type {int}
%default_type {int}
%extra_argument {*parseState}
%stack_size 0
%token LEFT RIGHT FULL OUTER OFFSET.

%type statement {*stmtNode}
%type read {*readNode}
%type select {*readNode}
%type first_item {*itemNode}
%type more_items {[]*itemNode}
%type items {[]*itemNode}
%type item {*itemNode}
%type call {*callNode}
%type argument {*callNode}
%type from {*fromNode}
%type source {*sourceNode}
%type tail {*tailNode}
%type where_opt {*condNode}
%type group_opt {[]int}
%type having_opt {*condNode}
%type order_opt {[]*keyNode}
%type words {[]int}
%type keys {[]*keyNode}
%type key {*keyNode}
%type sort_item {*keyNode}
%type condition {*condNode}
%type conjunction {*condNode}
%type negation {*condNode}
%type primary {*condNode}
%type comparison {*condNode}
%type operand {*operandNode}
%type values {[]int}
%type describe {*stmtNode}
%type insert {*stmtNode}
%type update {*stmtNode}
%type assignments {[]assignNode}
%type assignment {assignNode}
%type delete {*stmtNode}

%include {
// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f

package aql
}

%syntax_error { yyarg.refused = true }
%parse_accept { yyarg.accepted = true }

statement(A) ::= read(R) end. { A = &stmtNode{kind: KindSelect, read: R}; yyarg.root = A }
statement(A) ::= describe(D) end. { A = D; yyarg.root = A }
statement(A) ::= insert(I) end. { A = I; yyarg.root = A }
statement(A) ::= update(U) end. { A = U; yyarg.root = A }
statement(A) ::= delete(D) end. { A = D; yyarg.root = A }

end(A) ::=. { A = -1 }
end(A) ::= SEMI(S). { A = S }

read(A) ::= select(S). { A = S }
read(A) ::= from(F) tail(T). { A = &readNode{sel: -1, from: F, tail: T} }
read(A) ::= first_item(I) more_items(M) from(F) tail(T). { A = &readNode{sel: -1, items: append([]*itemNode{I}, M...), from: F, tail: T} }

select(A) ::= SELECT(S) items(I) from(F) tail(T). { A = &readNode{sel: S, items: I, from: F, tail: T} }
select(A) ::= SELECT(S) DISTINCT items(I) from(F) tail(T). { A = &readNode{sel: S, distinct: true, items: I, from: F, tail: T} }

first_item(A) ::= STAR(S). { A = &itemNode{kind: itemStar, tok: S, label: -1} }
first_item(A) ::= call(C) label_opt(L). { A = &itemNode{kind: itemCall, tok: C.fn, call: C, label: L} }

more_items(A) ::=. { A = nil }
more_items(A) ::= more_items(M) COMMA item(I). { A = append(M, I) }

items(A) ::= item(I). { A = []*itemNode{I} }
items(A) ::= items(L) COMMA item(I). { A = append(L, I) }

item(A) ::= STAR(S). { A = &itemNode{kind: itemStar, tok: S, label: -1} }
item(A) ::= QSTAR(Q). { A = &itemNode{kind: itemQStar, tok: Q, label: -1} }
item(A) ::= WORD(W) label_opt(L). { A = &itemNode{kind: itemWord, tok: W, label: L} }
item(A) ::= call(C) label_opt(L). { A = &itemNode{kind: itemCall, tok: C.fn, call: C, label: L} }

label_opt(A) ::=. { A = -1 }
label_opt(A) ::= AS WORD(W). { A = W }
label_opt(A) ::= AS TEXT(T). { A = T }

call(A) ::= WORD(W) LPAREN(L) argument(G) RPAREN. { A = G; A.fn = W; A.lparen = L }

argument(A) ::= STAR(S). { A = &callNode{argKind: argStar, arg: S} }
argument(A) ::= NUMBER(N). { A = &callNode{argKind: argNumber, arg: N} }
argument(A) ::= WORD(W). { A = &callNode{argKind: argWord, arg: W} }
argument(A) ::= DISTINCT WORD(W). { A = &callNode{argKind: argWord, arg: W, distinct: true} }

from(A) ::= FROM source(S). { A = &fromNode{src: S, join: -1} }
from(A) ::= FROM source(S) JOIN(J) source(T) ON WORD(L) EQ WORD(R). { A = &fromNode{src: S, join: J, other: T, left: L, right: R} }
from(A) ::= FROM source(S) INNER(J) JOIN source(T) ON WORD(L) EQ WORD(R). { A = &fromNode{src: S, join: J, other: T, left: L, right: R} }

source(A) ::= WORD(E). { A = &sourceNode{entity: E, alias: -1} }
source(A) ::= WORD(E) WORD(L). { A = &sourceNode{entity: E, alias: L} }
source(A) ::= WORD(E) AS WORD(L). { A = &sourceNode{entity: E, alias: L} }

tail(A) ::= where_opt(W) group_opt(G) having_opt(H) order_opt(O) limit_opt(L). { A = &tailNode{where: W, group: G, having: H, order: O, limit: L} }

where_opt(A) ::=. { A = nil }
where_opt(A) ::= WHERE condition(C). { A = C }

group_opt(A) ::=. { A = nil }
group_opt(A) ::= GROUP BY words(W). { A = W }

having_opt(A) ::=. { A = nil }
having_opt(A) ::= HAVING condition(C). { A = C }

order_opt(A) ::=. { A = nil }
order_opt(A) ::= ORDER BY keys(K). { A = K }

limit_opt(A) ::=. { A = -1 }
limit_opt(A) ::= LIMIT NUMBER(N). { A = N }

words(A) ::= WORD(W). { A = []int{W} }
words(A) ::= words(L) COMMA WORD(W). { A = append(L, W) }

keys(A) ::= key(K). { A = []*keyNode{K} }
keys(A) ::= keys(L) COMMA key(K). { A = append(L, K) }

key(A) ::= sort_item(S). { A = S }
key(A) ::= sort_item(S) ASC. { A = S }
key(A) ::= sort_item(S) DESC. { A = S; A.desc = true }

sort_item(A) ::= WORD(W). { A = &keyNode{kind: keyWord, tok: W} }
sort_item(A) ::= TEXT(T). { A = &keyNode{kind: keyText, tok: T} }
sort_item(A) ::= call(C). { A = &keyNode{kind: keyCall, tok: C.fn, call: C} }

condition(A) ::= conjunction(C). { A = C }
condition(A) ::= condition(L) OR conjunction(R). { A = joinCond(condOr, L, R) }

conjunction(A) ::= negation(N). { A = N }
conjunction(A) ::= conjunction(L) AND negation(R). { A = joinCond(condAnd, L, R) }

negation(A) ::= primary(P). { A = P }
negation(A) ::= NOT(N) negation(M). { A = &condNode{op: condNot, tok: N, kids: []*condNode{M}} }

primary(A) ::= comparison(C). { A = C }
primary(A) ::= LPAREN condition(C) RPAREN. { A = C; A.paren = true }

comparison(A) ::= operand(O) EQ(E) value(V). { A = cmpCond(O, formEq, E, -1, V) }
comparison(A) ::= operand(O) CMP(E) value(V). { A = cmpCond(O, formCmp, E, -1, V) }
comparison(A) ::= operand(O) IS(E) NULL. { A = cmpCond(O, formIsNull, E, -1) }
comparison(A) ::= operand(O) IS(E) NOT NULL. { A = cmpCond(O, formIsNotNull, E, -1) }
comparison(A) ::= operand(O) not_opt(N) IN(E) LPAREN values(V) RPAREN. { A = cmpCond(O, formInList, E, N, V...) }
comparison(A) ::= operand(O) not_opt(N) IN(E) LPAREN select(S) RPAREN. { A = cmpCond(O, formInQuery, E, N); A.cmp.sub = S }
comparison(A) ::= operand(O) not_opt(N) CONTAINS(E) value(V). { A = cmpCond(O, formContains, E, N, V) }
comparison(A) ::= operand(O) not_opt(N) LIKE(E) value(V). { A = cmpCond(O, formLike, E, N, V) }
comparison(A) ::= operand(O) not_opt(N) BETWEEN(E) value(V) AND value(W). { A = cmpCond(O, formBetween, E, N, V, W) }
comparison(A) ::= operand(O) not_opt(N) MATCH(E) TEXT(V). { A = cmpCond(O, formMatch, E, N, V) }

not_opt(A) ::=. { A = -1 }
not_opt(A) ::= NOT(N). { A = N }

operand(A) ::= WORD(W). { A = &operandNode{word: W} }
operand(A) ::= call(C). { A = &operandNode{word: -1, call: C} }

values(A) ::= value(V). { A = []int{V} }
values(A) ::= values(L) COMMA value(V). { A = append(L, V) }

value(A) ::= TEXT(T). { A = T }
value(A) ::= NUMBER(N). { A = N }
value(A) ::= TRUE(T). { A = T }
value(A) ::= FALSE(F). { A = F }
value(A) ::= NULL(N). { A = N }

describe(A) ::= DESCRIBE. { A = &stmtNode{kind: KindDescribe, entity: -1} }
describe(A) ::= DESCRIBE WORD(W). { A = &stmtNode{kind: KindDescribe, entity: W} }

insert(A) ::= INSERT INTO WORD(E) LPAREN words(C) RPAREN VALUES(V) LPAREN values(L) RPAREN. { A = &stmtNode{kind: KindInsert, entity: E, columns: C, values: L, valuesTok: V} }

update(A) ::= UPDATE WORD(E) SET assignments(S) WHERE condition(C). { A = &stmtNode{kind: KindUpdate, entity: E, assigns: S, where: C} }

assignments(A) ::= assignment(S). { A = []assignNode{S} }
assignments(A) ::= assignments(L) COMMA assignment(S). { A = append(L, S) }

assignment(A) ::= WORD(W) EQ value(V). { A = assignNode{field: W, value: V} }

delete(A) ::= DELETE FROM WORD(E) WHERE condition(C). { A = &stmtNode{kind: KindDelete, entity: E, where: C} }
