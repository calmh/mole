package main

import (
	"testing"
)

var testcases = []struct {
	src []string
	s   string
	n   int
	dst []string
}{
	{[]string{"a", "b", "c"}, "a", 3, []string{"b", "c", "a"}},
	{[]string{"a", "b", "c"}, "b", 3, []string{"a", "c", "b"}},
	{[]string{"a", "b", "c"}, "c", 3, []string{"a", "b", "c"}},

	{[]string{"a", "b", "c"}, "a", 0, []string{}},
	{[]string{"a", "b", "c"}, "a", 1, []string{"a"}},
	{[]string{"a", "b", "c"}, "a", 2, []string{"c", "a"}},
	{[]string{"a", "b", "c"}, "a", 3, []string{"b", "c", "a"}},
	{[]string{"a", "b", "c"}, "a", 9, []string{"b", "c", "a"}},

	{[]string{"a", "b", "c"}, "d", 9, []string{"a", "b", "c", "d"}},
	{[]string{"a", "b", "c"}, "d", 4, []string{"a", "b", "c", "d"}},
	{[]string{"a", "b", "c"}, "d", 3, []string{"b", "c", "d"}},
	{[]string{"a", "b", "c"}, "d", 2, []string{"c", "d"}},
	{[]string{"a", "b", "c"}, "d", 1, []string{"d"}},
	{[]string{"a", "b", "c"}, "d", 0, []string{}},
}

func sliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestNewIPList(t *testing.T) {
	for _, tc := range testcases {
		if d := newIPList(tc.src, tc.s, tc.n); !sliceEquals(d, tc.dst) {
			t.Errorf("incorrect; %v", tc.dst)
		}
	}
}
