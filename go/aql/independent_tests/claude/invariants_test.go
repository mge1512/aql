// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
// tests by: claude

package independent

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/mge1512/aql/go/aql"
)

// INVARIANTS and POSTCONDITIONS, checked over a corpus. The corpus is the
// statements of the EXAMPLES, every prefix of them, and a few mutations.

var corpusSeeds = []string{
	`SELECT DISTINCT status, COUNT(*) AS n FROM tickets WHERE priority >= 2 GROUP BY status HAVING COUNT(*) > 5 ORDER BY n DESC, status LIMIT 10`,
	`select t.* from tickets t inner join projects p on t.project = p.id where p.name <> "Apollo" order by t.name;`,
	`SELECT COUNT(1) FROM tickets WHERE estimate > -2.5 AND due >= 2026-11-26`,
	`SELECT name FROM tickets WHERE project IN (SELECT id FROM projects WHERE owner = 'kim')`,
	`SELECT id FROM tickets WHERE status NOT IN ('closed', 'done') AND NOT (priority = 1 OR priority = 2)`,
	`SELECT id FROM tickets WHERE due BETWEEN 2026-01-01 AND 2026-03-31 AND owner = NULL`,
	`SELECT COUNT(DISTINCT owner) AS owners, MAX(priority) FROM tickets`,
	`INSERT INTO tickets (project, name, priority) VALUES ('P-1', 'Export to CSV', 2)`,
	`UPDATE tickets SET status = 'closed', owner = 'kim' WHERE id = 'T-17' AND status = 'open'`,
	`DELETE FROM tickets WHERE id = 'T-17';`,
	`describe Tickets;`,
	`SELECT id FROM tickets WHERE name = 'Rider's cup'`,
	`SELECT * FROM tickets t LEFT JOIN projects p ON t.project = p.id`,
	`SELECT * FROM tickets LIMIT 10 OFFSET 20`,
	`UPDATE tickets SET name = 'Rider's cup' status = 'open' WHERE id = 'T-17'`,
	`SELECT id FROM tickets WHERE body MATCH 'budget review' AND status != 'closed'`,
	`SELECT * FROM tickets WHERE status = abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxabcdefghijklmnop`,
}

func corpus() []string {
	var out []string
	seen := map[string]bool{}
	add := func(s string) {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, s := range corpusSeeds {
		r := []rune(s)
		for i := 0; i <= len(r); i++ {
			add(string(r[:i]))
		}
		add(strings.ReplaceAll(s, "'", ""))
		add(strings.ReplaceAll(s, " ", "  "))
		add(strings.ReplaceAll(s, "(", "(("))
		add(s + " " + s)
	}
	return out
}

var excerptRe = regexp.MustCompile(`at position [0-9]+ \((.*?)\)(: expected|\. One statement|\. Forms|\.)`)

// parse-statement returns a Statement or an AQLError for every text, and
// the kind of every error is syntax. Every method is deterministic.
func TestEveryTextGivesAStatementOrAnError(t *testing.T) {
	for _, kinds := range [][]aql.StatementKind{readSet, writeSet} {
		for _, text := range corpus() {
			st, err := aql.ParseStatement(text, kinds, guide)
			if (st == nil) == (err == nil) {
				t.Fatalf("%q: want exactly one of statement and error, got %v and %v", text, st, err)
			}
			st2, err2 := aql.ParseStatement(text, kinds, guide)
			if show(st) != show(st2) || (err == nil) != (err2 == nil) || (err != nil && err.Error() != err2.Error()) {
				t.Fatalf("%q: the result is not deterministic", text)
			}
			if err == nil {
				continue
			}
			e, ok := err.(*aql.AQLError)
			if !ok {
				t.Fatalf("%q: error %T", text, err)
			}
			if e.Kind != aql.ErrorSyntax {
				t.Fatalf("%q: kind %q", text, e.Kind)
			}
			if e.Message == "" {
				t.Fatalf("%q: empty message", text)
			}
		}
	}
}

// No message shows more than ExcerptLength characters from the place of
// a failure, and no message repeats a word of more than MaxEcho
// characters.
func TestMessagesAreShort(t *testing.T) {
	for _, kinds := range [][]aql.StatementKind{readSet, writeSet} {
		for _, text := range corpus() {
			_, err := aql.ParseStatement(text, kinds, guide)
			if err == nil {
				continue
			}
			msg := err.Error()
			if m := excerptRe.FindStringSubmatch(msg); m != nil {
				if n := utf8.RuneCountInString(m[1]); n > aql.ExcerptLength {
					t.Errorf("%q: excerpt %q has %d characters", text, m[1], n)
				}
			}
			for _, f := range strings.Fields(msg) {
				if strings.Contains(f, "abcdefghijklmnopqrstuvwxyzabcdefghijklmno") {
					t.Errorf("%q: message repeats more than MaxEcho characters: %s", text, msg)
				}
			}
		}
	}
}

// Every syntax error shows a form or a statement, or says in plain words
// what to send instead.
func TestMessagesSayWhatToSend(t *testing.T) {
	// S02 and W08 show neither a form nor a statement; the report lists
	// this under the ambiguities of the specification. They are left out.
	exempt := regexp.MustCompile(`^(unexpected character .* at position [0-9]+\.$|[0-9]+ fields but [0-9]+ values$)`)
	plain := []string{"Send", "send", "Use", "use:", "Write", "write", "Leave", "Text goes", "Inside text", "Functions:", "Forms:", "Form:", "Example:", "the form", "To read records", "DESCRIBE alone", "Use single quotes", "an expression"}
	for _, kinds := range [][]aql.StatementKind{readSet, writeSet} {
		for _, text := range corpus() {
			_, err := aql.ParseStatement(text, kinds, guide)
			if err == nil {
				continue
			}
			msg := err.Error()
			if exempt.MatchString(msg) {
				continue
			}
			found := false
			for _, p := range plain {
				if strings.Contains(msg, p) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%q: message says nothing about what to send: %s", text, msg)
			}
		}
	}
}

// A text that the strict reading accepts is never read by the second
// reading: an apostrophe between letters that closes a text is kept.
func TestStrictReadingIsNeverReplaced(t *testing.T) {
	st := parseOK(t, readSet, `SELECT id FROM tickets WHERE a = 'x'AND b = 'y'`)
	r := st.(aql.ReadStatement)
	want := aql.AllOf{Members: []aql.Condition{cmpw("a", aql.OpEq, aql.Text("x")), cmpw("b", aql.OpEq, aql.Text("y"))}}
	if show(r.Where) != show(want) {
		t.Fatalf("got %s", show(r.Where))
	}
}

// The failing token does not depend on the tokens after it.
func TestFailingTokenDoesNotDependOnTheRest(t *testing.T) {
	base := `SELECT * FROM tickets ORDER BY name WHERE status = 'open'`
	_, e1 := aql.ParseStatement(base, readSet, guide)
	_, e2 := aql.ParseStatement(base+` AND priority = 2 LIMIT 5`, readSet, guide)
	if e1 == nil || e2 == nil {
		t.Fatalf("want two errors")
	}
	p1 := regexp.MustCompile(`position [0-9]+`).FindString(e1.Error())
	p2 := regexp.MustCompile(`position [0-9]+`).FindString(e2.Error())
	if p1 == "" || p1 != p2 {
		t.Fatalf("failing places differ: %q and %q", p1, p2)
	}
}

// The result has every WORD as the statement writes it.
func TestWordsKeepTheirCase(t *testing.T) {
	expectStatement(t, readSet, `SELECT Name, T.Owner FROM Tickets T WHERE Status = 'Open' GROUP BY Name`,
		aql.ReadStatement{
			Items:   []aql.SelectItem{item("Name"), item("T.Owner")},
			Source:  aql.Source{Entity: "Tickets", Alias: wp("T")},
			Where:   cmpw("Status", aql.OpEq, aql.Text("Open")),
			GroupBy: []aql.Word{"Name"},
		})
}
