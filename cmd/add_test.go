package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sshgo/sshgo/internal/config"
)

func TestRunAdd_Success(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := config.SaveConfig(cfgPath, &config.Config{}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	err := runAddWithConfig(cfgPath, "my-server", "192.168.1.10", 22, "deploy", "", "")
	if err != nil {
		t.Fatalf("runAdd() error: %v", err)
	}
	cfg, _ := config.LoadConfig(cfgPath)
	p := cfg.FindProfile("my-server")
	if p == nil {
		t.Fatal("profile not found after add")
	}
	if p.Host != "192.168.1.10" || p.User != "deploy" || p.Port != 22 {
		t.Errorf("profile data mismatch: %+v", p)
	}
}

func TestRunAdd_DuplicateName(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := config.SaveConfig(cfgPath, &config.Config{
		Profiles: []config.Profile{{Name: "my-server", Host: "1.2.3.4", Port: 22, User: "root"}},
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	err := runAddWithConfig(cfgPath, "my-server", "5.6.7.8", 22, "admin", "", "")
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

func TestRunAdd_WithIdentityFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	keyPath := filepath.Join(dir, "id_ed25519_test")
	if err := os.WriteFile(keyPath, []byte("fake-key"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := config.SaveConfig(cfgPath, &config.Config{}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := runAddWithConfig(cfgPath, "keyed-server", "10.0.0.1", 22, "deploy", "", keyPath)
	if err != nil {
		t.Fatalf("runAdd() error: %v", err)
	}
	cfg, _ := config.LoadConfig(cfgPath)
	p := cfg.FindProfile("keyed-server")
	if p == nil {
		t.Fatal("profile not found after add")
	}
	if p.IdentityFile != keyPath {
		t.Errorf("expected identity_file %q, got %q", keyPath, p.IdentityFile)
	}
}

func TestRunAdd_InvalidIdentityFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := config.SaveConfig(cfgPath, &config.Config{}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := runAddWithConfig(cfgPath, "bad-key", "10.0.0.2", 22, "deploy", "", "/nonexistent/path/id_rsa")
	if err == nil {
		t.Fatal("expected error for nonexistent identity file")
	}

	cfg, _ := config.LoadConfig(cfgPath)
	if cfg.FindProfile("bad-key") != nil {
		t.Fatal("profile should not be saved when identity file is invalid")
	}
}

func TestRunAdd_InsecureIdentityFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	keyPath := filepath.Join(dir, "leaky_key")
	if err := os.WriteFile(keyPath, []byte("k"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := config.SaveConfig(cfgPath, &config.Config{}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := runAddWithConfig(cfgPath, "leaky", "10.0.0.3", 22, "deploy", "", keyPath)
	if err == nil {
		t.Fatal("expected error for identity file with insecure permissions (0o644)")
	}
	if !strings.Contains(err.Error(), "chmod 600") {
		t.Errorf("error should mention chmod 600, got: %v", err)
	}

	cfg, _ := config.LoadConfig(cfgPath)
	if cfg.FindProfile("leaky") != nil {
		t.Fatal("profile should not be saved when identity file has bad perms")
	}
}
