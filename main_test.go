package main

import "testing"

func TestValidStatus(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"operational", true},
		{"degraded", true},
		{"down", true},
		{"broken", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := validStatus(tc.in); got != tc.want {
			t.Errorf("validStatus(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
