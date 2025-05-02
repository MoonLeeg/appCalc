package agent

import "testing"

func TestCompute(t *testing.T) {
	tests := []struct {
		name    string
		arg1    float64
		arg2    float64
		op      string
		want    float64
		wantErr bool
	}{
		{"Addition", 2, 3, "+", 5, false},
		{"Subtraction", 10, 3, "-", 7, false},
		{"Multiplication", 3, 4, "*", 12, false},
		{"Division", 12, 3, "/", 4, false},
		{"DivideByZero", 10, 0, "/", 0, true},
		{"UnknownOp", 2, 3, "%", 0, true},
	}
	for _, tc := range tests {
		got, err := compute(tc.arg1, tc.arg2, tc.op)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: compute(%v, %v, %q) error = %v, wantErr %v", tc.name, tc.arg1, tc.arg2, tc.op, err, tc.wantErr)
			continue
		}
		if !tc.wantErr && got != tc.want {
			t.Errorf("%s: compute(%v, %v, %q) = %v, want %v", tc.name, tc.arg1, tc.arg2, tc.op, got, tc.want)
		}
	}
}