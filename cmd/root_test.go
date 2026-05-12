package cmd

import (
	"errors"
	"testing"

	"github.com/sshgo/sshgo/internal/config"
)

func TestProfileShortcutNameUsesExistingProfile(t *testing.T) {
	cfg := &config.Config{
		Profiles: []config.Profile{{Name: "dev-byteplus-hk", Host: "10.26.20.3", Port: 22, User: "liheng"}},
	}

	name, ok := profileShortcutName([]string{"dev-byteplus-hk"}, cfg)
	if !ok {
		t.Fatal("expected profile shortcut to be enabled")
	}
	if name != "dev-byteplus-hk" {
		t.Fatalf("expected dev-byteplus-hk, got %q", name)
	}
}

func TestProfileShortcutNameIgnoresCommands(t *testing.T) {
	cfg := &config.Config{
		Profiles: []config.Profile{{Name: "connect", Host: "10.26.20.3", Port: 22, User: "liheng"}},
	}

	if _, ok := profileShortcutName([]string{"connect"}, cfg); ok {
		t.Fatal("expected command name not to be treated as profile shortcut")
	}
}

func TestWithStoredJumpPasswords(t *testing.T) {
	p := config.Profile{
		Name: "dev-byteplus-hk",
		JumpHosts: []config.JumpHost{{
			Name: "dev-jumper-hk",
			Host: "8.212.50.246",
			Port: 22,
			User: "liheng",
		}},
	}

	got := withStoredJumpPasswords(p, func(name string) (string, error) {
		if name != "dev-jumper-hk" {
			t.Fatalf("unexpected credential lookup for %q", name)
		}
		return "jump-secret", nil
	})

	if got.JumpHosts[0].Password != "jump-secret" {
		t.Fatalf("expected stored jump password to be attached, got %q", got.JumpHosts[0].Password)
	}
}

func TestWithStoredJumpPasswordsIgnoresMissingPassword(t *testing.T) {
	p := config.Profile{
		Name: "dev-byteplus-hk",
		JumpHosts: []config.JumpHost{{
			Name: "dev-jumper-hk",
			Host: "8.212.50.246",
			Port: 22,
			User: "liheng",
		}},
	}

	got := withStoredJumpPasswords(p, func(string) (string, error) {
		return "", errors.New("not found")
	})

	if got.JumpHosts[0].Password != "" {
		t.Fatalf("expected missing password to be ignored, got %q", got.JumpHosts[0].Password)
	}
}

func TestWithPromptedJumpPasswords(t *testing.T) {
	p := config.Profile{
		Name: "dev-byteplus-hk",
		JumpHosts: []config.JumpHost{{
			Name: "dev-jumper-hk",
			Host: "8.212.50.246",
			Port: 22,
			User: "liheng",
		}},
	}

	got, err := withPromptedJumpPasswords(p, func(jump config.JumpHost) (string, error) {
		if jump.Name != "dev-jumper-hk" {
			t.Fatalf("unexpected jump prompt for %q", jump.Name)
		}
		return "jump-secret", nil
	})
	if err != nil {
		t.Fatalf("withPromptedJumpPasswords() error: %v", err)
	}
	if got.JumpHosts[0].Password != "jump-secret" {
		t.Fatalf("expected prompted jump password to be attached, got %q", got.JumpHosts[0].Password)
	}
}

func TestWithPromptedJumpPasswordsSkipsIdentityFile(t *testing.T) {
	p := config.Profile{
		Name: "dev-byteplus-hk",
		JumpHosts: []config.JumpHost{{
			Name:         "dev-jumper-hk",
			Host:         "8.212.50.246",
			Port:         22,
			User:         "liheng",
			IdentityFile: "~/.ssh/id_rsa",
		}},
	}

	got, err := withPromptedJumpPasswords(p, func(config.JumpHost) (string, error) {
		t.Fatal("did not expect prompt for jump host with identity file")
		return "", nil
	})
	if err != nil {
		t.Fatalf("withPromptedJumpPasswords() error: %v", err)
	}
	if got.JumpHosts[0].Password != "" {
		t.Fatalf("expected no prompted password, got %q", got.JumpHosts[0].Password)
	}
}
