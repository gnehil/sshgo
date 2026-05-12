package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func TestLoadSave_RoundTrip(t *testing.T) {
	dir := setupTestDir(t)
	cfgPath := filepath.Join(dir, "config.yaml")
	original := &Config{
		Profiles: []Profile{{Name: "test", Host: "192.168.1.1", Port: 22, User: "admin"}},
		Groups:   []Group{{Name: "prod", Description: "prod"}},
	}
	if err := SaveConfig(cfgPath, original); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}
	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if len(loaded.Profiles) != 1 || loaded.Profiles[0].Name != "test" {
		t.Errorf("expected profile 'test', got %+v", loaded.Profiles)
	}
}

func TestLoadConfig_NonExistent(t *testing.T) {
	dir := setupTestDir(t)
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg == nil || len(cfg.Profiles) != 0 {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestBackup_BeforeSave(t *testing.T) {
	dir := setupTestDir(t)
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := SaveConfig(cfgPath, &Config{
		Profiles: []Profile{{Name: "old", Host: "1.2.3.4", Port: 22, User: "root"}},
	}); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}
	if err := SaveConfig(cfgPath, &Config{
		Profiles: []Profile{{Name: "new", Host: "5.6.7.8", Port: 22, User: "root"}},
	}); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	hasBackup := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak.") {
			hasBackup = true
			break
		}
	}
	if !hasBackup {
		t.Error("expected backup file to exist")
	}
}

func TestResolveJumpHostsUsesReferencedProfile(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "dev-jumper-hk", Host: "8.212.50.246", Port: 22, User: "liheng", IdentityFile: "~/.ssh/dev"},
		},
	}
	p := Profile{
		Name: "dev-byteplus-hk",
		Host: "10.26.20.3",
		Port: 22,
		User: "liheng",
		JumpHosts: []JumpHost{{
			Name: "root@dev-jumper-hk",
			Host: "dev-jumper-hk",
			Port: 22,
			User: "root",
		}},
	}

	resolved := cfg.ResolveJumpHosts(p)

	if len(resolved.JumpHosts) != 1 {
		t.Fatalf("expected one jump host, got %d", len(resolved.JumpHosts))
	}
	jump := resolved.JumpHosts[0]
	if jump.Host != "8.212.50.246" {
		t.Fatalf("expected jump host address from referenced profile, got %q", jump.Host)
	}
	if jump.User != "liheng" {
		t.Fatalf("expected jump user from referenced profile, got %q", jump.User)
	}
	if jump.IdentityFile != "~/.ssh/dev" {
		t.Fatalf("expected identity file from referenced profile, got %q", jump.IdentityFile)
	}
}

func TestSaveConfigOmitsRuntimeJumpPassword(t *testing.T) {
	dir := setupTestDir(t)
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := &Config{
		Profiles: []Profile{{
			Name: "dev-byteplus-hk",
			Host: "10.26.20.3",
			Port: 22,
			User: "liheng",
			JumpHosts: []JumpHost{{
				Name:     "dev-jumper-hk",
				Host:     "8.212.50.246",
				Port:     22,
				User:     "liheng",
				Password: "jump-secret",
			}},
		}},
	}

	if err := SaveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if strings.Contains(string(data), "jump-secret") {
		t.Fatalf("runtime jump password was written to config:\n%s", data)
	}
}
