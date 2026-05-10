package history

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestRecord_ThenList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	tr := NewTracker(path)
	tr.Record("my-server")
	tr.Record("my-server")
	tr.Record("db-server")
	entries := tr.List()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "db-server" {
		t.Errorf("expected db-server first, got %s", entries[0].Name)
	}
	if entries[1].Name != "my-server" || entries[1].ConnectCount != 2 {
		t.Errorf("expected my-server count=2 at index 1, got %+v", entries[1])
	}
}

func TestRecent_Limit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	tr := NewTracker(path)
	for i := 0; i < 20; i++ {
		tr.Record(fmt.Sprintf("server-%d", i))
	}
	recent := tr.Recent(5)
	if len(recent) != 5 {
		t.Fatalf("expected 5 recent, got %d", len(recent))
	}
	if recent[0].Name != "server-19" {
		t.Errorf("expected server-19, got %s", recent[0].Name)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	tr := NewTracker(path)
	tr.Record("test")
	_ = tr.Save()
	tr2 := NewTracker(path)
	entries := tr2.List()
	if len(entries) != 1 || entries[0].Name != "test" {
		t.Errorf("expected test entry, got %+v", entries)
	}
}