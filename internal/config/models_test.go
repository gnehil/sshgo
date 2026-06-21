package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileValidate_Success(t *testing.T) {
	p := Profile{Name: "my-server", Host: "192.168.1.10", Port: 22, User: "deploy"}
	if err := p.Validate(&Config{}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestProfileValidate_MissingName(t *testing.T) {
	p := Profile{Name: "", Host: "192.168.1.10", Port: 22, User: "deploy"}
	if err := p.Validate(&Config{}); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestProfileValidate_InvalidName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"-bad", true}, {"bad name", true}, {"bad.name", true}, {"good-name", false}, {"good123", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Profile{Name: tt.name, Host: "192.168.1.10", Port: 22, User: "deploy"}
			err := p.Validate(&Config{})
			if (err != nil) != tt.want {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.want)
			}
		})
	}
}

func TestProfileValidate_InvalidPort(t *testing.T) {
	p := Profile{Name: "test", Host: "192.168.1.10", Port: 99999, User: "deploy"}
	if err := p.Validate(&Config{}); err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestGroup_InvalidName(t *testing.T) {
	p := &Profile{Name: "test", Host: "192.168.1.10", Port: 22, User: "deploy", Group: "nonexistent"}
	if err := p.Validate(&Config{}); err == nil {
		t.Fatal("expected error for undefined group")
	}
}

func TestGroup_ValidGroup(t *testing.T) {
	p := &Profile{Name: "test", Host: "192.168.1.10", Port: 22, User: "deploy", Group: "prod"}
	cfg := &Config{Groups: []Group{{Name: "prod", Description: "prod servers"}}}
	if err := p.Validate(cfg); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestExpandTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := ExpandTilde("~/.ssh/id_rsa")
	want := filepath.Join(home, ".ssh/id_rsa")
	if got != want {
		t.Errorf("ExpandTilde() = %v, want %v", got, want)
	}
}

func TestValidateIdentityFile_SecurePerms(t *testing.T) {
	dir := t.TempDir()
	tests := []os.FileMode{0o600, 0o400, 0o640}
	for _, mode := range tests {
		t.Run(mode.String(), func(t *testing.T) {
			path := filepath.Join(dir, "key_"+mode.String())
			if err := os.WriteFile(path, []byte("k"), mode); err != nil {
				t.Fatalf("setup: %v", err)
			}
			if err := ValidateIdentityFile(path); err != nil {
				t.Errorf("ValidateIdentityFile(%o) error = %v, want nil", mode, err)
			}
		})
	}
}

func TestValidateIdentityFile_InsecurePerms(t *testing.T) {
	dir := t.TempDir()
	tests := []os.FileMode{0o644, 0o666, 0o755, 0o777, 0o604, 0o605}
	for _, mode := range tests {
		t.Run(mode.String(), func(t *testing.T) {
			path := filepath.Join(dir, "key_"+mode.String())
			if err := os.WriteFile(path, []byte("k"), mode); err != nil {
				t.Fatalf("setup: %v", err)
			}
			err := ValidateIdentityFile(path)
			if err == nil {
				t.Errorf("ValidateIdentityFile(%o) error = nil, want error", mode)
				return
			}
			if !strings.Contains(err.Error(), "chmod 600") {
				t.Errorf("error message should mention chmod 600, got: %v", err)
			}
		})
	}
}

func TestValidateIdentityFile_NotFound(t *testing.T) {
	err := ValidateIdentityFile("/nonexistent/path/id_rsa")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestValidateIdentityFile_Empty(t *testing.T) {
	if err := ValidateIdentityFile(""); err != nil {
		t.Errorf("empty path should be a no-op, got: %v", err)
	}
}

func TestProfileValidate_InsecureJumpHostIdentity(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "leaky_jump_key")
	if err := os.WriteFile(key, []byte("k"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	p := &Profile{
		Name: "test",
		Host: "10.0.0.1",
		Port: 22,
		User: "u",
		JumpHosts: []JumpHost{{
			Name:         "bastion",
			Host:         "b1",
			Port:         22,
			User:         "u",
			IdentityFile: key,
		}},
	}
	err := p.Validate(&Config{})
	if err == nil {
		t.Fatal("expected error for jump host identity with insecure perms")
	}
	if !strings.Contains(err.Error(), "jump_hosts[0]") {
		t.Errorf("error should reference jump_hosts[0], got: %v", err)
	}
}

func TestProfileValidate_SecureJumpHostIdentity(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "ok_jump_key")
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	p := &Profile{
		Name: "test",
		Host: "10.0.0.1",
		Port: 22,
		User: "u",
		JumpHosts: []JumpHost{{
			Name:         "bastion",
			Host:         "b1",
			Port:         22,
			User:         "u",
			IdentityFile: key,
		}},
	}
	if err := p.Validate(&Config{}); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}