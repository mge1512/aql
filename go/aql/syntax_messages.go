// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f

package aql

// BEHAVIOR syntax-messages: the texts of the errors.

import (
	"strconv"
	"strings"
	"unicode"
)

const readForm = "SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]"

const writeForms = "INSERT INTO entity (field, ...) VALUES ('value', ...)" +
	" | UPDATE entity SET field = 'value', ... WHERE id = 'id'" +
	" | DELETE FROM entity WHERE id = 'id'"

// messages builds the messages for one statement.
type messages struct {
	runes []rune
	guide Guide
}

func newMessages(text string, guide Guide) *messages {
	return &messages{runes: []rune(text), guide: guide}
}

func syntaxError(message string) *AQLError {
	return &AQLError{Kind: ErrorSyntax, Message: message}
}

// echo cuts a word of the statement to MaxEcho characters (step 2).
func echo(word string) string {
	r := []rune(word)
	if len(r) <= MaxEcho {
		return word
	}
	return string(r[:MaxEcho]) + "..."
}

// capitals writes the letters a to z as A to Z (step 3).
func capitals(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 'a' + 'A'
		}
	}
	return string(b)
}

// isSpaceChar is the class space of the token table.
func isSpaceChar(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', '\u0085':
		return true
	}
	return unicode.Is(unicode.Z, r)
}

// isLetterChar is the class letter of the token table.
func isLetterChar(r rune) bool { return unicode.IsLetter(r) }

// isDigitChar is the class digit of the token table.
func isDigitChar(r rune) bool { return r >= '0' && r <= '9' }

// innerApostrophe returns the index of the first apostrophe directly
// between a letter or digit and a letter, or -1.
func innerApostrophe(r []rune) int {
	for i := 1; i+1 < len(r); i++ {
		if r[i] == '\'' && (isLetterChar(r[i-1]) || isDigitChar(r[i-1])) && isLetterChar(r[i+1]) {
			return i
		}
	}
	return -1
}

// apostropheHint is the placeholder <apostrophe hint>.
func (m *messages) apostropheHint() string {
	i := innerApostrophe(m.runes)
	if i < 0 {
		return ""
	}
	l, r := i, i
	for l > 0 && !isSpaceChar(m.runes[l-1]) {
		l--
	}
	for r+1 < len(m.runes) && !isSpaceChar(m.runes[r+1]) {
		r++
	}
	run := strings.Trim(string(m.runes[l:r+1]), "'\"%(),;")
	if len([]rune(run)) > MaxEcho {
		return ""
	}
	return " The text contains an apostrophe (" + run + "). Inside text, write an apostrophe twice: " +
		strings.ReplaceAll(run, "'", "''")
}

// where is the placeholder <where> for the token at index i; an index at
// or after the end of the tokens is END.
func (m *messages) where(tokens []token, i int) string {
	if i >= len(tokens) {
		return "the end of the statement"
	}
	start := tokens[i].position - 1
	end := start + ExcerptLength
	if end > len(m.runes) {
		end = len(m.runes)
	}
	excerpt := strings.TrimFunc(string(m.runes[start:end]), isSpaceChar)
	return "position " + strconv.Itoa(tokens[i].position) + " (" + excerpt + ")"
}

func (m *messages) entityList() string { return strings.Join(m.guide.Entities, ", ") }

func (m *messages) firstEntity() string {
	if len(m.guide.Entities) == 0 {
		return ""
	}
	return m.guide.Entities[0]
}

// Messages for the limits.

func msgL01(n int) string {
	return "the statement has " + strconv.Itoa(n) + " characters, and a statement has at most " +
		strconv.Itoa(MaxStatementLength) + ". Send a shorter statement."
}

func msgL02() string {
	return "the statement has more than " + strconv.Itoa(MaxTokens) +
		" words, values and signs. Send a shorter statement, for example with fewer values in a list."
}

func msgL03(n int) string {
	return "the parentheses at position " + strconv.Itoa(n) + " are nested more than " +
		strconv.Itoa(MaxNesting) + " levels deep. Write the statement with fewer parentheses."
}

func msgL04(n int) string {
	return "the pattern of LIKE has " + strconv.Itoa(n) + " characters, and a pattern has at most " +
		strconv.Itoa(MaxPatternLength) + ". Use a shorter pattern, or CONTAINS with a part of the text."
}

// Messages for a statement that cannot be read.

func (m *messages) msgS01(quote rune, n int) string {
	name := "apostrophe"
	hint := m.apostropheHint()
	if quote == '"' {
		name = "double quote"
	}
	if hint == "" {
		if quote == '"' {
			hint = " Use single quotes for text: 'like this'."
		} else {
			hint = " Text goes in single quotes: 'like this'."
		}
	}
	return "the text that starts with the " + name + " at position " + strconv.Itoa(n) + " is not closed." + hint
}

func (m *messages) msgS02(c string, n int) string {
	return "unexpected character " + c + " at position " + strconv.Itoa(n) + "." + m.apostropheHint()
}

// msgS03 is S03; what is the text <what> of a row, or empty.
func (m *messages) msgS03(where, what string) string {
	expected := ""
	if what != "" {
		expected = ": expected " + what
	}
	return "cannot read the statement at " + where + expected + "." + m.apostropheHint() +
		" Form: " + readForm + ". Example: " + m.guide.Example
}

func (m *messages) msgS04() string {
	return "a statement starts with SELECT or DESCRIBE." + m.apostropheHint() +
		" Form: " + readForm + ". Example: " + m.guide.Example
}

func msgS07() string {
	return "query only reads. Use the execute tool for INSERT, UPDATE and DELETE."
}

func msgS08(keyword string) string {
	return capitals(echo(keyword)) + " JOIN is not supported. JOIN returns the records that have a partner in the other entity." +
		" Form: SELECT fields FROM entity a JOIN entity b ON a.field = b.field"
}

func msgS09() string {
	return "a statement has one JOIN and reads two entities. For a third entity, send a second statement," +
		" or use: field IN (SELECT field FROM entity WHERE ...)"
}

func msgS10() string {
	return "OFFSET is not supported. Use WHERE to narrow the result and LIMIT to shorten it."
}

func msgS11() string {
	return "a subquery stands only after IN, in the form: field IN (SELECT field FROM entity WHERE ...)." +
		" Send any other subquery as its own statement first."
}

func (m *messages) msgS12(name string) string {
	return "unknown function " + capitals(echo(name)) + ". Functions: COUNT, SUM, AVG, MIN, MAX." +
		" Example: SELECT COUNT(*) FROM " + m.firstEntity()
}

func msgS13(word string) string {
	w := echo(word)
	return "text goes in single quotes: write '" + w + "' instead of " + w
}

func (m *messages) msgS14(entity string) string {
	e := echo(entity)
	return "DESCRIBE " + e + " lists the fields of " + e + " and takes nothing else. To read records: " + m.guide.Example
}

func (m *messages) msgS17() string {
	return "DESCRIBE takes the name of one entity: " + m.entityList() + ". DESCRIBE alone lists the entities."
}

func msgS18(function string) string {
	return capitals(echo(function)) + "( ) cannot stand in WHERE. Use it after GROUP BY, in the form:" +
		" SELECT field, COUNT(*) FROM entity GROUP BY field HAVING COUNT(*) > 1"
}

func msgS19() string {
	return "a subquery cannot contain another subquery. Send the inner one as its own statement first."
}

func msgS20() string {
	return "a subquery selects one field, in the form: field IN (SELECT field FROM entity WHERE ...)"
}

// Messages of write statements.

func msgW01() string {
	return "execute only changes data. Use the query tool for SELECT and DESCRIBE."
}

func (m *messages) msgW02() string {
	return "a statement starts with INSERT, UPDATE or DELETE." + m.apostropheHint() + " Forms: " + writeForms
}

func (m *messages) msgW03(problem string) string {
	return problem + m.apostropheHint() + " Forms: " + writeForms
}

func (m *messages) msgW04(keyword string) string {
	return keyword + " needs WHERE id = '...' with the id of the one record to change." + m.apostropheHint() +
		" Forms: " + writeForms
}

func msgW05() string {
	return "one statement changes one record: the WHERE must start with id = '...'. Further conditions are joined" +
		" with AND and have the form field = 'value'. To change several records, find their ids with the query" +
		" tool and send one statement per record."
}

func msgW06() string {
	return "after WHERE id = '...', further conditions must have the form AND field = 'value'." +
		" For a field without a value write: AND field = ''"
}

func msgW07() string {
	return "one INSERT creates one record. Send one statement per record."
}

func msgW08(n, m int) string {
	return strconv.Itoa(n) + " fields but " + strconv.Itoa(m) + " values"
}

func msgW10() string {
	return "a field can only be set to a value, not to an expression"
}

func msgW11() string {
	return "NULL cannot be written. Leave the field out."
}
