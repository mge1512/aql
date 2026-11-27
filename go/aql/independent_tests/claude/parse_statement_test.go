// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
// tests by: claude

package independent

import (
	"strings"
	"testing"

	"github.com/mge1512/aql/go/aql"
)

// EXAMPLES of the bundle parse-statement.

func TestReadStatementWithEveryClause(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT DISTINCT status, COUNT(*) AS n FROM tickets WHERE priority >= 2 GROUP BY status HAVING COUNT(*) > 5 ORDER BY n DESC, status LIMIT 10`,
		aql.ReadStatement{
			Items:    []aql.SelectItem{item("status"), aql.OperandItem{Operand: countStar(), Label: sp("n")}},
			Distinct: true,
			Source:   src("tickets"),
			Where:    cmpw("priority", aql.OpGe, aql.Number("2")),
			GroupBy:  []aql.Word{"status"},
			Having:   aql.Comparison{Operand: countStar(), Operator: aql.OpGt, Value: aql.Number("5")},
			OrderBy:  []aql.OrderKey{{Key: aql.Word("n"), Descending: true}, {Key: aql.Word("status")}},
			Limit:    up(10),
		})
}

func TestSelectIsImplied(t *testing.T) {
	expectStatement(t, readSet, `FROM tickets WHERE status = 'open'`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{aql.Star{}},
			Source: src("tickets"),
			Where:  cmpw("status", aql.OpEq, aql.Text("open")),
		})
	expectStatement(t, readSet, `COUNT(*) FROM tickets`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{aql.OperandItem{Operand: countStar()}},
			Source: src("tickets"),
		})
}

func TestJoinOfTwoEntities(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT t.name, p.name AS project FROM tickets t JOIN projects p ON t.project = p.id WHERE p.status = 'active'`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("t.name"), aql.OperandItem{Operand: aql.Word("p.name"), Label: sp("project")}},
			Source: aql.Source{Entity: "tickets", Alias: wp("t")},
			Join:   &aql.Join{Source: aql.Source{Entity: "projects", Alias: wp("p")}, Left: "t.project", Right: "p.id"},
			Where:  cmpw("p.status", aql.OpEq, aql.Text("active")),
		})
}

func TestSubqueryAfterIn(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT name FROM tickets WHERE project IN (SELECT id FROM projects WHERE owner = 'kim')`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("name")},
			Source: src("tickets"),
			Where: aql.InQuery{Operand: aql.Word("project"), Query: aql.ReadStatement{
				Items:  []aql.SelectItem{item("id")},
				Source: src("projects"),
				Where:  cmpw("owner", aql.OpEq, aql.Text("kim")),
			}},
		})
}

func TestMatchCondition(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT id FROM tickets WHERE body MATCH 'budget review' AND status != 'closed'`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where: aql.AllOf{Members: []aql.Condition{
				cmpw("body", aql.OpMatch, aql.Text("budget review")),
				cmpw("status", aql.OpNe, aql.Text("closed")),
			}},
		})
}

func TestBetweenAndNullAreRewritten(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT id FROM tickets WHERE due BETWEEN 2026-01-01 AND 2026-03-31 AND owner = NULL`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where: aql.AllOf{Members: []aql.Condition{
				aql.AllOf{Members: []aql.Condition{
					cmpw("due", aql.OpGe, aql.Text("2026-01-01")),
					cmpw("due", aql.OpLe, aql.Text("2026-03-31")),
				}},
				aql.NullTest{Operand: aql.Word("owner")},
			}},
		})
}

func TestNotBeforeInAndBeforeACondition(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT id FROM tickets WHERE status NOT IN ('closed', 'done') AND NOT (priority = 1 OR priority = 2)`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where: aql.AllOf{Members: []aql.Condition{
				aql.Negation{Member: aql.InList{Operand: aql.Word("status"), Values: []aql.Value{aql.Text("closed"), aql.Text("done")}}},
				aql.Negation{Member: aql.AnyOf{Members: []aql.Condition{
					cmpw("priority", aql.OpEq, aql.Number("1")),
					cmpw("priority", aql.OpEq, aql.Number("2")),
				}}},
			}},
		})
}

func TestAggregatesWithDistinctAndLabel(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT COUNT(DISTINCT owner) AS owners, MAX(priority) FROM tickets`,
		aql.ReadStatement{
			Items: []aql.SelectItem{
				aql.OperandItem{Operand: aql.Aggregate{Function: aql.FunctionCount, Argument: wp("owner"), Distinct: true}, Label: sp("owners")},
				aql.OperandItem{Operand: aql.Aggregate{Function: aql.FunctionMax, Argument: wp("priority")}},
			},
			Source: src("tickets"),
		})
}

func TestWriteStatements(t *testing.T) {
	expectStatement(t, writeSet,
		`INSERT INTO tickets (project, name, priority) VALUES ('P-1', 'Export to CSV', 2)`,
		aql.InsertStatement{
			Entity:  "tickets",
			Columns: []aql.Word{"project", "name", "priority"},
			Values:  []aql.Value{aql.Text("P-1"), aql.Text("Export to CSV"), aql.Number("2")},
		})
	expectStatement(t, writeSet,
		`UPDATE tickets SET status = 'closed', owner = 'kim' WHERE id = 'T-17' AND status = 'open'`,
		aql.UpdateStatement{
			Entity:      "tickets",
			Assignments: []aql.Assignment{{Field: "status", Value: aql.Text("closed")}, {Field: "owner", Value: aql.Text("kim")}},
			Guards:      []aql.Assignment{{Field: "id", Value: aql.Text("T-17")}, {Field: "status", Value: aql.Text("open")}},
		})
	expectStatement(t, writeSet,
		`DELETE FROM tickets WHERE id = 'T-17';`,
		aql.DeleteStatement{
			Entity: "tickets",
			Guards: []aql.Assignment{{Field: "id", Value: aql.Text("T-17")}},
		})
}

func TestDescribeStatements(t *testing.T) {
	expectStatement(t, readSet, `DESCRIBE`, aql.DescribeStatement{})
	expectStatement(t, readSet, `describe Tickets;`, aql.DescribeStatement{Entity: wp("Tickets")})
}

func TestSecondReadingTakesTheApostrophe(t *testing.T) {
	expectStatement(t, readSet, `SELECT id FROM tickets WHERE name = 'Rider's cup'`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where:  cmpw("name", aql.OpEq, aql.Text("Rider's cup")),
		})
}

func TestStrictReadingIsKept(t *testing.T) {
	expectStatement(t, readSet, `SELECT id FROM tickets WHERE code = 'A'OR code = 'B'`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where: aql.AnyOf{Members: []aql.Condition{
				cmpw("code", aql.OpEq, aql.Text("A")),
				cmpw("code", aql.OpEq, aql.Text("B")),
			}},
		})
}

func TestStatementAtTheTokenLimit(t *testing.T) {
	text := "SELECT id FROM tickets WHERE " + strings.Repeat("NOT ", 992) + "id = 1"
	st := parseOK(t, readSet, text)
	r, ok := st.(aql.ReadStatement)
	if !ok {
		t.Fatalf("want a ReadStatement, got %s", show(st))
	}
	if show(r.Items[0]) != "id" || show(r.Source) != "{tickets}" {
		t.Fatalf("items or source wrong: %s", show(r))
	}
	c := r.Where
	for i := 0; i < 992; i++ {
		n, isNeg := c.(aql.Negation)
		if !isNeg {
			t.Fatalf("level %d: want a Negation, got %s", i, show(c))
		}
		c = n.Member
	}
	if show(c) != show(cmpw("id", aql.OpEq, aql.Number("1"))) {
		t.Fatalf("innermost member: got %s", show(c))
	}
}

func TestStatementAboveTheLengthLimit(t *testing.T) {
	text := "SELECT id FROM tickets WHERE name = '" + strings.Repeat("x", 70000) + "'"
	expectError(t, readSet, text,
		"the statement has 70038 characters, and a statement has at most 65536. Send a shorter statement.")
}

func TestStatementWithoutAKind(t *testing.T) {
	expectError(t, readSet, `tickets`,
		"a statement starts with SELECT or DESCRIBE. Form: "+readForm+". Example: "+example)
}

func TestWriteStatementWithTheReadSet(t *testing.T) {
	expectError(t, readSet, `UPDATE tickets SET status = 'closed' WHERE id = 'T-17'`,
		"query only reads. Use the execute tool for INSERT, UPDATE and DELETE.")
}

func TestReadStatementWithTheWriteSet(t *testing.T) {
	expectError(t, writeSet, `SELECT * FROM tickets`,
		"execute only changes data. Use the query tool for SELECT and DESCRIBE.")
}

func TestUnknownFunction(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE COALESCE(owner, 'x') = 'x'`,
		"unknown function COALESCE. Functions: COUNT, SUM, AVG, MIN, MAX. Example: SELECT COUNT(*) FROM tickets")
}

func TestSecondReadingFailsToo(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE name = 'Rider's`,
		"cannot read the statement at position 44 (s): expected the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order. The text contains an apostrophe (Rider's). Inside text, write an apostrophe twice: Rider''s Form: "+readForm+". Example: "+example)
}

func TestNestingAboveTheLimit(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE (((((((((((((((((a = 1)))))))))))))))))`,
		"the parentheses at position 46 are nested more than 16 levels deep. Write the statement with fewer parentheses.")
}

// Further boundaries and error paths of parse-statement.

// MaxNesting open parentheses are accepted.
func TestNestingAtTheLimit(t *testing.T) {
	st := parseOK(t, readSet, `SELECT id FROM tickets WHERE ((((((((((((((((a = 1))))))))))))))))`)
	r := st.(aql.ReadStatement)
	if show(r.Where) != show(cmpw("a", aql.OpEq, aql.Number("1"))) {
		t.Fatalf("parentheses must add nothing, got %s", show(r.Where))
	}
}

// MaxStatementLength characters are accepted; one more is refused. The
// length counts characters, not bytes.
func TestLengthLimitBoundary(t *testing.T) {
	head := "SELECT id FROM tickets WHERE name = '"
	fill := MaxFill(len([]rune(head)) + 1)
	ok := head + strings.Repeat("ä", fill) + "'"
	if n := len([]rune(ok)); n != aql.MaxStatementLength {
		t.Fatalf("fixture has %d characters", n)
	}
	parseOK(t, readSet, ok)
	over := head + strings.Repeat("ä", fill+1) + "'"
	expectError(t, readSet, over,
		"the statement has 65537 characters, and a statement has at most 65536. Send a shorter statement.")
}

// MaxFill is the number of fill characters that makes a statement of
// MaxStatementLength characters, for a fixed part of n characters.
func MaxFill(n int) int { return aql.MaxStatementLength - n }

// Text that follows a complete statement is an error for every kind.
func TestTextAfterACompleteStatement(t *testing.T) {
	expectError(t, readSet, `DESCRIBE tickets;;`,
		"DESCRIBE tickets lists the fields of tickets and takes nothing else. To read records: "+example)
	expectError(t, writeSet, `INSERT INTO tickets (name) VALUES ('x') x`,
		"unexpected text at position 41 (x). One statement, one record. Forms: "+writeForms)
	expectError(t, writeSet, `UPDATE tickets SET a = 'b' WHERE id = 'T-1'; x`,
		"unexpected text at position 46 (x). One statement, one record. Forms: "+writeForms)
}

// No token, with the read set and with the write set.
func TestEmptyStatement(t *testing.T) {
	expectError(t, readSet, ``,
		"a statement starts with SELECT or DESCRIBE. Form: "+readForm+". Example: "+example)
	expectError(t, writeSet, `   `,
		"a statement starts with INSERT, UPDATE or DELETE. Forms: "+writeForms)
}

// Function names are read without regard to case; the function of the
// Aggregate is in small letters.
func TestFunctionNamesWithoutCase(t *testing.T) {
	expectStatement(t, readSet, `select sum(estimate), Avg(estimate), mIn(due) from tickets`,
		aql.ReadStatement{
			Items: []aql.SelectItem{
				aql.OperandItem{Operand: aql.Aggregate{Function: aql.FunctionSum, Argument: wp("estimate")}},
				aql.OperandItem{Operand: aql.Aggregate{Function: aql.FunctionAvg, Argument: wp("estimate")}},
				aql.OperandItem{Operand: aql.Aggregate{Function: aql.FunctionMin, Argument: wp("due")}},
			},
			Source: src("tickets"),
		})
}

// A statement that starts with * is read as if SELECT stood before it; a
// label can be a TEXT; an ORDER BY key can be a TEXT or a call.
func TestLabelsAndKeys(t *testing.T) {
	expectStatement(t, readSet,
		`COUNT(*) AS 'how many', status FROM tickets t GROUP BY status ORDER BY 'how many' DESC, COUNT(*) ASC`,
		aql.ReadStatement{
			Items:   []aql.SelectItem{aql.OperandItem{Operand: countStar(), Label: sp("how many")}, item("status")},
			Source:  aql.Source{Entity: "tickets", Alias: wp("t")},
			GroupBy: []aql.Word{"status"},
			OrderBy: []aql.OrderKey{{Key: aql.QuotedLabel{Label: "how many"}, Descending: true}, {Key: countStar()}},
		})
	expectStatement(t, readSet, `* FROM tickets AS t`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{aql.Star{}},
			Source: aql.Source{Entity: "tickets", Alias: wp("t")},
		})
}

// IS NULL, IS NOT NULL, != NULL, NULL in a list, TRUE and FALSE.
func TestNullsAndBooleans(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT id FROM tickets WHERE a IS NULL AND b IS NOT NULL AND c != NULL AND d IN (1, NULL, 'x') AND e = true AND f = FALSE`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where: aql.AllOf{Members: []aql.Condition{
				aql.NullTest{Operand: aql.Word("a")},
				aql.NullTest{Operand: aql.Word("b"), Negated: true},
				aql.NullTest{Operand: aql.Word("c"), Negated: true},
				aql.InList{Operand: aql.Word("d"), Values: []aql.Value{aql.Number("1"), aql.Null{}, aql.Text("x")}},
				cmpw("e", aql.OpEq, aql.Boolean(true)),
				cmpw("f", aql.OpEq, aql.Boolean(false)),
			}},
		})
}

// NOT before CONTAINS, LIKE, BETWEEN and MATCH; precedence of NOT, AND, OR.
func TestNotBeforeOperatorsAndPrecedence(t *testing.T) {
	expectStatement(t, readSet,
		`SELECT id FROM tickets WHERE a NOT CONTAINS 'x' OR b NOT LIKE 'y%' AND c NOT BETWEEN 1 AND 2 OR NOT d NOT MATCH 'z'`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where: aql.AnyOf{Members: []aql.Condition{
				aql.Negation{Member: cmpw("a", aql.OpContains, aql.Text("x"))},
				aql.AllOf{Members: []aql.Condition{
					aql.Negation{Member: cmpw("b", aql.OpLike, aql.Text("y%"))},
					aql.Negation{Member: aql.AllOf{Members: []aql.Condition{
						cmpw("c", aql.OpGe, aql.Number("1")),
						cmpw("c", aql.OpLe, aql.Number("2")),
					}}},
				}},
				aql.Negation{Member: aql.Negation{Member: cmpw("d", aql.OpMatch, aql.Text("z"))}},
			}},
		})
}

// The second reading is made only when the strict reading fails.
func TestSecondReadingWithDigitBeforeApostrophe(t *testing.T) {
	expectStatement(t, readSet, `SELECT id FROM tickets WHERE name = 'the 90's style'`,
		aql.ReadStatement{
			Items:  []aql.SelectItem{item("id")},
			Source: src("tickets"),
			Where:  cmpw("name", aql.OpEq, aql.Text("the 90's style")),
		})
}

// A call with the kind select must be a known function also in ORDER BY.
func TestUnknownFunctionInOrderBy(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets ORDER BY lower(name)`,
		"unknown function LOWER. Functions: COUNT, SUM, AVG, MIN, MAX. Example: SELECT COUNT(*) FROM tickets")
}

// The guide example of the fixture is accepted (PRECONDITIONS and the
// INVARIANT on examples of messages).
func TestGuideExampleIsAccepted(t *testing.T) {
	st := parseOK(t, readSet, guide.Example)
	if _, ok := st.(aql.ReadStatement); !ok {
		t.Fatalf("want a ReadStatement, got %s", show(st))
	}
	for _, s := range []string{
		"SELECT COUNT(*) FROM tickets",
		"SELECT field, COUNT(*) FROM entity GROUP BY field HAVING COUNT(*) > 1",
		"SELECT fields FROM entity a JOIN entity b ON a.field = b.field",
	} {
		parseOK(t, readSet, s)
	}
	for _, s := range []string{
		"INSERT INTO entity (field) VALUES ('value')",
		"UPDATE entity SET field = 'value' WHERE id = 'id'",
		"DELETE FROM entity WHERE id = 'id'",
	} {
		parseOK(t, writeSet, s)
	}
}

// The statement kinds of the two sets.
func TestKindsOfTheSets(t *testing.T) {
	kinds := func(st aql.Statement) aql.StatementKind { return st.Kind() }
	if k := kinds(parseOK(t, readSet, "DESCRIBE")); k != aql.KindDescribe {
		t.Errorf("kind %q", k)
	}
	if k := kinds(parseOK(t, readSet, "SELECT * FROM tickets")); k != aql.KindSelect {
		t.Errorf("kind %q", k)
	}
	if k := kinds(parseOK(t, writeSet, "DELETE FROM tickets WHERE id = 'x'")); k != aql.KindDelete {
		t.Errorf("kind %q", k)
	}
	expectError(t, writeSet, "DESCRIBE tickets",
		"execute only changes data. Use the query tool for SELECT and DESCRIBE.")
	expectError(t, readSet, "INSERT INTO tickets (a) VALUES (1)",
		"query only reads. Use the execute tool for INSERT, UPDATE and DELETE.")
	expectError(t, readSet, "DELETE FROM tickets WHERE id = 'x'",
		"query only reads. Use the execute tool for INSERT, UPDATE and DELETE.")
}
