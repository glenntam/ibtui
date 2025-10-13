package main

import "testing"

func TestRenderTabStrings(t *testing.T) {
	m := &model{}
	// Call render functions that should return fixed strings
	if s := m.renderWatchlistContent(); s == "" {
		t.Fatalf("renderWatchlistContent returned empty string")
	}
	if s := m.renderOrderEntryContent(); s == "" {
		t.Fatalf("renderOrderEntryContent returned empty string")
	}
	if s := m.renderOpenOrdersContent(); s == "" {
		t.Fatalf("renderOpenOrdersContent returned empty string")
	}
	if s := m.renderAlgoContent(); s == "" {
		t.Fatalf("renderAlgoContent returned empty string")
	}
}
