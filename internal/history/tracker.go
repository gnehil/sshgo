package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Entry struct {
	Name          string    `json:"name"`
	LastConnected time.Time `json:"last_connected"`
	ConnectCount  int       `json:"connect_count"`
}

type Tracker struct {
	path string
	data map[string]*Entry
}

func NewTracker(path string) *Tracker {
	tr := &Tracker{path: path, data: make(map[string]*Entry)}
	tr.load()
	return tr
}

func (t *Tracker) load() {
	f, err := os.Open(t.path)
	if err != nil {
		return
	}
	defer f.Close()
	var entries []Entry
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return
	}
	for i := range entries {
		t.data[entries[i].Name] = &entries[i]
	}
}

func (t *Tracker) Record(name string) {
	if entry, ok := t.data[name]; ok {
		entry.LastConnected = time.Now()
		entry.ConnectCount++
	} else {
		t.data[name] = &Entry{Name: name, LastConnected: time.Now(), ConnectCount: 1}
	}
}

func (t *Tracker) List() []Entry {
	entries := make([]Entry, 0, len(t.data))
	for _, e := range t.data {
		entries = append(entries, *e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastConnected.After(entries[j].LastConnected)
	})
	return entries
}

func (t *Tracker) Recent(n int) []Entry {
	all := t.List()
	if len(all) > n {
		all = all[:n]
	}
	return all
}

func (t *Tracker) Save() error {
	entries := t.List()
	dir := filepath.Dir(t.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.Create(t.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(entries)
}