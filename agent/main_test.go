package main

import (
	"testing"
)

func TestCompute(t *testing.T) {
	tests := []struct {
		name    string
		op      string
		arg1    float64
		arg2    float64
		want    float64
		wantErr bool
	}{
		{"Addition", "+", 2, 3, 5, false},
		{"Subtraction", "-", 10, 3, 7, false},
		{"Multiplication", "*", 3, 4, 12, false},
		{"Division", "/", 12, 3, 4, false},
		{"DivideByZero", "/", 10, 0, 0, true},
		{"UnknownOp", "%", 2, 3, 0, true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := compute(tc.arg1, tc.arg2, tc.op)
			if (err != nil) != tc.wantErr {
				t.Errorf("compute() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("compute() = %v, want %v", got, tc.want)
			}
		})
	}
}
