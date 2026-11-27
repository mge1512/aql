// generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
// tests by: claude

// Package independent holds the black-box tests of libaql. The tests use
// the public interface of the package aql only.
package independent

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mge1512/aql/go/aql"
)

// The fixture guide of the EXAMPLES.
var guide = aql.Guide{
	Entities: []string{"tickets", "projects", "users"},
	Example:  "SELECT id, name, status FROM tickets WHERE project = 'P-1'",
}

const readForm = "SELECT fields FROM entity [JOIN entity ON a.field = b.field] [WHERE field = 'value' [AND ...]] [GROUP BY field [HAVING COUNT(*) > 1]] [ORDER BY field [DESC]] [LIMIT n]"
const example = "SELECT id, name, status FROM tickets WHERE project = 'P-1'"
const writeForms = "INSERT INTO entity (field, ...) VALUES ('value', ...) | UPDATE entity SET field = 'value', ... WHERE id = 'id' | DELETE FROM entity WHERE id = 'id'"

// tail of an S03 message without an apostrophe hint
const s03Tail = ". Form: " + readForm + ". Example: " + example

// show writes a value of the public types in one canonical form. A member
// that is none, false or an empty list is left out, as in the EXAMPLES.
func show(v any) string {
	switch x := v.(type) {
	case nil:
		return "<none>"
	case aql.ReadStatement:
		var b []string
		items := []string{}
		for _, it := range x.Items {
			items = append(items, show(it))
		}
		b = append(b, "items:["+strings.Join(items, ", ")+"]")
		if x.Distinct {
			b = append(b, "distinct:true")
		}
		b = append(b, "source:"+show(x.Source))
		if x.Join != nil {
			b = append(b, "join:"+show(*x.Join))
		}
		if x.Where != nil {
			b = append(b, "where:"+show(x.Where))
		}
		if len(x.GroupBy) > 0 {
			g := []string{}
			for _, w := range x.GroupBy {
				g = append(g, string(w))
			}
			b = append(b, "group_by:["+strings.Join(g, ", ")+"]")
		}
		if x.Having != nil {
			b = append(b, "having:"+show(x.Having))
		}
		if len(x.OrderBy) > 0 {
			o := []string{}
			for _, k := range x.OrderBy {
				o = append(o, show(k))
			}
			b = append(b, "order_by:["+strings.Join(o, ", ")+"]")
		}
		if x.Limit != nil {
			b = append(b, fmt.Sprintf("limit:%d", *x.Limit))
		}
		return "Read{" + strings.Join(b, " ") + "}"
	case aql.DescribeStatement:
		if x.Entity == nil {
			return "Describe{}"
		}
		return "Describe{" + string(*x.Entity) + "}"
	case aql.InsertStatement:
		c := []string{}
		for _, w := range x.Columns {
			c = append(c, string(w))
		}
		vs := []string{}
		for _, v := range x.Values {
			vs = append(vs, show(v))
		}
		return "Insert{" + string(x.Entity) + " columns:[" + strings.Join(c, ", ") + "] values:[" + strings.Join(vs, ", ") + "]}"
	case aql.UpdateStatement:
		return "Update{" + string(x.Entity) + " set:" + showAssignments(x.Assignments) + " guards:" + showAssignments(x.Guards) + "}"
	case aql.DeleteStatement:
		return "Delete{" + string(x.Entity) + " guards:" + showAssignments(x.Guards) + "}"
	case aql.Star:
		return "*"
	case aql.QualifiedStar:
		return "QualifiedStar{" + string(x.Qualifier) + "}"
	case aql.OperandItem:
		if x.Label == nil {
			return show(x.Operand)
		}
		return "{" + show(x.Operand) + " label:" + *x.Label + "}"
	case aql.Word:
		return string(x)
	case aql.Aggregate:
		arg := "*"
		if x.Argument != nil {
			arg = string(*x.Argument)
		}
		if x.Distinct {
			arg = "distinct " + arg
		}
		return string(x.Function) + "(" + arg + ")"
	case aql.Source:
		if x.Alias == nil {
			return "{" + string(x.Entity) + "}"
		}
		return "{" + string(x.Entity) + " alias:" + string(*x.Alias) + "}"
	case aql.Join:
		return "{" + show(x.Source) + " " + string(x.Left) + "=" + string(x.Right) + "}"
	case aql.OrderKey:
		s := show(x.Key)
		if x.Descending {
			s += " desc"
		}
		return s
	case aql.QuotedLabel:
		return "label'" + string(x.Label) + "'"
	case aql.Comparison:
		return show(x.Operand) + " " + string(x.Operator) + " " + show(x.Value)
	case aql.NullTest:
		if x.Negated {
			return show(x.Operand) + " IS NOT NULL"
		}
		return show(x.Operand) + " IS NULL"
	case aql.InList:
		vs := []string{}
		for _, v := range x.Values {
			vs = append(vs, show(v))
		}
		return "InList{" + show(x.Operand) + " [" + strings.Join(vs, ", ") + "]}"
	case aql.InQuery:
		return "InQuery{" + show(x.Operand) + " " + show(x.Query) + "}"
	case aql.AllOf:
		return "AllOf[" + showConditions(x.Members) + "]"
	case aql.AnyOf:
		return "AnyOf[" + showConditions(x.Members) + "]"
	case aql.Negation:
		return "Not{" + show(x.Member) + "}"
	case aql.Text:
		return "'" + string(x) + "'"
	case aql.Number:
		return "#" + string(x)
	case aql.Boolean:
		return fmt.Sprintf("%v", bool(x))
	case aql.Null:
		return "NULL"
	}
	return fmt.Sprintf("<unknown %T>", v)
}

func showConditions(cs []aql.Condition) string {
	s := []string{}
	for _, c := range cs {
		s = append(s, show(c))
	}
	return strings.Join(s, ", ")
}

func showAssignments(as []aql.Assignment) string {
	s := []string{}
	for _, a := range as {
		s = append(s, string(a.Field)+" = "+show(a.Value))
	}
	return "[" + strings.Join(s, ", ") + "]"
}

// small constructors for the expected values

func wp(s string) *aql.Word { w := aql.Word(s); return &w }
func sp(s string) *string   { return &s }
func up(n uint32) *uint32   { return &n }

func cmpw(field string, op aql.Operator, v aql.Value) aql.Comparison {
	return aql.Comparison{Operand: aql.Word(field), Operator: op, Value: v}
}

func item(field string) aql.OperandItem { return aql.OperandItem{Operand: aql.Word(field)} }

func src(entity string) aql.Source { return aql.Source{Entity: aql.Word(entity)} }

func countStar() aql.Aggregate { return aql.Aggregate{Function: aql.FunctionCount} }

// parse helpers

func parseOK(t *testing.T, kinds []aql.StatementKind, text string) aql.Statement {
	t.Helper()
	st, err := aql.ParseStatement(text, kinds, guide)
	if err != nil {
		t.Fatalf("ParseStatement(%q): unexpected error %v", text, err)
	}
	if st == nil {
		t.Fatalf("ParseStatement(%q): no statement and no error", text)
	}
	return st
}

func expectStatement(t *testing.T, kinds []aql.StatementKind, text string, want aql.Statement) {
	t.Helper()
	got := parseOK(t, kinds, text)
	if show(got) != show(want) {
		t.Errorf("ParseStatement(%q)\n got: %s\nwant: %s", text, show(got), show(want))
	}
}

func expectError(t *testing.T, kinds []aql.StatementKind, text string, message string) {
	t.Helper()
	st, err := aql.ParseStatement(text, kinds, guide)
	if err == nil {
		t.Fatalf("ParseStatement(%q): expected an error, got %s", text, show(st))
	}
	var e *aql.AQLError
	if !asAQLError(err, &e) {
		t.Fatalf("ParseStatement(%q): error %T is not *aql.AQLError", text, err)
	}
	if e.Kind != aql.ErrorSyntax {
		t.Errorf("ParseStatement(%q): kind %q, want syntax", text, e.Kind)
	}
	if e.Message != message {
		t.Errorf("ParseStatement(%q)\n got message: %s\nwant message: %s", text, e.Message, message)
	}
	if e.Error() != message {
		t.Errorf("Error() = %q, want the message", e.Error())
	}
}

func asAQLError(err error, target **aql.AQLError) bool {
	e, ok := err.(*aql.AQLError)
	if ok {
		*target = e
	}
	return ok
}

var readSet = aql.ReadSet()
var writeSet = aql.WriteSet()
