package orchestrator

import (
	"testing"
)

func TestParserAST(t *testing.T) {
	tests := []struct{ input, want string }{
		{"2+2", "(2+2)"},
		{"(1+2)*3", "((1+2)*3)"},
		{"-5+10", "((-5)+10)"},
		{"4*(3-1)", "(4*(3-1))"},
		{" 7 - 2 / 1 ", "(7-(2/1))"},
	}
	for _, tc := range tests {
		p := NewParser(tc.input)
		node, err := p.Parse()
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tc.input, err)
			continue
		}
		got := node.String()
		if got != tc.want {
			t.Errorf("Parse(%q).String() = %q, want %q", tc.input, got, tc.want)
		}
	}
}