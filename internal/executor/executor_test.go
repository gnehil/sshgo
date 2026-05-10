package executor

import (
	"testing"
	"github.com/sshgo/sshgo/internal/config"
)

func TestBuildSSHCommand_Simple(t *testing.T) {
	p := config.Profile{Name: "test", Host: "192.168.1.10", Port: 22, User: "deploy"}
	bin, args := BuildSSHCommand(p)
	if bin != "ssh" {
		t.Errorf("expected bin 'ssh', got %s", bin)
	}
	if args[len(args)-1] != "deploy@192.168.1.10" {
		t.Errorf("expected target 'deploy@192.168.1.10', got %s", args[len(args)-1])
	}
}

func TestBuildSSHCommand_CustomPort(t *testing.T) {
	p := config.Profile{Name: "test", Host: "10.0.0.1", Port: 2222, User: "admin"}
	_, args := BuildSSHCommand(p)
	found := false
	for i, a := range args {
		if a == "-p" && i+1 < len(args) && args[i+1] == "2222" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected -p 2222 in args: %v", args)
	}
}

func TestBuildSSHCommand_JumpHosts(t *testing.T) {
	p := config.Profile{Name: "target", Host: "10.0.0.5", Port: 22, User: "deploy",
		JumpHosts: []config.JumpHost{{Name: "bastion", Host: "jump.example.com", Port: 22, User: "jumpuser"}}}
	_, args := BuildSSHCommand(p)
	found := false
	for i, a := range args {
		if a == "-J" && i+1 < len(args) && args[i+1] == "jumpuser@jump.example.com" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected -J jumpuser@jump.example.com in args: %v", args)
	}
}

func TestBuildSSHCommand_PortForward(t *testing.T) {
	p := config.Profile{Name: "target", Host: "10.0.0.5", Port: 22, User: "deploy",
		ForwardPorts: []config.ForwardPort{{LocalPort: 8080, RemoteHost: "localhost", RemotePort: 80}}}
	_, args := BuildSSHCommand(p)
	found := false
	for i, a := range args {
		if a == "-L" && i+1 < len(args) && args[i+1] == "8080:localhost:80" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected -L 8080:localhost:80 in args: %v", args)
	}
}

func TestHasIdentityKey(t *testing.T) {
	if HasIdentityKey(config.Profile{IdentityFile: "~/.ssh/key"}) {
		t.Log("key-based auth detected")
	}
	if HasIdentityKey(config.Profile{}) {
		t.Error("empty profile should not have identity key")
	}
}