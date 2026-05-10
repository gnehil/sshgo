package config

import (
	"os"
	"path/filepath"
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