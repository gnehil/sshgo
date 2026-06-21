package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_NoIssues(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "k")
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	cfg := &Config{Profiles: []Profile{
		{Name: "ok", Host: "1.2.3.4", Port: 22, User: "u", IdentityFile: key},
	}}
	if got := Doctor(cfg); len(got) != 0 {
		t.Errorf("expected no issues, got %+v", got)
	}
}

func TestDoctor_MainIdentityMissing(t *testing.T) {
	cfg := &Config{Profiles: []Profile{
		{Name: "bad", Host: "1.2.3.4", Port: 22, User: "u", IdentityFile: "/no/such/file"},
	}}
	issues := Doctor(cfg)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Profile != "bad" {
		t.Errorf("expected profile 'bad', got %q", issues[0].Profile)
	}
	if issues[0].Jump != "" {
		t.Errorf("expected empty Jump for main identity, got %q", issues[0].Jump)
	}
}

func TestDoctor_MainIdentityInsecurePerms(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "leaky")
	if err := os.WriteFile(key, []byte("k"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	cfg := &Config{Profiles: []Profile{
		{Name: "leaky", Host: "1.2.3.4", Port: 22, User: "u", IdentityFile: key},
	}}
	issues := Doctor(cfg)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if !strings.Contains(issues[0].Err.Error(), "chmod 600") {
		t.Errorf("expected chmod hint, got: %v", issues[0].Err)
	}
}

func TestDoctor_JumpHostIdentityIssues(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "leaky_jump")
	if err := os.WriteFile(key, []byte("k"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	cfg := &Config{Profiles: []Profile{
		{Name: "multi", Host: "1.2.3.4", Port: 22, User: "u", JumpHosts: []JumpHost{
			{Name: "b1", Host: "b1", Port: 22, User: "u", IdentityFile: key},
			{Name: "b2", Host: "b2", Port: 22, User: "u"},
		}},
	}}
	issues := Doctor(cfg)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Jump != "jump[0]" {
		t.Errorf("expected jump[0], got %q", issues[0].Jump)
	}
}

func TestDoctor_MultipleIssuesAcrossProfiles(t *testing.T) {
	cfg := &Config{Profiles: []Profile{
		{Name: "a", Host: "1.2.3.4", Port: 22, User: "u", IdentityFile: "/no/such/a"},
		{Name: "b", Host: "1.2.3.4", Port: 22, User: "u", IdentityFile: "/no/such/b"},
		{Name: "c", Host: "1.2.3.4", Port: 22, User: "u"},
	}}
	issues := Doctor(cfg)
	if len(issues) != 2 {
		t.Errorf("expected 2 issues (a and b), got %d", len(issues))
	}
}
