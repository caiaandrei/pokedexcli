package main

import (
	"testing"
)

func TestCleanupInput(t *testing.T) {
	cases := map[string]struct {
		input    string
		expected []string
	}{
		"one space":                           {input: "hello world", expected: []string{"hello", "world"}},
		"multiple spaces: start, end, middle": {input: "  hello   world   ", expected: []string{"hello", "world"}},
	}

	for name, c := range cases {
		actual := cleanupInput(c.input)
		for i := range actual {
			actualWord := actual[i]
			expectedWord := c.expected[i]
			if actualWord != expectedWord {
				t.Fatalf("%s: expected: %v and got %v", name, expectedWord, actualWord)
			}
		}
		t.Logf("Test - %s - passed!", name)
	}
}
