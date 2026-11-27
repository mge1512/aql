// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f

package aql

// BEHAVIOR tokenize: splits the text of a statement into tokens. The
// functions lexStrict and lexLenient come from re2c (lexer.re, tokens.re)
// and decide which row of the token table fits. This file only acts by
// the kind of the row and finds the values of a token.

import (
	"strings"
	"unicode/utf8"
)

// token is a Token of the specification. kind is the token number of the
// generated parser; END is the number 0 and is never in a token list.
type token struct {
	kind     int
	text     string // the run; for CMP the operator, with <> as !=
	value    string // the content of a TEXT, the run of a NUMBER
	position int    // position of the first character, counted from 1
}

// keywords maps a Keyword to its token number.
var keywords = map[string]int{
	"SELECT": tokSELECT, "DISTINCT": tokDISTINCT, "FROM": tokFROM, "AS": tokAS,
	"JOIN": tokJOIN, "INNER": tokINNER, "LEFT": tokLEFT, "RIGHT": tokRIGHT,
	"FULL": tokFULL, "OUTER": tokOUTER, "ON": tokON, "WHERE": tokWHERE,
	"AND": tokAND, "OR": tokOR, "NOT": tokNOT, "IS": tokIS, "NULL": tokNULL,
	"IN": tokIN, "CONTAINS": tokCONTAINS, "LIKE": tokLIKE, "BETWEEN": tokBETWEEN,
	"MATCH": tokMATCH, "TRUE": tokTRUE, "FALSE": tokFALSE, "GROUP": tokGROUP,
	"BY": tokBY, "HAVING": tokHAVING, "ORDER": tokORDER, "ASC": tokASC,
	"DESC": tokDESC, "LIMIT": tokLIMIT, "OFFSET": tokOFFSET, "DESCRIBE": tokDESCRIBE,
	"INSERT": tokINSERT, "INTO": tokINTO, "VALUES": tokVALUES, "UPDATE": tokUPDATE,
	"SET": tokSET, "DELETE": tokDELETE,
}

// rowKinds maps the kind of a row to the token number.
var rowKinds = map[int]int{
	rowTEXT: tokTEXT, rowNUMBER: tokNUMBER, rowQSTAR: tokQSTAR, rowWORD: tokWORD,
	rowCMP: tokCMP, rowEQ: tokEQ, rowLPAREN: tokLPAREN, rowRPAREN: tokRPAREN,
	rowCOMMA: tokCOMMA, rowSEMI: tokSEMI, rowSTAR: tokSTAR,
}

// tokenize runs the steps of the BEHAVIOR. The text has at most
// MaxStatementLength characters.
func tokenize(text string, lenient bool, m *messages) ([]token, *AQLError) {
	var tokens []token
	lex := lexStrict
	if lenient {
		lex = lexLenient
	}
	place, position := 0, 1
	// Step 1: repeat until no character is left.
	for {
		// Step 2: the row of this place.
		row, end := lex(text, place)
		if row == rowEnd {
			break
		}
		if end <= place {
			row, end = rowNone, place+1
		}
		run := text[place:end]
		// Step 3: act by the kind of the row.
		switch row {
		case rowNone:
			c, _ := utf8.DecodeRuneInString(text[place:])
			return nil, syntaxError(m.msgS02(string(c), position))
		case rowSkip:
		case rowOpen:
			c, _ := utf8.DecodeRuneInString(run)
			return nil, syntaxError(m.msgS01(c, position))
		case rowBad:
			c, _ := utf8.DecodeRuneInString(run)
			return nil, syntaxError(m.msgS02(string(c), position))
		default:
			tokens = append(tokens, newToken(row, run, position))
			// Step 4: the token limit.
			if len(tokens) > MaxTokens {
				return nil, syntaxError(msgL02())
			}
		}
		position += utf8.RuneCountInString(run)
		place = end
	}
	return tokens, nil
}

// newToken gives a run its kind and values by the rules for tokens.
func newToken(row int, run string, position int) token {
	t := token{kind: rowKinds[row], text: run, position: position}
	switch row {
	case rowTEXT:
		t.value = textContent(run)
	case rowNUMBER:
		t.value = run
	case rowWORD:
		if k, ok := keywords[capitals(run)]; ok {
			t.kind = k
		}
	case rowCMP:
		if run == "<>" {
			t.text = "!="
		}
	}
	return t
}

// textContent is the content of a TEXT. A date without quotes is its run.
func textContent(run string) string {
	if run == "" || (run[0] != '\'' && run[0] != '"') {
		return run
	}
	quote := rune(run[0])
	inner := []rune(run[1 : len(run)-1])
	var b strings.Builder
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		switch {
		case c == '\\' && i+1 < len(inner):
			i++
			switch inner[i] {
			case 'n':
				b.WriteRune('\n')
			case 't':
				b.WriteRune('\t')
			default:
				b.WriteRune(inner[i])
			}
		case c == quote && i+1 < len(inner) && inner[i+1] == quote:
			i++
			b.WriteRune(quote)
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}
