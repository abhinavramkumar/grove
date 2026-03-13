package app

import (
	"testing"
	"time"
)

func TestClampCursor_InBounds(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	m.Cursor = 1
	m.ClampCursor()
	if m.Cursor != 1 {
		t.Fatalf("expected cursor=1, got %d", m.Cursor)
	}
}

func TestClampCursor_OutOfBounds(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	m.Cursor = 10
	m.ClampCursor()
	if m.Cursor != 2 {
		t.Fatalf("expected cursor=2 (clamped), got %d", m.Cursor)
	}
}

func TestClampCursor_EmptyList(t *testing.T) {
	m := makeTestListModel(nil)
	m.Cursor = 5
	m.ClampCursor()
	if m.Cursor != 0 {
		t.Fatalf("expected cursor=0, got %d", m.Cursor)
	}
}

func TestClearFilter(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	m.StartFilter()
	m.FilterText = "auth"
	m.Cursor = 2

	m.ClearFilter()

	if m.Filtering {
		t.Error("expected Filtering=false")
	}
	if m.FilterText != "" {
		t.Errorf("expected empty FilterText, got %q", m.FilterText)
	}
	if m.Cursor != 0 {
		t.Errorf("expected cursor=0, got %d", m.Cursor)
	}
}

func TestFormatDuration_LessThanMinute(t *testing.T) {
	if got := formatDuration(30 * time.Second); got != "<1m" {
		t.Fatalf("expected '<1m', got %q", got)
	}
}

func TestFormatDuration_Minutes(t *testing.T) {
	if got := formatDuration(15 * time.Minute); got != "15m" {
		t.Fatalf("expected '15m', got %q", got)
	}
}

func TestFormatDuration_Hours(t *testing.T) {
	if got := formatDuration(5*time.Hour + 30*time.Minute); got != "5h 30m" {
		t.Fatalf("expected '5h 30m', got %q", got)
	}
}

func TestFormatDuration_Days(t *testing.T) {
	if got := formatDuration(48*time.Hour + 3*time.Hour); got != "2d 3h" {
		t.Fatalf("expected '2d 3h', got %q", got)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		in     string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"this is long", 5, "this~"},
		{"ab", 1, "."},
		{"anything", 0, ""},
	}
	for _, tt := range tests {
		if got := truncate(tt.in, tt.maxLen); got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.in, tt.maxLen, got, tt.want)
		}
	}
}

func TestTruncateLeft(t *testing.T) {
	tests := []struct {
		in     string
		maxLen int
		want   string
	}{
		{"/short/path", 20, "/short/path"},
		{"/very/long/directory/path", 10, "~tory/path"},
		{"ab", 2, "ab"},
		{"abc", 2, ".."},
		{"anything", 0, ""},
	}
	for _, tt := range tests {
		if got := truncateLeft(tt.in, tt.maxLen); got != tt.want {
			t.Errorf("truncateLeft(%q, %d) = %q, want %q", tt.in, tt.maxLen, got, tt.want)
		}
	}
}
