// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f

package aql

// BEHAVIOR classify-failure: gives the error for tokens that are not a
// statement of the grammar. What can come next is always a trial with a
// new generated parser; no other code decides what the grammar allows.

// canComeNext says whether a token of the kind can follow the tokens. The
// kind 0 is END. The parser accepts the tokens themselves.
func canComeNext(tokens []token, kind int) bool {
	st := &parseState{}
	p := aqlParseAlloc()
	defer aqlParseFinalize(p)
	for i, t := range tokens {
		aqlParse(p, t.kind, i, st)
	}
	aqlParse(p, kind, len(tokens), st)
	if kind == 0 {
		return st.accepted && !st.refused
	}
	return !st.refused
}

// failure is the place of a failure and what the trials say about it.
type failure struct {
	tokens []token
	o      int // index of the failing token O; len(tokens) is END
	cache  map[int]bool
	m      *messages
}

// kindAt is the kind of the token at index i, or -1 when there is none.
// END has the kind 0.
func (f *failure) kindAt(i int) int {
	if i < 0 {
		return -1
	}
	if i >= len(f.tokens) {
		return 0
	}
	return f.tokens[i].kind
}

func (f *failure) textAt(i int) string {
	if i < 0 || i >= len(f.tokens) {
		return ""
	}
	return f.tokens[i].text
}

// O, P and P2 (step 2).
func (f *failure) O() int  { return f.kindAt(f.o) }
func (f *failure) P() int  { return f.kindAt(f.o - 1) }
func (f *failure) P2() int { return f.kindAt(f.o - 2) }

// can is the test "K can come next" of step 3.
func (f *failure) can(kind int) bool {
	if v, ok := f.cache[kind]; ok {
		return v
	}
	v := canComeNext(f.tokens[:f.o], kind)
	f.cache[kind] = v
	return v
}

// before says whether a token of the kind stands before O.
func (f *failure) before(kind int) bool {
	for _, t := range f.tokens[:f.o] {
		if t.kind == kind {
			return true
		}
	}
	return false
}

func (f *failure) where() string { return f.m.where(f.tokens, f.o) }

// classifyFailure runs the steps of the BEHAVIOR. failing is the index of
// the failing token that the parser found (step 1); len(tokens) is END.
func classifyFailure(tokens []token, kind StatementKind, failing int, m *messages) *AQLError {
	f := &failure{tokens: tokens, o: failing, cache: map[int]bool{}, m: m}
	// Step 4: the table of the kind.
	switch kind {
	case KindSelect:
		return syntaxError(f.readTable())
	case KindDescribe:
		return syntaxError(f.describeTable())
	}
	return syntaxError(f.writeTable(kind))
}

func (f *failure) readTable() string {
	o, p, p2 := f.O(), f.P(), f.P2()
	s03 := func(what string) string { return f.m.msgS03(f.where(), what) }
	switch {
	case o == tokOFFSET || (p2 == tokLIMIT && p == tokNUMBER && (o == tokCOMMA || o == tokNUMBER)): // R01
		return msgS10()
	case (o == tokLEFT || o == tokRIGHT || o == tokFULL || o == tokOUTER) && f.can(0): // R02
		return msgS08(f.textAt(f.o))
	case (o == tokJOIN || o == tokINNER) && f.before(tokJOIN): // R03
		return msgS09()
	case (o == tokSELECT && p == tokLPAREN) || (o == tokLPAREN && f.kindAt(f.o+1) == tokSELECT && f.o+1 < len(f.tokens)): // R04
		return msgS11()
	case o == tokWORD && f.can(tokTEXT) && !f.can(tokWORD): // R05
		return msgS13(f.textAt(f.o))
	case p == tokFROM || p == tokJOIN: // R06
		return s03("an entity (" + f.m.entityList() + ")")
	case p == tokDISTINCT: // R07
		return s03("a field after DISTINCT")
	case p == tokSELECT: // R08
		return s03("a field, * or COUNT(*) after SELECT")
	case p == tokAS: // R09
		return s03("a name after AS")
	case p == tokGROUP: // R10
		return s03("BY after GROUP")
	case p == tokORDER: // R11
		return s03("BY after ORDER")
	case p == tokLIMIT: // R12
		return s03("a whole number after LIMIT")
	case p == tokINNER: // R13
		return s03("JOIN after INNER")
	case p == tokIS || (p == tokNOT && p2 == tokIS): // R14
		return s03("NULL after IS")
	case p == tokIN: // R15
		return s03("'(' after IN")
	case f.can(tokAND) && !f.can(tokOR): // R16
		return s03("AND in BETWEEN")
	case f.can(tokON): // R17
		return s03("ON and the join condition, in the form: ON a.field = b.field")
	case f.can(tokEQ) && !f.can(tokCMP): // R18
		return s03("= after the field " + echo(f.textAt(f.o-1)))
	case f.can(tokFROM): // R19
		return s03("FROM and an entity")
	case f.can(tokCONTAINS) && p == tokWORD: // R20
		return s03("a comparison after the field " + echo(f.textAt(f.o-1)) + " (=, !=, <, >, CONTAINS, LIKE, IN, IS NULL)")
	case f.can(tokCONTAINS): // R21
		return s03("a comparison (=, !=, <, >, CONTAINS, LIKE, IN, IS NULL)")
	case f.can(tokTEXT) && !f.can(tokWORD): // R22
		return s03("a value in single quotes")
	case f.can(tokSTAR) && p == tokLPAREN && p2 == tokWORD && capitals(f.textAt(f.o-2)) == "COUNT": // R23
		return s03("a field or * inside COUNT( )")
	case f.can(tokSTAR) && p == tokLPAREN: // R24
		return s03("a field inside " + capitals(echo(f.textAt(f.o-2))) + "( )")
	case f.can(tokSTAR): // R25
		return s03("a field, * or COUNT(*)")
	case f.can(tokRPAREN): // R26
		return s03("')'")
	case f.can(0): // R27
		return s03("the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order")
	case f.can(tokWORD): // R28
		return s03("a field name")
	}
	return s03("") // R29
}

func (f *failure) describeTable() string {
	if len(f.tokens) > 1 && f.tokens[1].kind == tokWORD && f.o > 1 { // D01
		return f.m.msgS14(f.tokens[1].text)
	}
	return f.m.msgS17() // D02
}

func (f *failure) writeTable(kind StatementKind) string {
	o, p, p2 := f.O(), f.P(), f.P2()
	w03 := f.m.msgW03
	switch {
	case o == tokWORD && f.can(tokTEXT) && !f.can(tokWORD): // X01
		return msgS13(f.textAt(f.o))
	case (o == 0 || o == tokSEMI) && f.can(tokWHERE): // X02
		keyword := "UPDATE"
		if kind == KindDelete {
			keyword = "DELETE"
		}
		return f.m.msgW04(keyword)
	case kind == KindUpdate && f.can(tokWHERE): // X03
		return msgW10()
	case kind == KindDelete && f.can(tokWHERE): // X04
		return w03("expected WHERE after the entity.")
	case kind == KindInsert && o == tokCOMMA && f.can(0): // X05
		return msgW07()
	case p == tokINSERT: // X06
		return w03("expected INTO after INSERT.")
	case p == tokDELETE: // X07
		return w03("expected FROM after DELETE.")
	case p == tokINTO || p == tokUPDATE || (p == tokFROM && p2 == tokDELETE): // X08
		return w03("expected an entity at " + f.where() + ".")
	case f.can(tokSET): // X09
		return w03("expected SET after the entity.")
	case kind == KindInsert && f.can(tokLPAREN) && p == tokVALUES: // X10
		return w03("expected ( after VALUES.")
	case kind == KindInsert && f.can(tokLPAREN): // X11
		return w03("expected ( and the field names after the entity.")
	case f.can(tokVALUES): // X12
		return w03("expected VALUES at " + f.where() + ".")
	case f.can(tokEQ) && p == tokWORD: // X13
		return w03("expected = after the field " + echo(f.textAt(f.o-1)) + ".")
	case f.can(tokTEXT): // X14
		return w03("expected a value in single quotes at " + f.where() + ".")
	case kind == KindInsert && f.can(tokRPAREN) && f.before(tokVALUES): // X15
		return w03("expected ) after the values at " + f.where() + ".")
	case kind == KindInsert && f.can(tokRPAREN): // X16
		return w03("expected ) after the field names at " + f.where() + ".")
	case f.can(tokWORD): // X17
		return w03("expected a field name at " + f.where() + ".")
	case f.can(0): // X18
		return w03("unexpected text at " + f.where() + ". One statement, one record.")
	}
	return w03("cannot read the statement at " + f.where() + ".") // X19
}
