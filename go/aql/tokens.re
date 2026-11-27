// classes of characters
letter     = L;
digit      = [0-9];
space      = [\t\n\v\f\r\x20\u0085] | Z;
any        = [^];
apostrophe = ['];
quote      = ["];
backslash  = [\\];

// one piece of the content of a text
part_1       = [^'\\] | apostrophe apostrophe | backslash any;
part_2       = [^"\\] | quote quote | backslash any;
part_lenient = (letter | digit) (apostrophe letter)+;

// patterns of the rows
spaces   = space+;
text_1   = apostrophe part_1* apostrophe;
text_2   = quote part_2* quote;
date     = digit{4} "-" digit{2} "-" digit{2};
number   = "-"? digit+ ("." digit+)?;
qstar    = (letter | "_") (letter | digit | "_")* ".*";
word     = (letter | "_") (letter | digit | "_" | ".")*;
compare  = "!=" | "<>" | "<=" | ">=" | "<" | ">";
equals   = "=";
lparen   = "(";
rparen   = ")";
comma    = ",";
semi     = ";";
star     = "*";
open_1   = apostrophe part_1* backslash?;
open_2   = quote part_2* backslash?;
bad_date = date (letter | digit | "_" | "-")+;
bad_num  = number (letter | digit | "_" | ".")+;

// lenient patterns of the rows text_1 and open_1
text_1_lenient = apostrophe (part_lenient | part_1)* apostrophe;
open_1_lenient = apostrophe (part_lenient | part_1)* backslash?;
