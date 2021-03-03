package xhook

import "testing"

func TestKeyToggle(t *testing.T) {
	err := KeyToggle("a", true)
	if err != nil {
		t.Fatalf("keyToggle: %s", err)
	}
}
