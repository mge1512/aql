// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
// tests by: claude

package independent

import (
	"testing"
)

// EXAMPLES of the bundle classify-failure.

func TestFromIsMissing(t *testing.T) {
	expectError(t, readSet, `SELECT name WHERE status = 'open'`,
		"cannot read the statement at position 13 (WHERE status = 'open'): expected FROM and an entity"+s03Tail)
}

func TestEntityIsMissing(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM`,
		"cannot read the statement at the end of the statement: expected an entity (tickets, projects, users)"+s03Tail)
}

func TestTextWithoutQuotes(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE status = open`,
		"text goes in single quotes: write 'open' instead of open")
}

func TestIsWithoutNull(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE status IS 'open'`,
		"cannot read the statement at position 39 ('open'): expected NULL after IS"+s03Tail)
}

func TestLeftJoinIsRefused(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets t LEFT JOIN projects p ON t.project = p.id`,
		"LEFT JOIN is not supported. JOIN returns the records that have a partner in the other entity. Form: SELECT fields FROM entity a JOIN entity b ON a.field = b.field")
}

func TestSecondJoinIsRefused(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets t JOIN projects p ON t.project = p.id JOIN users u ON t.owner = u.id`,
		"a statement has one JOIN and reads two entities. For a third entity, send a second statement, or use: field IN (SELECT field FROM entity WHERE ...)")
}

func TestOffsetIsRefused(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets LIMIT 10 OFFSET 20`,
		"OFFSET is not supported. Use WHERE to narrow the result and LIMIT to shorten it.")
}

func TestLimitWithTwoNumbers(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets LIMIT 10, 20`,
		"OFFSET is not supported. Use WHERE to narrow the result and LIMIT to shorten it.")
}

func TestSubqueryAsAValue(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE priority = (SELECT MAX(priority) FROM tickets)`,
		"a subquery stands only after IN, in the form: field IN (SELECT field FROM entity WHERE ...). Send any other subquery as its own statement first.")
}

func TestClausesInAnotherOrder(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets ORDER BY name WHERE status = 'open'`,
		"cannot read the statement at position 37 (WHERE status = 'open'): expected the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order"+s03Tail)
}

func TestFieldWithoutAComparison(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE status 'open'`,
		"cannot read the statement at position 36 ('open'): expected a comparison after the field status (=, !=, <, >, CONTAINS, LIKE, IN, IS NULL)"+s03Tail)
}

func TestBetweenWithoutAnd(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE due BETWEEN '2026-01-01' , '2026-02-01'`,
		"cannot read the statement at position 54 (, '2026-02-01'): expected AND in BETWEEN"+s03Tail)
}

func TestJoinWithoutOn(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets t JOIN projects p WHERE t.project = p.id`,
		"cannot read the statement at position 41 (WHERE t.project = p.id): expected ON and the join condition, in the form: ON a.field = b.field"+s03Tail)
}

func TestParenthesisThatIsNotClosed(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE (status = 'open'`,
		"cannot read the statement at the end of the statement: expected ')'"+s03Tail)
}

func TestCountWithoutAnArgument(t *testing.T) {
	expectError(t, readSet, `SELECT COUNT() FROM tickets`,
		"cannot read the statement at position 14 () FROM tickets): expected a field or * inside COUNT( )"+s03Tail)
}

func TestSecondStatementAfterTheFirst(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets; SELECT * FROM projects`,
		"cannot read the statement at position 24 (SELECT * FROM projects): expected the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order"+s03Tail)
}

func TestDescribeWithMoreText(t *testing.T) {
	expectError(t, readSet, `DESCRIBE tickets WHERE status = 'open'`,
		"DESCRIBE tickets lists the fields of tickets and takes nothing else. To read records: "+example)
}

func TestDescribeWithAText(t *testing.T) {
	expectError(t, readSet, `DESCRIBE 'tickets'`,
		"DESCRIBE takes the name of one entity: tickets, projects, users. DESCRIBE alone lists the entities.")
}

func TestInsertWithoutInto(t *testing.T) {
	expectError(t, writeSet, `INSERT tickets (name) VALUES ('x')`,
		"expected INTO after INSERT. Forms: "+writeForms)
}

func TestInsertWithoutFieldNames(t *testing.T) {
	expectError(t, writeSet, `INSERT INTO tickets VALUES ('x')`,
		"expected ( and the field names after the entity. Forms: "+writeForms)
}

func TestInsertWithTwoRows(t *testing.T) {
	expectError(t, writeSet, `INSERT INTO tickets (name) VALUES ('x'), ('y')`,
		"one INSERT creates one record. Send one statement per record.")
}

func TestUpdateWithoutWhere(t *testing.T) {
	expectError(t, writeSet, `UPDATE tickets SET status = 'closed'`,
		"UPDATE needs WHERE id = '...' with the id of the one record to change. Forms: "+writeForms)
}

func TestUpdateWithAnExpression(t *testing.T) {
	expectError(t, writeSet, `UPDATE tickets SET priority = 2 * 3 WHERE id = 'T-17'`,
		"a field can only be set to a value, not to an expression")
}

func TestDeleteWithAnAlias(t *testing.T) {
	expectError(t, writeSet, `DELETE FROM tickets t WHERE id = 'T-17'`,
		"expected WHERE after the entity. Forms: "+writeForms)
}

func TestTextAfterAWriteStatement(t *testing.T) {
	expectError(t, writeSet, `DELETE FROM tickets WHERE id = 'T-17' LIMIT 1`,
		"unexpected text at position 39 (LIMIT 1). One statement, one record. Forms: "+writeForms)
}

func TestValueWithoutQuotesInInsert(t *testing.T) {
	expectError(t, writeSet, `INSERT INTO tickets (name) VALUES (open)`,
		"text goes in single quotes: write 'open' instead of open")
}

// Further rows of the tables.

// R01 with a NUMBER after LIMIT and a NUMBER.
func TestLimitWithTwoNumbersWithoutComma(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets LIMIT 10 20`,
		"OFFSET is not supported. Use WHERE to narrow the result and LIMIT to shorten it.")
}

// R02 with keywords in small letters gives the KEYWORD in capitals.
func TestOuterJoinInSmallLetters(t *testing.T) {
	expectError(t, readSet, `select * from tickets full outer join projects on a = b`,
		"FULL JOIN is not supported. JOIN returns the records that have a partner in the other entity. Form: SELECT fields FROM entity a JOIN entity b ON a.field = b.field")
}

// R03 with INNER as the second join.
func TestSecondInnerJoin(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets t JOIN projects p ON t.project = p.id INNER JOIN users u ON t.owner = u.id`,
		"a statement has one JOIN and reads two entities. For a third entity, send a second statement, or use: field IN (SELECT field FROM entity WHERE ...)")
}

// R04 with SELECT after an LPAREN of a primary.
func TestSubqueryInParentheses(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE (SELECT id FROM projects)`,
		"a subquery stands only after IN, in the form: field IN (SELECT field FROM entity WHERE ...). Send any other subquery as its own statement first.")
}

// R08, R09, R10, R11, R12, R13, R15.
func TestRowsForTheTokenBefore(t *testing.T) {
	expectError(t, readSet, `SELECT FROM tickets`,
		"cannot read the statement at position 8 (FROM tickets): expected a field, * or COUNT(*) after SELECT"+s03Tail)
	expectError(t, readSet, `SELECT DISTINCT FROM tickets`,
		"cannot read the statement at position 17 (FROM tickets): expected a field after DISTINCT"+s03Tail)
	expectError(t, readSet, `SELECT id AS FROM tickets`,
		"cannot read the statement at position 14 (FROM tickets): expected a name after AS"+s03Tail)
	expectError(t, readSet, `SELECT id FROM tickets GROUP status`,
		"cannot read the statement at position 30 (status): expected BY after GROUP"+s03Tail)
	expectError(t, readSet, `SELECT id FROM tickets ORDER status`,
		"cannot read the statement at position 30 (status): expected BY after ORDER"+s03Tail)
	expectError(t, readSet, `SELECT id FROM tickets LIMIT 'ten'`,
		"cannot read the statement at position 30 ('ten'): expected a whole number after LIMIT"+s03Tail)
	expectError(t, readSet, `SELECT id FROM tickets t INNER projects`,
		"cannot read the statement at position 32 (projects): expected JOIN after INNER"+s03Tail)
	expectError(t, readSet, `SELECT id FROM tickets WHERE a IN 'x'`,
		"cannot read the statement at position 35 ('x'): expected '(' after IN"+s03Tail)
	expectError(t, readSet, `SELECT id FROM tickets WHERE a IS NOT 'x'`,
		"cannot read the statement at position 39 ('x'): expected NULL after IS"+s03Tail)
}

// R17 after JOIN and an entity; R06 after JOIN.
func TestJoinRows(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets JOIN`,
		"cannot read the statement at the end of the statement: expected an entity (tickets, projects, users)"+s03Tail)
}

// R21: no WORD before the place of the comparison.
func TestComparisonAfterACall(t *testing.T) {
	expectError(t, readSet, `SELECT status FROM tickets GROUP BY status HAVING COUNT(*)`,
		"cannot read the statement at the end of the statement: expected a comparison (=, !=, <, >, CONTAINS, LIKE, IN, IS NULL)"+s03Tail)
}

// R22: a value after an operator.
func TestValueAfterAnOperator(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE a = `,
		"cannot read the statement at the end of the statement: expected a value in single quotes"+s03Tail)
}

// R24: STAR can come next after the LPAREN of SUM.
func TestSumWithoutAnArgument(t *testing.T) {
	expectError(t, readSet, `SELECT sum() FROM tickets`,
		"cannot read the statement at position 12 () FROM tickets): expected a field inside SUM( )"+s03Tail)
}

// R28: a field name after GROUP BY.
func TestFieldNameAfterGroupBy(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets GROUP BY 'x'`,
		"cannot read the statement at position 32 ('x'): expected a field name"+s03Tail)
}

// D02 for DESCRIBE with a number.
func TestDescribeWithANumber(t *testing.T) {
	expectError(t, readSet, `DESCRIBE 12`,
		"DESCRIBE takes the name of one entity: tickets, projects, users. DESCRIBE alone lists the entities.")
}

// Rows of the write table.
func TestWriteTableRows(t *testing.T) {
	expectError(t, writeSet, `DELETE FROM tickets;`,
		"DELETE needs WHERE id = '...' with the id of the one record to change. Forms: "+writeForms)
	expectError(t, writeSet, `DELETE tickets WHERE id = 'x'`,
		"expected FROM after DELETE. Forms: "+writeForms)
	expectError(t, writeSet, `DELETE FROM 'tickets' WHERE id = 'x'`,
		"expected an entity at position 13 ('tickets' WHERE id = 'x'). Forms: "+writeForms)
	expectError(t, writeSet, `UPDATE 'tickets' SET a = 1 WHERE id = 'x'`,
		"expected an entity at position 8 ('tickets' SET a = 1 WHER). Forms: "+writeForms)
	expectError(t, writeSet, `UPDATE tickets a = 1 WHERE id = 'x'`,
		"expected SET after the entity. Forms: "+writeForms)
	expectError(t, writeSet, `INSERT INTO tickets (name) VALUES 'x'`,
		"expected ( after VALUES. Forms: "+writeForms)
	expectError(t, writeSet, `INSERT INTO tickets (name) ('x')`,
		"expected VALUES at position 28 (('x')). Forms: "+writeForms)
	expectError(t, writeSet, `UPDATE tickets SET a 1 WHERE id = 'x'`,
		"expected = after the field a. Forms: "+writeForms)
	expectError(t, writeSet, `UPDATE tickets SET a = WHERE id = 'x'`,
		"expected a value in single quotes at position 24 (WHERE id = 'x'). Forms: "+writeForms)
	expectError(t, writeSet, `INSERT INTO tickets (name) VALUES ('x'`,
		"expected ) after the values at the end of the statement. Forms: "+writeForms)
	expectError(t, writeSet, `INSERT INTO tickets (name VALUES ('x')`,
		"expected ) after the field names at position 27 (VALUES ('x')). Forms: "+writeForms)
	expectError(t, writeSet, `INSERT INTO tickets (`,
		"expected a field name at the end of the statement. Forms: "+writeForms)
}
