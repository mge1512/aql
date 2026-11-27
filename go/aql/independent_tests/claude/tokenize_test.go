// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
// tests by: claude

package independent

import (
	"strings"
	"testing"

	"github.com/mge1512/aql/go/aql"
)

// EXAMPLES of the bundle tokenize.

func TestTextsWithQuotesInside(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT id FROM tickets WHERE name = 'Rider''s cup' AND note = "say \"hi\""`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where: aql.AllOf{Members: []aql.Condition{
				cmpw("name", aql.OpEq, aql.Text("Rider's cup")),
				cmpw("note", aql.OpEq, aql.Text(`say "hi"`)),
			}},
		})
}

func TestKeywordsInSmallLetters(t *testing.T) {
	expectStatement(t, readSet,
		`select t.* from tickets t inner join projects p on t.project = p.id where p.name <> "Apollo" order by t.name;`,
		aql.ReadStatement{
			Items:   []aql.SelectItem{aql.QualifiedStar{Qualifier: "t"}},
			Source:  aql.Source{Entity: "tickets", Alias: wp("t")},
			Join:    &aql.Join{Source: aql.Source{Entity: "projects", Alias: wp("p")}, Left: "t.project", Right: "p.id"},
			Where:   cmpw("p.name", aql.OpNe, aql.Text("Apollo")),
			OrderBy: []aql.OrderKey{{Key: aql.Word("t.name")}},
		})
}

func TestNumbersAndADateWithoutQuotes(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT COUNT(1) FROM tickets WHERE estimate > -2.5 AND due >= 2026-11-26`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{aql.OperandItem{Operand: countStar()}},
			Source: src("tickets"),
			Where: aql.AllOf{Members: []aql.Condition{
				cmpw("estimate", aql.OpGt, aql.Number("-2.5")),
				cmpw("due", aql.OpGe, aql.Text("2026-11-26")),
			}},
		})
}

func TestStatementAboveTheTokenLimit(t *testing.T) {
	text := "SELECT id FROM tickets WHERE name IN (" + strings.Repeat("'a', ", 600) + "'z')"
	expectError(t, readSet, text,
		"the statement has more than 1000 words, values and signs. Send a shorter statement, for example with fewer values in a list.")
}

func TestTextThatIsNotClosed(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE name = 'open`,
		"the text that starts with the apostrophe at position 37 is not closed. Text goes in single quotes: 'like this'.")
}

func TestCharacterWithoutARow(t *testing.T) {
	expectError(t, readSet, `SELECT [name] FROM tickets`,
		"unexpected character [ at position 8.")
}

func TestNumberFollowedByLetters(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE estimate > 12abc`,
		"unexpected character 1 at position 41.")
}

// Further error paths and boundaries of tokenize.

// S01 for a text in double quotes has its own text hint.
func TestDoubleQuotedTextThatIsNotClosed(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE name = "open`,
		"the text that starts with the double quote at position 37 is not closed. Use single quotes for text: 'like this'.")
}

// A bad_date run: a date that goes on with other characters.
func TestDateFollowedByLetters(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE due = 2026-11-26x`,
		"unexpected character 2 at position 36.")
}

// Exactly MaxTokens tokens are accepted; one more is refused.
func TestTokenLimitBoundary(t *testing.T) {
	// 5 tokens, then "'a', " gives 2 tokens each, then "'z')" gives 2.
	// IN ( is 2 more: SELECT id FROM tickets WHERE name IN ( = 8 tokens.
	// 8 + 2*495 + 2 = 1000.
	ok := "SELECT id FROM tickets WHERE name IN (" + strings.Repeat("'a', ", 495) + "'z')"
	st := parseOK(t, readSet, ok)
	r := st.(aql.ReadStatement)
	in, isIn := r.Where.(aql.InList)
	if !isIn || len(in.Values) != 496 {
		t.Fatalf("want an InList of 496 values, got %s", show(r.Where))
	}
	over := ok + ";" + ";" // 1002 tokens
	expectError(t, readSet, over,
		"the statement has more than 1000 words, values and signs. Send a shorter statement, for example with fewer values in a list.")
	one := ok[:len(ok)-1] + ", 'y')" // 1002 tokens as well
	expectError(t, readSet, one,
		"the statement has more than 1000 words, values and signs. Send a shorter statement, for example with fewer values in a list.")
}

// The token limit is checked while the tokens are read: 1001 tokens with
// a statement that would be complete is refused.
func TestTokenLimitPlusOne(t *testing.T) {
	text := "SELECT id FROM tickets WHERE " + strings.Repeat("NOT ", 993) + "id = 1"
	expectError(t, readSet, text,
		"the statement has more than 1000 words, values and signs. Send a shorter statement, for example with fewer values in a list.")
}

// A keyword is found with the letters A to Z only: the letter U+017F in
// place of the s of select gives a WORD, so the statement has no kind.
func TestKeywordOnlyWithLettersAToZ(t *testing.T) {
	expectError(t, readSet, "\u017felect id FROM tickets",
		"a statement starts with SELECT or DESCRIBE. Form: "+readForm+". Example: "+example)
}

// Escapes inside a text: \n is a line break, \t a tab, \x is x.
func TestBackslashEscapes(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT id FROM tickets WHERE note = 'a\nb\tc\xd\'e'`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where:  cmpw("note", aql.OpEq, aql.Text("a\nb\tcxd'e")),
		})
}

// A line break inside a text is part of its content; white space of
// other kinds separates tokens.
func TestWhiteSpaceAndLineBreak(t *testing.T) {
	expectStatement(t, readSet,
		"SELECT\tid\u00a0FROM\u2003tickets\r\nWHERE note = 'a\nb'",
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where:  cmpw("note", aql.OpEq, aql.Text("a\nb")),
		})
}

// A NUMBER keeps every digit.
func TestNumberKeepsEveryDigit(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT id FROM tickets WHERE estimate = 12345678901234567890.000100`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where:  cmpw("estimate", aql.OpEq, aql.Number("12345678901234567890.000100")),
		})
}

// Positions count characters, not bytes.
func TestPositionsCountCharacters(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE name = 'äöü' [`,
		"unexpected character [ at position 43.")
}

// Each CMP operator and its text.
func TestCompareOperators(t *testing.T) {
	ops := map[string]aql.Operator{"!=": aql.OpNe, "<>": aql.OpNe, "<=": aql.OpLe, ">=": aql.OpGe, "<": aql.OpLt, ">": aql.OpGt, "=": aql.OpEq}
	for text, op := range ops {
		expectStatement(t, readSet, "SELECT id FROM tickets WHERE priority "+text+" 3",
			aql.ReadStatement{
				Items:  []aql.SelectItem{item("id")},
				Source: src("tickets"),
				Where:  cmpw("priority", op, aql.Number("3")),
			})
	}
}
