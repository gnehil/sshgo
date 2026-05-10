package cmd

import (
	"path/filepath"
	"testing"

	"github.com/sshgo/sshgo/internal/config"
)

func TestRunAdd_Success(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := config.SaveConfig(cfgPath, &config.Config{}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	err := runAddWithConfig(cfgPath, "my-server", "192.168.1.10", 22, "deploy", "")
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
	err := runAddWithConfig(cfgPath, "my-server", "5.6.7.8", 22, "admin", "")
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}