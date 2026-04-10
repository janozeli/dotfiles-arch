package main

import (
	"slices"
	"testing"
)

func TestFilterEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{"all empty", []string{"", "", ""}, nil},
		{"no empty", []string{"a", "b"}, []string{"a", "b"}},
		{"mixed", []string{"a", "", "b"}, []string{"a", "b"}},
		{"none", []string{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterEmpty(tt.input...)
			if !slices.Equal(got, tt.want) {
				t.Errorf("filterEmpty(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
