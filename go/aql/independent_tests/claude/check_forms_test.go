// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
// tests by: claude

package independent

import (
	"strings"
	"testing"

	"github.com/mge1512/aql/go/aql"
)

// EXAMPLES of the bundle check-forms.

func TestLikePatternAboveTheLimit(t *testing.T) {
	text := "SELECT id FROM tickets WHERE name LIKE '" + strings.Repeat("x", 201) + "'"
	expectError(t, readSet, text,
		"the pattern of LIKE has 201 characters, and a pattern has at most 200. Use a shorter pattern, or CONTAINS with a part of the text.")
}

func TestAggregateInWhere(t *testing.T) {
	expectError(t, readSet, `SELECT status FROM tickets WHERE COUNT(*) > 5 GROUP BY status`,
		"COUNT( ) cannot stand in WHERE. Use it after GROUP BY, in the form: SELECT field, COUNT(*) FROM entity GROUP BY field HAVING COUNT(*) > 1")
}

func TestSubqueryInsideASubquery(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE project IN (SELECT id FROM projects WHERE owner IN (SELECT id FROM users WHERE team = 'x'))`,
		"a subquery cannot contain another subquery. Send the inner one as its own statement first.")
}

func TestSubqueryWithTwoItems(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE project IN (SELECT id, name FROM projects)`,
		"a subquery selects one field, in the form: field IN (SELECT field FROM entity WHERE ...)")
}

func TestStarInsideSum(t *testing.T) {
	expectError(t, readSet, `SELECT SUM(*) FROM tickets`,
		"cannot read the statement at position 12 (*) FROM tickets): expected a field inside SUM( )"+s03Tail)
}

func TestLimitThatIsNotWhole(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets LIMIT 2.5`,
		"cannot read the statement at position 29 (2.5): expected a whole number after LIMIT"+s03Tail)
}

func TestNullAfterAnOrderingOperator(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE priority > NULL`,
		"cannot read the statement at position 40 (NULL): expected a value in single quotes"+s03Tail)
}

func TestNullAsAWrittenValue(t *testing.T) {
	expectError(t, writeSet, `UPDATE tickets SET owner = NULL WHERE id = 'T-17'`,
		"NULL cannot be written. Leave the field out.")
}

func TestColumnsAndValuesDiffer(t *testing.T) {
	expectError(t, writeSet, `INSERT INTO tickets (name, priority) VALUES ('x')`,
		"2 fields but 1 values")
}

func TestOrInTheWhereOfAWriteStatement(t *testing.T) {
	expectError(t, writeSet, `UPDATE tickets SET status = 'closed' WHERE id = 'T-17' OR id = 'T-18'`,
		"one statement changes one record: the WHERE must start with id = '...'. Further conditions are joined with AND and have the form field = 'value'. To change several records, find their ids with the query tool and send one statement per record.")
}

func TestOtherComparisonInAGuard(t *testing.T) {
	expectError(t, writeSet, `UPDATE tickets SET status = 'closed' WHERE id = 'T-17' AND priority > 2`,
		"after WHERE id = '...', further conditions must have the form AND field = 'value'. For a field without a value write: AND field = ''")
}

// Further rows of the table and boundaries.

const w05 = "one statement changes one record: the WHERE must start with id = '...'. Further conditions are joined with AND and have the form field = 'value'. To change several records, find their ids with the query tool and send one statement per record."
const w06 = "after WHERE id = '...', further conditions must have the form AND field = 'value'. For a field without a value write: AND field = ''"

// A LIKE pattern of MaxPatternLength characters is accepted.
func TestLikePatternAtTheLimit(t *testing.T) {
	text := "SELECT id FROM tickets WHERE name LIKE '" + strings.Repeat("ö", 200) + "'"
	parseOK(t, readSet, text)
	expectError(t, readSet, "SELECT id FROM tickets WHERE name NOT LIKE '"+strings.Repeat("ö", 201)+"'",
		"the pattern of LIKE has 201 characters, and a pattern has at most 200. Use a shorter pattern, or CONTAINS with a part of the text.")
}

// C02: COUNT with a number other than 1. C03: SUM, AVG, MIN, MAX with a
// number.
func TestNumbersInsideCalls(t *testing.T) {
	expectError(t, readSet, `SELECT COUNT(2) FROM tickets`,
		"cannot read the statement at position 14 (2) FROM tickets): expected a field or * inside COUNT( )"+s03Tail)
	expectError(t, readSet, `SELECT avg(3) FROM tickets`,
		"cannot read the statement at position 12 (3) FROM tickets): expected a field inside AVG( )"+s03Tail)
	expectError(t, readSet, `SELECT MAX(*) FROM tickets`,
		"cannot read the statement at position 12 (*) FROM tickets): expected a field inside MAX( )"+s03Tail)
}

// C07: negative LIMIT and LIMIT with more than 9 digits; 9 digits pass.
func TestLimitBoundaries(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets LIMIT -1`,
		"cannot read the statement at position 29 (-1): expected a whole number after LIMIT"+s03Tail)
	expectError(t, readSet, `SELECT * FROM tickets LIMIT 1234567890`,
		"cannot read the statement at position 29 (1234567890): expected a whole number after LIMIT"+s03Tail)
	expectStatement(t, readSet, `SELECT * FROM tickets LIMIT 999999999`,
		aql.ReadStatement{Items: []aql.SelectItem{aql.Star{}}, Source: src("tickets"), Limit: up(999999999)})
}

// C08: NULL after CONTAINS and LIKE, and as a value of BETWEEN.
func TestNullAfterOtherOperators(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE a CONTAINS NULL`,
		"cannot read the statement at position 40 (NULL): expected a value in single quotes"+s03Tail)
	expectError(t, readSet, `SELECT * FROM tickets WHERE a LIKE NULL`,
		"cannot read the statement at position 36 (NULL): expected a value in single quotes"+s03Tail)
	expectError(t, readSet, `SELECT * FROM tickets WHERE a BETWEEN 1 AND NULL`,
		"cannot read the statement at position 45 (NULL): expected a value in single quotes"+s03Tail)
}

// C04 counts the WHERE of a subquery; HAVING is no WHERE condition.
func TestCallInWhereOfASubquery(t *testing.T) {
	expectError(t, readSet, `SELECT id FROM tickets WHERE project IN (SELECT id FROM projects WHERE max(x) = 1)`,
		"MAX( ) cannot stand in WHERE. Use it after GROUP BY, in the form: SELECT field, COUNT(*) FROM entity GROUP BY field HAVING COUNT(*) > 1")
	parseOK(t, readSet, `SELECT status FROM tickets GROUP BY status HAVING COUNT(*) > 1`)
}

// C06 for the item * and t.* in a subquery.
func TestSubqueryWithAStar(t *testing.T) {
	msg := "a subquery selects one field, in the form: field IN (SELECT field FROM entity WHERE ...)"
	expectError(t, readSet, `SELECT id FROM tickets WHERE project IN (SELECT * FROM projects)`, msg)
	expectError(t, readSet, `SELECT id FROM tickets WHERE project IN (SELECT p.* FROM projects p)`, msg)
}

// The finding whose token comes first wins: C08 at an earlier NULL beats
// C04 at a later call.
func TestFirstFindingWins(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE a > NULL AND COUNT(*) > 1`,
		"cannot read the statement at position 33 (NULL AND COUNT(*) > 1): expected a value in single quotes"+s03Tail)
}

// C10 for a value of VALUES.
func TestNullInInsert(t *testing.T) {
	expectError(t, writeSet, `INSERT INTO tickets (name) VALUES (NULL)`,
		"NULL cannot be written. Leave the field out.")
}

// C11 with more values than columns.
func TestMoreValuesThanColumns(t *testing.T) {
	expectError(t, writeSet, `INSERT INTO tickets (name) VALUES ('x', 'y', 3)`,
		"1 fields but 3 values")
}

// C12 at NOT and at LPAREN; C13 at IS, IN, LIKE and NULL as a value.
func TestGuardsOfDelete(t *testing.T) {
	expectError(t, writeSet, `DELETE FROM tickets WHERE NOT id = 'x'`, w05)
	expectError(t, writeSet, `DELETE FROM tickets WHERE (id = 'x')`, w05)
	expectError(t, writeSet, `DELETE FROM tickets WHERE id IS NULL`, w06)
	expectError(t, writeSet, `DELETE FROM tickets WHERE id IN ('a', 'b')`, w06)
	expectError(t, writeSet, `DELETE FROM tickets WHERE id = 'x' AND name LIKE 'y'`, w06)
	expectError(t, writeSet, `DELETE FROM tickets WHERE id = NULL`, w06)
	expectError(t, writeSet, `UPDATE tickets SET a = 1 WHERE id = 'x' AND b = NULL`, w06)
}

// A guard may have a NUMBER or a Boolean value.
func TestGuardsWithOtherValues(t *testing.T) {
	expectStatement(t, writeSet, `UPDATE tickets SET done = TRUE, priority = 3 WHERE id = 17 AND active = false`,
		aql.UpdateStatement{
			Entity:      "tickets",
			Assignments: []aql.Assignment{{Field: "done", Value: aql.Boolean(true)}, {Field: "priority", Value: aql.Number("3")}},
			Guards:      []aql.Assignment{{Field: "id", Value: aql.Number("17")}, {Field: "active", Value: aql.Boolean(false)}},
		})
}
