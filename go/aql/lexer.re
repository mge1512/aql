// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f

package aql

// This file is the input of the lexer generator re2c. It has the two
// tokenizers of the BEHAVIOR tokenize: one rule for each row of the token
// table, in the order of the table. The patterns come from tokens.re,
// which is the definitions block of the specification.

/*!include:re2c "unicode_categories.re" */

// The kinds of the rows. A rule returns the kind of its row.
const (
	rowNone = iota // no row fits
	rowEnd         // no character is left
	rowSkip
	rowTEXT
	rowNUMBER
	rowQSTAR
	rowWORD
	rowCMP
	rowEQ
	rowLPAREN
	rowRPAREN
	rowCOMMA
	rowSEMI
	rowSTAR
	rowOpen
	rowBad
)

func peek(s string, cur, lim int) byte {
	if cur < lim {
		return s[cur]
	}
	return 0
}

// lexStrict finds the row for the place start, with the patterns of the
// strict reading. It returns the kind of the row and the place after the
// run.
func lexStrict(s string, start int) (int, int) {
	lim := len(s)
	cur := start
	mar := start
	_ = mar
	/*!re2c
	re2c:define:YYCTYPE = byte;
	re2c:define:YYPEEK = "peek(s, cur, lim)";
	re2c:define:YYSKIP = "cur += 1";
	re2c:define:YYBACKUP = "mar = cur";
	re2c:define:YYRESTORE = "cur = mar";
	re2c:define:YYLESSTHAN = "lim <= cur";
	re2c:yyfill:enable = 0;
	re2c:eof = 0;
	!include "tokens.re";

	spaces           { return rowSkip, cur }
	text_1           { return rowTEXT, cur }
	text_2           { return rowTEXT, cur }
	date             { return rowTEXT, cur }
	number           { return rowNUMBER, cur }
	qstar            { return rowQSTAR, cur }
	word             { return rowWORD, cur }
	compare          { return rowCMP, cur }
	equals           { return rowEQ, cur }
	lparen           { return rowLPAREN, cur }
	rparen           { return rowRPAREN, cur }
	comma            { return rowCOMMA, cur }
	semi             { return rowSEMI, cur }
	star             { return rowSTAR, cur }
	open_1           { return rowOpen, cur }
	open_2           { return rowOpen, cur }
	bad_date         { return rowBad, cur }
	bad_num          { return rowBad, cur }
	$                { return rowEnd, cur }
	any              { return rowNone, cur }
	*                { return rowNone, cur }
	*/
}

// lexLenient is lexStrict with the lenient patterns of the rows text_1
// and open_1, for the second reading.
func lexLenient(s string, start int) (int, int) {
	lim := len(s)
	cur := start
	mar := start
	_ = mar
	/*!re2c
	re2c:define:YYCTYPE = byte;
	re2c:define:YYPEEK = "peek(s, cur, lim)";
	re2c:define:YYSKIP = "cur += 1";
	re2c:define:YYBACKUP = "mar = cur";
	re2c:define:YYRESTORE = "cur = mar";
	re2c:define:YYLESSTHAN = "lim <= cur";
	re2c:yyfill:enable = 0;
	re2c:eof = 0;

	spaces           { return rowSkip, cur }
	text_1_lenient   { return rowTEXT, cur }
	text_2           { return rowTEXT, cur }
	date             { return rowTEXT, cur }
	number           { return rowNUMBER, cur }
	qstar            { return rowQSTAR, cur }
	word             { return rowWORD, cur }
	compare          { return rowCMP, cur }
	equals           { return rowEQ, cur }
	lparen           { return rowLPAREN, cur }
	rparen           { return rowRPAREN, cur }
	comma            { return rowCOMMA, cur }
	semi             { return rowSEMI, cur }
	star             { return rowSTAR, cur }
	open_1_lenient   { return rowOpen, cur }
	open_2           { return rowOpen, cur }
	bad_date         { return rowBad, cur }
	bad_num          { return rowBad, cur }
	$                { return rowEnd, cur }
	any              { return rowNone, cur }
	*                { return rowNone, cur }
	*/
}
