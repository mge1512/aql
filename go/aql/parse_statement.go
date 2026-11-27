// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f

package aql

// BEHAVIOR parse-statement: reads the text of one statement. The parser
// comes from lemongrass (grammar.y); its actions build the nodes below,
// which keep the indices of tokens only. check-forms looks at the nodes,
// and the reading rules turn them into a Statement.

import (
	"strconv"
	"unicode/utf8"
)

// parseState is the extra argument of the generated parser.
type parseState struct {
	refused  bool
	accepted bool
	root     *stmtNode
}

// Nodes of the grammar. Every int is the index of a token in the token
// list, or -1 for none.

type stmtNode struct {
	kind      StatementKind
	read      *readNode
	entity    int
	columns   []int
	values    []int
	valuesTok int
	assigns   []assignNode
	where     *condNode
}

type readNode struct {
	sel      int // the SELECT, or -1 when SELECT is implied
	distinct bool
	items    []*itemNode // nil: the item is *
	from     *fromNode
	tail     *tailNode
}

const (
	itemStar = iota
	itemQStar
	itemWord
	itemCall
)

type itemNode struct {
	kind  int
	tok   int
	call  *callNode
	label int // the WORD or TEXT after AS
}

const (
	argStar = iota
	argNumber
	argWord
)

type callNode struct {
	fn       int // the WORD of the call
	lparen   int
	argKind  int
	arg      int
	distinct bool
}

type fromNode struct {
	src   *sourceNode
	join  int // the JOIN or INNER, or -1
	other *sourceNode
	left  int
	right int
}

type sourceNode struct {
	entity int
	alias  int
}

type tailNode struct {
	where  *condNode
	group  []int
	having *condNode
	order  []*keyNode
	limit  int
}

const (
	keyWord = iota
	keyText
	keyCall
)

type keyNode struct {
	kind int
	tok  int
	call *callNode
	desc bool
}

const (
	condOr = iota
	condAnd
	condNot
	condCmp
)

type condNode struct {
	op    int
	tok   int
	kids  []*condNode
	cmp   *cmpNode
	paren bool // the condition stands in parentheses
}

const (
	formEq = iota
	formCmp
	formIsNull
	formIsNotNull
	formInList
	formInQuery
	formContains
	formLike
	formBetween
	formMatch
)

type cmpNode struct {
	operand *operandNode
	form    int
	opTok   int   // the EQ, CMP, IS, IN, CONTAINS, LIKE, BETWEEN or MATCH
	notTok  int   // the NOT of not_opt, or -1
	values  []int // the values, in their order
	sub     *readNode
}

type operandNode struct {
	word int
	call *callNode
}

type assignNode struct {
	field int
	value int
}

// joinCond joins two conditions by AND or OR. Conditions joined by the
// same operator are the members of one node; a condition in parentheses
// stays one member.
func joinCond(op int, left, right *condNode) *condNode {
	if left.op == op && !left.paren {
		left.kids = append(left.kids, right)
		return left
	}
	return &condNode{op: op, tok: -1, kids: []*condNode{left, right}}
}

func cmpCond(operand *operandNode, form, opTok, notTok int, values ...int) *condNode {
	return &condNode{op: condCmp, tok: opTok, cmp: &cmpNode{
		operand: operand, form: form, opTok: opTok, notTok: notTok, values: values,
	}}
}

// ParseStatement reads the text of one statement. allowedKinds is the
// read set (ReadSet) or the write set (WriteSet); guide.Entities is not
// empty, and guide.Example is a read statement that ParseStatement
// accepts. The result is the Statement, or an error of the type *AQLError
// with the kind syntax.
func ParseStatement(text string, allowedKinds []StatementKind, guide Guide) (Statement, error) {
	m := newMessages(text, guide)
	// Step 1: length.
	if n := utf8.RuneCountInString(text); n > MaxStatementLength {
		return nil, syntaxError(msgL01(n))
	}
	// Step 2: strict reading.
	st, err := readStatement(text, false, allowedKinds, m)
	if err == nil {
		return st, nil
	}
	// Step 3: second reading.
	if innerApostrophe(m.runes) >= 0 {
		if st2, err2 := readStatement(text, true, allowedKinds, m); err2 == nil {
			return st2, nil
		}
	}
	return nil, err
}

func hasKind(kinds []StatementKind, k StatementKind) bool {
	for _, x := range kinds {
		if x == k {
			return true
		}
	}
	return false
}

var functionNames = map[string]bool{"COUNT": true, "SUM": true, "AVG": true, "MIN": true, "MAX": true}

// readStatement runs tokenize and steps 4 to 9.
func readStatement(text string, lenient bool, allowed []StatementKind, m *messages) (Statement, *AQLError) {
	tokens, err := tokenize(text, lenient, m)
	if err != nil {
		return nil, err
	}
	// Step 4: nesting.
	open := 0
	for _, t := range tokens {
		if t.kind == tokLPAREN {
			open++
		}
		if open > MaxNesting {
			return nil, syntaxError(msgL03(t.position))
		}
		if t.kind == tokRPAREN {
			open--
		}
	}
	// Step 5: the kind.
	readSet := hasKind(allowed, KindSelect) || hasKind(allowed, KindDescribe)
	var kind StatementKind
	if len(tokens) > 0 {
		switch tokens[0].kind {
		case tokSELECT, tokFROM, tokSTAR:
			kind = KindSelect
		case tokWORD:
			if len(tokens) > 1 && tokens[1].kind == tokLPAREN {
				kind = KindSelect
			}
		case tokDESCRIBE:
			kind = KindDescribe
		case tokINSERT:
			kind = KindInsert
		case tokUPDATE:
			kind = KindUpdate
		case tokDELETE:
			kind = KindDelete
		}
	}
	if kind == "" {
		if readSet {
			return nil, syntaxError(m.msgS04())
		}
		return nil, syntaxError(m.msgW02())
	}
	// Step 6: the kind must be allowed.
	if !hasKind(allowed, kind) {
		if kind == KindInsert || kind == KindUpdate || kind == KindDelete {
			return nil, syntaxError(msgS07())
		}
		return nil, syntaxError(msgW01())
	}
	// Step 7: function names.
	if kind == KindSelect {
		for i := 0; i+1 < len(tokens); i++ {
			if tokens[i].kind == tokWORD && tokens[i+1].kind == tokLPAREN && !functionNames[capitals(tokens[i].text)] {
				return nil, syntaxError(m.msgS12(tokens[i].text))
			}
		}
	}
	// Step 8: the grammar.
	root, failing := runParser(tokens)
	if root == nil {
		return nil, classifyFailure(tokens, kind, failing, m)
	}
	// Step 9: check-forms, then the reading rules.
	if finding := checkForms(tokens, kind, root, m); finding != nil {
		return nil, finding
	}
	b := builder{tokens: tokens}
	return b.statement(root), nil
}

// runParser gives the tokens to a new parser. It returns the statement,
// or nil and the index of the failing token; len(tokens) stands for END.
func runParser(tokens []token) (*stmtNode, int) {
	st := &parseState{}
	p := aqlParseAlloc()
	defer aqlParseFinalize(p)
	for i, t := range tokens {
		aqlParse(p, t.kind, i, st)
		if st.refused {
			return nil, i
		}
	}
	aqlParse(p, 0, len(tokens), st)
	if st.refused || !st.accepted || st.root == nil {
		return nil, len(tokens)
	}
	return st.root, -1
}

// builder applies the reading rules.
type builder struct {
	tokens []token
}

func (b *builder) word(i int) Word { return Word(b.tokens[i].text) }

func (b *builder) wordPtr(i int) *Word {
	if i < 0 {
		return nil
	}
	w := b.word(i)
	return &w
}

func (b *builder) value(i int) Value {
	t := b.tokens[i]
	switch t.kind {
	case tokTEXT:
		return Text(t.value)
	case tokNUMBER:
		return Number(t.value)
	case tokTRUE:
		return Boolean(true)
	case tokFALSE:
		return Boolean(false)
	}
	return Null{}
}

func (b *builder) statement(s *stmtNode) Statement {
	switch s.kind {
	case KindDescribe:
		return DescribeStatement{Entity: b.wordPtr(s.entity)}
	case KindInsert:
		ins := InsertStatement{Entity: b.word(s.entity)}
		for _, c := range s.columns {
			ins.Columns = append(ins.Columns, b.word(c))
		}
		for _, v := range s.values {
			ins.Values = append(ins.Values, b.value(v))
		}
		return ins
	case KindUpdate:
		up := UpdateStatement{Entity: b.word(s.entity), Guards: b.guards(s.where)}
		for _, a := range s.assigns {
			up.Assignments = append(up.Assignments, Assignment{Field: b.word(a.field), Value: b.value(a.value)})
		}
		return up
	case KindDelete:
		return DeleteStatement{Entity: b.word(s.entity), Guards: b.guards(s.where)}
	}
	return b.read(s.read)
}

// guards gives one guard for each comparison of the condition, in their
// order. check-forms made sure that each has the form WORD = value.
func (b *builder) guards(c *condNode) []Assignment {
	if c == nil {
		return nil
	}
	if c.op == condCmp {
		if c.cmp.operand.word < 0 || len(c.cmp.values) != 1 {
			return nil
		}
		return []Assignment{{Field: b.word(c.cmp.operand.word), Value: b.value(c.cmp.values[0])}}
	}
	var out []Assignment
	for _, k := range c.kids {
		out = append(out, b.guards(k)...)
	}
	return out
}

func (b *builder) read(r *readNode) ReadStatement {
	out := ReadStatement{Distinct: r.distinct}
	if r.items == nil {
		out.Items = []SelectItem{Star{}}
	}
	for _, it := range r.items {
		out.Items = append(out.Items, b.item(it))
	}
	out.Source = b.source(r.from.src)
	if r.from.join >= 0 {
		out.Join = &Join{Source: b.source(r.from.other), Left: b.word(r.from.left), Right: b.word(r.from.right)}
	}
	t := r.tail
	if t.where != nil {
		out.Where = b.condition(t.where)
	}
	for _, g := range t.group {
		out.GroupBy = append(out.GroupBy, b.word(g))
	}
	if t.having != nil {
		out.Having = b.condition(t.having)
	}
	for _, k := range t.order {
		var key SortKey
		switch k.kind {
		case keyWord:
			key = b.word(k.tok)
		case keyText:
			key = QuotedLabel{Label: Text(b.tokens[k.tok].value)}
		default:
			key = b.aggregate(k.call)
		}
		out.OrderBy = append(out.OrderBy, OrderKey{Key: key, Descending: k.desc})
	}
	if t.limit >= 0 {
		n, _ := strconv.ParseUint(b.tokens[t.limit].value, 10, 32)
		limit := uint32(n)
		out.Limit = &limit
	}
	return out
}

func (b *builder) item(it *itemNode) SelectItem {
	switch it.kind {
	case itemStar:
		return Star{}
	case itemQStar:
		text := b.tokens[it.tok].text
		return QualifiedStar{Qualifier: Word(text[:len(text)-2])}
	}
	var item OperandItem
	if it.kind == itemWord {
		item.Operand = b.word(it.tok)
	} else {
		item.Operand = b.aggregate(it.call)
	}
	if it.label >= 0 {
		l := b.tokens[it.label]
		label := l.text
		if l.kind == tokTEXT {
			label = l.value
		}
		item.Label = &label
	}
	return item
}

func (b *builder) aggregate(c *callNode) Aggregate {
	fn := []byte(capitals(b.tokens[c.fn].text))
	for i, x := range fn {
		if x >= 'A' && x <= 'Z' {
			fn[i] = x - 'A' + 'a'
		}
	}
	a := Aggregate{Function: AggregateFunction(fn), Distinct: c.distinct}
	if c.argKind == argWord {
		a.Argument = b.wordPtr(c.arg)
	}
	return a
}

func (b *builder) source(s *sourceNode) Source {
	return Source{Entity: b.word(s.entity), Alias: b.wordPtr(s.alias)}
}

func (b *builder) operand(o *operandNode) Operand {
	if o.call != nil {
		return b.aggregate(o.call)
	}
	return b.word(o.word)
}

func (b *builder) condition(c *condNode) Condition {
	switch c.op {
	case condOr, condAnd:
		members := make([]Condition, 0, len(c.kids))
		for _, k := range c.kids {
			members = append(members, b.condition(k))
		}
		if c.op == condOr {
			return AnyOf{Members: members}
		}
		return AllOf{Members: members}
	case condNot:
		return Negation{Member: b.condition(c.kids[0])}
	}
	cond := b.comparison(c.cmp)
	if c.cmp.notTok >= 0 {
		return Negation{Member: cond}
	}
	return cond
}

func (b *builder) comparison(c *cmpNode) Condition {
	operand := b.operand(c.operand)
	isNull := func(i int) bool { return b.tokens[i].kind == tokNULL }
	switch c.form {
	case formEq:
		if isNull(c.values[0]) {
			return NullTest{Operand: operand}
		}
		return Comparison{Operand: operand, Operator: OpEq, Value: b.value(c.values[0])}
	case formCmp:
		op := Operator(b.tokens[c.opTok].text)
		if op == OpNe && isNull(c.values[0]) {
			return NullTest{Operand: operand, Negated: true}
		}
		return Comparison{Operand: operand, Operator: op, Value: b.value(c.values[0])}
	case formIsNull:
		return NullTest{Operand: operand}
	case formIsNotNull:
		return NullTest{Operand: operand, Negated: true}
	case formInList:
		values := make([]Value, 0, len(c.values))
		for _, v := range c.values {
			values = append(values, b.value(v))
		}
		return InList{Operand: operand, Values: values}
	case formInQuery:
		return InQuery{Operand: operand, Query: b.read(c.sub)}
	case formContains:
		return Comparison{Operand: operand, Operator: OpContains, Value: b.value(c.values[0])}
	case formLike:
		return Comparison{Operand: operand, Operator: OpLike, Value: b.value(c.values[0])}
	case formBetween:
		return AllOf{Members: []Condition{
			Comparison{Operand: operand, Operator: OpGe, Value: b.value(c.values[0])},
			Comparison{Operand: operand, Operator: OpLe, Value: b.value(c.values[1])},
		}}
	}
	return Comparison{Operand: operand, Operator: OpMatch, Value: b.value(c.values[0])}
}
