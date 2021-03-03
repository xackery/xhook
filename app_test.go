package xhook

import "testing"

func TestTitle(t *testing.T) {
	title, err := Title()
	if err != nil {
		t.Fatalf("title: %s", err)
	}
	if len(title) > 0 {

	}
	//t.Fatalf("title: %s", title)
}
