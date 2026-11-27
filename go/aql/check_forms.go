// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f

package aql

// BEHAVIOR check-forms: finds the forms that AQL does not have in a
// statement that the grammar accepts.

import (
	"strings"
	"unicode/utf8"
)

// The rows of the table, in their order.
const (
	rowC01 = iota + 1
	rowC02
	rowC03
	rowC04
	rowC05
	rowC06
	rowC07
	rowC08
	rowC09
	rowC10
	rowC11
	rowC12
	rowC13
)

type finding struct {
	tok     int // the token at which the finding shows
	row     int
	message string
}

type formChecker struct {
	tokens   []token
	kind     StatementKind
	m        *messages
	findings []finding
}

func (c *formChecker) add(tok, row int, message string) {
	c.findings = append(c.findings, finding{tok: tok, row: row, message: message})
}

// s03 is the message S03 at the token tok with the text what.
func (c *formChecker) s03(tok int, what string) string {
	return c.m.msgS03(c.m.where(c.tokens, tok), what)
}

func (c *formChecker) isNull(i int) bool { return c.tokens[i].kind == tokNULL }

// checkForms runs the steps of the BEHAVIOR. The result is nil when there
// is no finding.
func checkForms(tokens []token, kind StatementKind, s *stmtNode, m *messages) *AQLError {
	c := &formChecker{tokens: tokens, kind: kind, m: m}
	// Step 1: collect the findings.
	switch s.kind {
	case KindSelect:
		c.read(s.read, 0)
	case KindInsert:
		for _, v := range s.values {
			if c.isNull(v) {
				c.add(v, rowC10, msgW11())
			}
		}
		if len(s.columns) != len(s.values) {
			c.add(s.valuesTok, rowC11, msgW08(len(s.columns), len(s.values)))
		}
	case KindUpdate, KindDelete:
		for _, a := range s.assigns {
			if c.isNull(a.value) {
				c.add(a.value, rowC10, msgW11())
			}
		}
		c.condition(s.where, false, 0)
		c.guards()
	}
	// Step 2: no finding.
	if len(c.findings) == 0 {
		return nil
	}
	// Step 3: the finding whose token comes first; at one token, the
	// first row of the table.
	best := c.findings[0]
	for _, f := range c.findings[1:] {
		if f.tok < best.tok || (f.tok == best.tok && f.row < best.row) {
			best = f
		}
	}
	return syntaxError(best.message)
}

// read collects the findings of a read statement; depth is 0 for the
// statement and 1 or more for a subquery.
func (c *formChecker) read(r *readNode, depth int) {
	for _, it := range r.items {
		if it.call != nil {
			c.call(it.call, false)
		}
	}
	t := r.tail
	if t.where != nil {
		c.condition(t.where, true, depth)
	}
	if t.having != nil {
		c.condition(t.having, false, depth)
	}
	for _, k := range t.order {
		if k.call != nil {
			c.call(k.call, false)
		}
	}
	if t.limit >= 0 {
		// C07: negative, not whole, or more than 9 digits.
		run := c.tokens[t.limit].text
		digits := 0
		for _, ch := range run {
			if ch >= '0' && ch <= '9' {
				digits++
			}
		}
		if strings.HasPrefix(run, "-") || strings.Contains(run, ".") || digits > 9 {
			c.add(t.limit, rowC07, c.s03(t.limit, "a whole number after LIMIT"))
		}
	}
}

// call collects C01 to C04 for one call. inWhere is true for the operand
// of a comparison in a WHERE condition.
func (c *formChecker) call(k *callNode, inWhere bool) {
	fn := capitals(c.tokens[k.fn].text)
	other := fn == "SUM" || fn == "AVG" || fn == "MIN" || fn == "MAX"
	switch {
	case k.argKind == argStar && other:
		c.add(k.arg, rowC01, c.s03(k.arg, "a field inside "+echo(fn)+"( )"))
	case k.argKind == argNumber && fn == "COUNT" && c.tokens[k.arg].text != "1":
		c.add(k.arg, rowC02, c.s03(k.arg, "a field or * inside COUNT( )"))
	case k.argKind == argNumber && other:
		c.add(k.arg, rowC03, c.s03(k.arg, "a field inside "+echo(fn)+"( )"))
	}
	if inWhere && c.kind == KindSelect {
		c.add(k.fn, rowC04, msgS18(fn))
	}
}

// condition collects the findings of a condition.
func (c *formChecker) condition(n *condNode, inWhere bool, depth int) {
	if n.op != condCmp {
		for _, k := range n.kids {
			c.condition(k, inWhere, depth)
		}
		return
	}
	k := n.cmp
	if k.operand.call != nil {
		c.call(k.operand.call, inWhere)
	}
	switch k.form {
	case formCmp:
		if c.isNull(k.values[0]) && c.tokens[k.opTok].text != "!=" {
			c.add(k.values[0], rowC08, c.s03(k.values[0], "a value in single quotes"))
		}
	case formContains:
		if c.isNull(k.values[0]) {
			c.add(k.values[0], rowC08, c.s03(k.values[0], "a value in single quotes"))
		}
	case formLike:
		v := k.values[0]
		if c.isNull(v) {
			c.add(v, rowC08, c.s03(v, "a value in single quotes"))
		}
		if c.tokens[v].kind == tokTEXT {
			if n := utf8.RuneCountInString(c.tokens[v].value); n > MaxPatternLength {
				c.add(v, rowC09, msgL04(n))
			}
		}
	case formBetween:
		for _, v := range k.values {
			if c.isNull(v) {
				c.add(v, rowC08, c.s03(v, "a value in single quotes"))
			}
		}
	case formInQuery:
		sub := k.sub
		if depth >= 1 {
			c.add(sub.sel, rowC05, msgS19())
		}
		if len(sub.items) != 1 || sub.items[0].kind == itemStar || sub.items[0].kind == itemQStar {
			c.add(sub.sel, rowC06, msgS20())
		}
		c.read(sub, depth+1)
	}
}

// guards collects C12 and C13 from the tokens of the WHERE condition of
// an UPDATE or a DELETE: every token after the first WHERE.
func (c *formChecker) guards() {
	i := 0
	for i < len(c.tokens) && c.tokens[i].kind != tokWHERE {
		i++
	}
	for i++; i < len(c.tokens); i++ {
		switch c.tokens[i].kind {
		case tokOR, tokNOT, tokLPAREN:
			c.add(i, rowC12, msgW05())
		case tokCMP, tokIS, tokIN, tokCONTAINS, tokLIKE, tokBETWEEN, tokMATCH, tokNULL:
			c.add(i, rowC13, msgW06())
		}
	}
}
