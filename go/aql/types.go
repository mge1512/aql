// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f

// Package aql is libaql, the front end of AQL (API Query Language) for MCP
// servers. AQL is a small SQL dialect: a language model reads with SELECT
// and DESCRIBE, and it changes one record with INSERT, UPDATE or DELETE.
//
// The package reads the text of one statement with ParseStatement. It
// returns the Statement, or an *AQLError whose message says what to write
// instead. The package looks up no name: an entity or a field stays as the
// statement writes it.
//
// The tokenizer comes from the lexer generator re2c and the parser from
// the parser generator lemongrass. The package uses the standard library
// of Go and no other library. It reads no file, makes no network call and
// reads no environment variable, and every function is deterministic.
package aql

// Limits of a statement. A character is a Unicode code point. Positions
// and lengths count characters; the first character of a statement has
// the position 1.
const (
	// MaxStatementLength is the largest number of characters of one
	// statement.
	MaxStatementLength = 65536
	// MaxTokens is the largest number of tokens of one statement.
	MaxTokens = 1000
	// MaxNesting is the largest number of parentheses that are open at
	// one place of a statement.
	MaxNesting = 16
	// MaxPatternLength is the largest number of characters of the
	// pattern of LIKE.
	MaxPatternLength = 200
	// MaxEcho is the largest number of characters of one word or text of
	// the statement that a message repeats.
	MaxEcho = 40
	// ExcerptLength is the number of characters of the statement that a
	// message shows from the place where the statement cannot be read.
	ExcerptLength = 24
)

// StatementKind is the kind of a statement.
type StatementKind string

// The statement kinds. The read set is select and describe; the write set
// is insert, update and delete.
const (
	KindSelect   StatementKind = "select"
	KindDescribe StatementKind = "describe"
	KindInsert   StatementKind = "insert"
	KindUpdate   StatementKind = "update"
	KindDelete   StatementKind = "delete"
)

// ReadSet returns the read set: select and describe. A query tool passes
// it to ParseStatement.
func ReadSet() []StatementKind { return []StatementKind{KindSelect, KindDescribe} }

// WriteSet returns the write set: insert, update and delete. An execute
// tool passes it to ParseStatement.
func WriteSet() []StatementKind { return []StatementKind{KindInsert, KindUpdate, KindDelete} }

// Word is a name as the statement writes it: an entity, a field, an alias,
// or a field with a qualifier and a period before it (t.name).
type Word string

// Value is a Text, a Number, a Boolean or a Null.
type Value interface{ isValue() }

// Text is the content of a text of the statement.
type Text string

// Number is an integer or decimal number. It keeps every digit that the
// statement writes, as the statement writes it.
type Number string

// Boolean is true or false.
type Boolean bool

// Null is no value.
type Null struct{}

func (Text) isValue()    {}
func (Number) isValue()  {}
func (Boolean) isValue() {}
func (Null) isValue()    {}

// AggregateFunction is the function of an Aggregate, in small letters.
type AggregateFunction string

// The aggregate functions.
const (
	FunctionCount AggregateFunction = "count"
	FunctionSum   AggregateFunction = "sum"
	FunctionAvg   AggregateFunction = "avg"
	FunctionMin   AggregateFunction = "min"
	FunctionMax   AggregateFunction = "max"
)

// Aggregate is a call of an aggregate function.
type Aggregate struct {
	Function AggregateFunction
	Argument *Word // nil means COUNT(*)
	Distinct bool
}

// Operand is a Word or an Aggregate.
type Operand interface{ isOperand() }

func (Word) isOperand()      {}
func (Aggregate) isOperand() {}

// SelectItem is a Star, a QualifiedStar or an OperandItem.
type SelectItem interface{ isSelectItem() }

// Star is the item *: all fields.
type Star struct{}

// QualifiedStar is the item t.*.
type QualifiedStar struct {
	Qualifier Word
}

// OperandItem is an item with an operand and an optional label.
type OperandItem struct {
	Operand Operand
	Label   *string // nil without AS
}

func (Star) isSelectItem()          {}
func (QualifiedStar) isSelectItem() {}
func (OperandItem) isSelectItem()   {}

// Operator is the operator of a Comparison.
type Operator string

// The operators of a Comparison.
const (
	OpEq       Operator = "="
	OpNe       Operator = "!="
	OpLt       Operator = "<"
	OpLe       Operator = "<="
	OpGt       Operator = ">"
	OpGe       Operator = ">="
	OpContains Operator = "contains"
	OpLike     Operator = "like"
	OpMatch    Operator = "match"
)

// Condition is a Comparison, a NullTest, an InList, an InQuery, an AllOf,
// an AnyOf or a Negation.
type Condition interface{ isCondition() }

// Comparison compares an operand with a value. The value is a Text, a
// Number or a Boolean.
type Comparison struct {
	Operand  Operand
	Operator Operator
	Value    Value
}

// NullTest is IS NULL, or IS NOT NULL when Negated is true.
type NullTest struct {
	Operand Operand
	Negated bool
}

// InList is operand IN (values). NULL stays one of the values.
type InList struct {
	Operand Operand
	Values  []Value
}

// InQuery is operand IN (subquery).
type InQuery struct {
	Operand Operand
	Query   ReadStatement
}

// AllOf is the conditions joined by AND, in their order.
type AllOf struct {
	Members []Condition
}

// AnyOf is the conditions joined by OR, in their order.
type AnyOf struct {
	Members []Condition
}

// Negation is NOT before a condition.
type Negation struct {
	Member Condition
}

func (Comparison) isCondition() {}
func (NullTest) isCondition()   {}
func (InList) isCondition()     {}
func (InQuery) isCondition()    {}
func (AllOf) isCondition()      {}
func (AnyOf) isCondition()      {}
func (Negation) isCondition()   {}

// Source is an entity with an optional alias.
type Source struct {
	Entity Word
	Alias  *Word // nil without an alias
}

// Join is JOIN source ON left = right.
type Join struct {
	Source Source
	Left   Word
	Right  Word
}

// QuotedLabel is a key of ORDER BY written as a text: ORDER BY 'label'.
type QuotedLabel struct {
	Label Text
}

// SortKey is the key of an OrderKey: a Word, an Aggregate or a
// QuotedLabel.
type SortKey interface{ isSortKey() }

func (Word) isSortKey()        {}
func (Aggregate) isSortKey()   {}
func (QuotedLabel) isSortKey() {}

// OrderKey is one key of ORDER BY.
type OrderKey struct {
	Key        SortKey
	Descending bool
}

// Statement is a ReadStatement, a DescribeStatement, an InsertStatement,
// an UpdateStatement or a DeleteStatement.
type Statement interface {
	// Kind returns the kind of the statement.
	Kind() StatementKind
}

// ReadStatement is a SELECT statement.
type ReadStatement struct {
	Items    []SelectItem
	Distinct bool
	Source   Source
	Join     *Join     // nil without JOIN
	Where    Condition // nil without WHERE
	GroupBy  []Word
	Having   Condition // nil without HAVING
	OrderBy  []OrderKey
	Limit    *uint32 // nil without LIMIT
}

// DescribeStatement is DESCRIBE, with or without an entity.
type DescribeStatement struct {
	Entity *Word // nil for DESCRIBE alone
}

// Assignment is field = value. The value is a Text, a Number or a Boolean.
type Assignment struct {
	Field Word
	Value Value
}

// InsertStatement is INSERT INTO entity (columns) VALUES (values). It has
// one value for each column.
type InsertStatement struct {
	Entity  Word
	Columns []Word
	Values  []Value
}

// UpdateStatement is UPDATE entity SET assignments WHERE guards.
type UpdateStatement struct {
	Entity      Word
	Assignments []Assignment // SET field = value
	Guards      []Assignment // WHERE field = value AND ...
}

// DeleteStatement is DELETE FROM entity WHERE guards.
type DeleteStatement struct {
	Entity Word
	Guards []Assignment
}

// Kind returns KindSelect.
func (ReadStatement) Kind() StatementKind { return KindSelect }

// Kind returns KindDescribe.
func (DescribeStatement) Kind() StatementKind { return KindDescribe }

// Kind returns KindInsert.
func (InsertStatement) Kind() StatementKind { return KindInsert }

// Kind returns KindUpdate.
func (UpdateStatement) Kind() StatementKind { return KindUpdate }

// Kind returns KindDelete.
func (DeleteStatement) Kind() StatementKind { return KindDelete }

// Guide is what a message needs from the server to name the entities and
// to show a statement that works.
type Guide struct {
	Entities []string // the names of the entities of the server
	Example  string   // a read statement that the server accepts
}

// ErrorKind is the kind of an AQLError.
type ErrorKind string

// ErrorSyntax: the statement cannot be read, or it is not one of the
// accepted forms.
const ErrorSyntax ErrorKind = "syntax"

// AQLError is the error of ParseStatement.
type AQLError struct {
	Kind    ErrorKind
	Message string
}

// Error returns the message.
func (e *AQLError) Error() string { return e.Message }
