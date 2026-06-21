package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sshgo/sshgo/internal/config"
)

func TestRunAddJump_PairingWithIdentity(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	keyA := filepath.Join(dir, "id_a")
	keyB := filepath.Join(dir, "id_b")
	for _, k := range []string{keyA, keyB} {
		if err := os.WriteFile(k, []byte("k"), 0600); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	if err := config.SaveConfig(cfgPath, &config.Config{
		Profiles: []config.Profile{{Name: "db", Host: "10.0.0.5", Port: 22, User: "dba"}},
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	jumps, err := buildJumpHosts(
		[]string{"admin@bastion1:22", "ops@bastion2:22"},
		[]string{keyA, keyB},
	)
	if err != nil {
		t.Fatalf("buildJumpHosts: %v", err)
	}
	if err := applyJumpHosts(cfg, "db", jumps); err != nil {
		t.Fatalf("applyJumpHosts: %v", err)
	}
	if err := config.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, _ := config.LoadConfig(cfgPath)
	p := got.FindProfile("db")
	if p == nil {
		t.Fatal("profile not found")
	}
	if len(p.JumpHosts) != 2 {
		t.Fatalf("expected 2 jump hosts, got %d", len(p.JumpHosts))
	}
	if p.JumpHosts[0].IdentityFile != keyA {
		t.Errorf("jump[0] identity: want %q, got %q", keyA, p.JumpHosts[0].IdentityFile)
	}
	if p.JumpHosts[1].IdentityFile != keyB {
		t.Errorf("jump[1] identity: want %q, got %q", keyB, p.JumpHosts[1].IdentityFile)
	}
}

func TestRunAddJump_PartialIdentity(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "id")
	if err := os.WriteFile(key, []byte("k"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	jumps, err := buildJumpHosts(
		[]string{"a@b1", "b@b2"},
		[]string{key},
	)
	if err != nil {
		t.Fatalf("buildJumpHosts: %v", err)
	}
	if jumps[0].IdentityFile != key {
		t.Errorf("jump[0] identity: want %q, got %q", key, jumps[0].IdentityFile)
	}
	if jumps[1].IdentityFile != "" {
		t.Errorf("jump[1] identity: want empty, got %q", jumps[1].IdentityFile)
	}
}

func TestRunAddJump_IdentityCountExceedsJumps(t *testing.T) {
	_, err := buildJumpHosts(
		[]string{"a@b1"},
		[]string{"/tmp/k1", "/tmp/k2"},
	)
	if err == nil {
		t.Fatal("expected error when --identity-file count exceeds --jump count")
	}
}

func TestRunAddJump_InvalidIdentityFile(t *testing.T) {
	_, err := buildJumpHosts(
		[]string{"a@b1"},
		[]string{"/nonexistent/path/id_rsa"},
	)
	if err == nil {
		t.Fatal("expected error for nonexistent identity file")
	}
}
