package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sshgo/sshgo/internal/config"
)

func TestRunDoctor_NoIssues(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	key := filepath.Join(dir, "k")
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := config.SaveConfig(cfgPath, &config.Config{
		Profiles: []config.Profile{{Name: "ok", Host: "1.2.3.4", Port: 22, User: "u", IdentityFile: key}},
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := runDoctorWithConfig(cfgPath); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestRunDoctor_ReportsIssues(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	leaky := filepath.Join(dir, "leaky")
	if err := os.WriteFile(leaky, []byte("k"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := config.SaveConfig(cfgPath, &config.Config{
		Profiles: []config.Profile{{Name: "leaky-srv", Host: "1.2.3.4", Port: 22, User: "u", IdentityFile: leaky}},
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	err := runDoctorWithConfig(cfgPath)
	if err == nil {
		t.Fatal("expected error when issues found")
	}
	if !strings.Contains(err.Error(), "issue") {
		t.Errorf("error should mention 'issue', got: %v", err)
	}
}

func TestRunDoctor_EmptyConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := config.SaveConfig(cfgPath, &config.Config{}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := runDoctorWithConfig(cfgPath); err != nil {
		t.Errorf("empty config should be healthy, got: %v", err)
	}
}
