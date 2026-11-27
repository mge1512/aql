#!/bin/sh
# Compares the token table and the grammar in this directory with the
# specification.
#
# Usage: sh check-spec.sh SPECIFICATION
#
# The script reads tokens.re, lexer.re and grammar.y. It needs the program
# lemongrass (variable LEMONGRASS). It ends with status 1 when a part
# differs.

set -u
[ $# -eq 1 ] && [ -r "$1" ] || { echo "usage: sh check-spec.sh SPECIFICATION" >&2; exit 2; }
spec=$1
lemongrass=${LEMONGRASS:-lemongrass}
tmp=$(mktemp -d) || exit 2
trap 'rm -rf "$tmp"' EXIT
failed=0
differ() { echo "check-spec: $1" >&2; failed=1; }

# block START: the first fenced block after the line that starts with START
block() {
  awk -v start="$1" 'BEGIN { mark = sprintf("%c%c%c", 96, 96, 96) }
    index($0, start) == 1 { found = 1 }
    found && index($0, mark) == 1 { fence++; next }
    found && fence == 1 { print }
    found && fence == 2 { exit }' "$spec"
}

# 1. The definitions of the token table: the same text.
block "Definitions. They are written" >"$tmp/definitions"
[ -s "$tmp/definitions" ] || differ "no definitions block in $spec"
cmp -s "$tmp/definitions" tokens.re ||
  differ "tokens.re is not the definitions block of the specification"

# 2. The rows: each tokenizer has one rule for each row, in the order of
#    the table. The second tokenizer has the lenient patterns.
block "Rows. Each line has" | awk '{ print $1 }' >"$tmp/rows"
[ -s "$tmp/rows" ] || differ "no rows block in $spec"
sed -e 's/^text_1$/text_1_lenient/' -e 's/^open_1$/open_1_lenient/' \
  "$tmp/rows" >"$tmp/rows-lenient"
# rules N: the names of the rules in the re2c block number N of lexer.re.
# The rule for any character that comes after the rows is left out.
rules() {
  awk -v want="$1" '/\/\*!re2c/ { n++; inside = 1; next }
    inside && /\*\// { inside = 0 }
    inside && n == want && $1 ~ /^[a-z_][a-z0-9_]*$/ && $2 == "{" && $1 != "any" { print $1 }' lexer.re
}
rules 1 >"$tmp/rules-strict"
rules 2 >"$tmp/rules-lenient"
cmp -s "$tmp/rows" "$tmp/rules-strict" ||
  differ "the rules of the first tokenizer in lexer.re are not the rows of the specification"
cmp -s "$tmp/rows-lenient" "$tmp/rules-lenient" ||
  differ "the rules of the second tokenizer in lexer.re are not the rows with the lenient patterns"

# 3. The grammar.
block "Grammar. It is written" | grep -v '^$' >"$tmp/grammar"
[ -s "$tmp/grammar" ] || differ "no grammar block in $spec"
# 3a. The rules as the generator reads them. It prints the rules that have
#     an action first, so the two lists are sorted before the comparison.
"$lemongrass" -g grammar.y >"$tmp/reprint" 2>"$tmp/reprint.err" ||
  differ "lemongrass -g grammar.y failed: $(head -1 "$tmp/reprint.err")"
grep '::=' "$tmp/reprint" | sort >"$tmp/rules-read"
sort "$tmp/grammar" >"$tmp/grammar-sorted"
cmp -s "$tmp/grammar-sorted" "$tmp/rules-read" ||
  differ "the rules that lemongrass reads from grammar.y are not the rules of the specification"
# 3b. The order of the rules in grammar.y. A rule starts at the beginning
#     of a line and ends at the first period. Labels are left out.
awk '/^[a-z_][a-z0-9_]*(\([A-Za-z0-9_]+\))? ::=/ {
    rule = substr($0, 1, index($0, "."))
    gsub(/\([A-Za-z0-9_]+\)/, "", rule)
    print rule }' grammar.y >"$tmp/rules-order"
cmp -s "$tmp/grammar" "$tmp/rules-order" ||
  differ "the rules of grammar.y are not in the order of the specification"

[ $failed -eq 0 ] && echo "check-spec: the token table and the grammar are those of the specification"
exit $failed
