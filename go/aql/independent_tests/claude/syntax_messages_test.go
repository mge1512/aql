// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
// tests by: claude

package independent

import (
	"strings"
	"testing"
)

// EXAMPLES of the bundle syntax-messages.

func TestLongWordIsCut(t *testing.T) {
	expectError(t, readSet, `SELECT * FROM tickets WHERE status = abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwx`,
		"text goes in single quotes: write 'abcdefghijklmnopqrstuvwxyzabcdefghijklmn...' instead of abcdefghijklmnopqrstuvwxyzabcdefghijklmn...")
}

func TestApostropheHintInAMessage(t *testing.T) {
	expectError(t, writeSet, `UPDATE tickets SET name = 'Rider's cup' status = 'open' WHERE id = 'T-17'`,
		"the text that starts with the apostrophe at position 73 is not closed. The text contains an apostrophe (Rider's). Inside text, write an apostrophe twice: Rider''s")
}

// Further rules of syntax-messages.

// A word of exactly MaxEcho characters is not cut; the limit counts
// characters, not bytes.
func TestWordAtMaxEchoIsNotCut(t *testing.T) {
	w := strings.Repeat("ä", 40)
	expectError(t, readSet, "SELECT * FROM tickets WHERE status = "+w,
		"text goes in single quotes: write '"+w+"' instead of "+w)
	w41 := strings.Repeat("ä", 41)
	expectError(t, readSet, "SELECT * FROM tickets WHERE status = "+w41,
		"text goes in single quotes: write '"+w+"...' instead of "+w+"...")
}

// <NAME> of S12 is cut and in capital letters. The cut word ends with
// three periods, and the period of the text of S12 follows it.
func TestLongUnknownFunctionIsCut(t *testing.T) {
	name := strings.Repeat("f", 45)
	expectError(t, readSet, "SELECT "+name+"(x) FROM tickets",
		"unknown function "+strings.Repeat("F", 40)+".... Functions: COUNT, SUM, AVG, MIN, MAX. Example: SELECT COUNT(*) FROM tickets")
}

// <where> shows at most ExcerptLength characters, without white space at
// its two ends.
func TestExcerptIsCutAndTrimmed(t *testing.T) {
	// O is WHERE at position 37; the next 24 characters are
	// "WHERE status = 'open'   " and the white space at the end goes.
	expectError(t, readSet, `SELECT * FROM tickets ORDER BY name WHERE status = 'open'   AND x = 1`,
		"cannot read the statement at position 37 (WHERE status = 'open'): expected the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order"+s03Tail)
	// A longer rest is cut after 24 characters.
	expectError(t, readSet, `SELECT * FROM tickets ORDER BY name WHERE priority = 1 AND status = 'open'`,
		"cannot read the statement at position 37 (WHERE priority = 1 AND s): expected the end of the statement, or JOIN, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT in this order"+s03Tail)
}

// The apostrophe hint follows S02 and S04 as well, and the run loses the
// characters ' " % ( ) , ; at its two ends.
func TestApostropheHintAfterOtherMessages(t *testing.T) {
	expectError(t, readSet, `SELECT [x] FROM tickets WHERE n = 'it's'`,
		"unexpected character [ at position 8. The text contains an apostrophe (it's). Inside text, write an apostrophe twice: it''s")
	expectError(t, readSet, `tickets x ("it's")`,
		"a statement starts with SELECT or DESCRIBE. The text contains an apostrophe (it's). Inside text, write an apostrophe twice: it''s Form: "+readForm+". Example: "+example)
}

// A run of more than MaxEcho characters gives no apostrophe hint.
func TestLongRunGivesNoHint(t *testing.T) {
	run := strings.Repeat("a", 40) + "'s"
	expectError(t, readSet, "tickets \""+run+"\"",
		"a statement starts with SELECT or DESCRIBE. Form: "+readForm+". Example: "+example)
}

// A doubled apostrophe stands between two apostrophes, not between two
// letters, and gives no hint.
func TestDoubledApostropheGivesNoHint(t *testing.T) {
	expectError(t, readSet, `tickets 'it''s'`,
		"a statement starts with SELECT or DESCRIBE. Form: "+readForm+". Example: "+example)
}
