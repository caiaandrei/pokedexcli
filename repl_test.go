package main

import (
	pokecache "pokedexcli/internal"
	"testing"
	"time"
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

func TestCacheAddGet(t *testing.T) {
	cache := pokecache.NewCache(1 * time.Second)

	cache.Add("abc", []byte("hello"))

	val, ok1 := cache.Get("abc")
	val2, ok2 := cache.Get("cab")

	if !ok1 {
		t.Fatal("expected key to exist")
	}

	if ok2 {
		t.Fatal("expected key NOT to exist")
	}

	if string(val) != "hello" {
		t.Fatalf("unexpected value: %s", string(val))
	}

	if string(val2) != "" {
		t.Fatalf("unexpected value: %s", string(val2))
	}

}
