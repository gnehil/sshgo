package sshconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSSHConfig_Basic(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config")
	content := "Host myserver\n    HostName 192.168.1.100\n    User deploy\n    Port 2222\n\nHost dbserver\n    HostName db.internal\n    User dbadmin\n"
	os.WriteFile(cfgFile, []byte(content), 0644)
	profiles, err := ParseSSHConfig(cfgFile)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].Name != "myserver" || profiles[0].Host != "192.168.1.100" || profiles[0].User != "deploy" || profiles[0].Port != 2222 {
		t.Errorf("profile 0 mismatch: %+v", profiles[0])
	}
}

func TestParseSSHConfig_ProxyJump(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config")
	content := "Host target\n    HostName 10.0.0.5\n    User admin\n    ProxyJump bastion.example.com\n"
	os.WriteFile(cfgFile, []byte(content), 0644)
	profiles, err := ParseSSHConfig(cfgFile)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(profiles[0].JumpHosts) != 1 {
		t.Fatalf("expected 1 jump host, got %d", len(profiles[0].JumpHosts))
	}
	if profiles[0].JumpHosts[0].Host != "bastion.example.com" {
		t.Errorf("jump host mismatch: %+v", profiles[0].JumpHosts[0])
	}
}