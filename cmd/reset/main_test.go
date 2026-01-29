package main

import (
	"go/ast"
	"testing"
)

type fakeExpr struct{ ast.Expr }

type resetFieldCodeTestCase struct {
	name   string
	parent string
	fname  string
	expr   ast.Expr
	expect string
}

func Test_resetFieldCode(t *testing.T) {
	tests := []resetFieldCodeTestCase{
		{
			name:   "int field",
			parent: "s",
			fname:  "i",
			expr:   &ast.Ident{Name: "int"},
			expect: "\ts.i = 0\n",
		},
		{
			name:   "string field",
			parent: "s",
			fname:  "str",
			expr:   &ast.Ident{Name: "string"},
			expect: "\ts.str = \"\"\n",
		},
		{
			name:   "bool field",
			parent: "s",
			fname:  "b",
			expr:   &ast.Ident{Name: "bool"},
			expect: "\ts.b = false\n",
		},
		{
			name:   "slice field",
			parent: "s",
			fname:  "arr",
			expr:   &ast.ArrayType{Elt: &ast.Ident{Name: "int"}},
			expect: "\ts.arr = s.arr[:0]\n",
		},
		{
			name:   "map field",
			parent: "s",
			fname:  "m",
			expr:   &ast.MapType{},
			expect: "\tclear(s.m)\n",
		},
		{
			name:   "pointer to int",
			parent: "s",
			fname:  "pi",
			expr:   &ast.StarExpr{X: &ast.Ident{Name: "int"}},
			expect: "\tif s.pi != nil { \t*s.pi = 0\n }\n",
		},
		{
			name:   "pointer to struct with Reset",
			parent: "s",
			fname:  "child",
			expr:   &ast.StarExpr{X: &ast.Ident{Name: "ResetableStruct"}},
			expect: "\tif s.child != nil { \tif resetter, ok := any(*s.child).(interface{ Reset() }); ok && *s.child != nil { resetter.Reset() }\n }\n",
		},
		{
			name:   "struct with Reset",
			parent: "s",
			fname:  "child",
			expr:   &ast.Ident{Name: "ResetableStruct"},
			expect: "\tif resetter, ok := any(s.child).(interface{ Reset() }); ok && s.child != nil { resetter.Reset() }\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resetFieldCode(tt.parent, tt.fname, tt.expr)
			if got != tt.expect {
				t.Errorf("resetFieldCode() = %q, want %q", got, tt.expect)
			}
		})
	}
}
