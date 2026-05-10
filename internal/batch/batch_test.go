package batch

import (
	"testing"
	"github.com/sshgo/sshgo/internal/config"
)

func TestMatchPattern_Exact(t *testing.T) {
	if !matchPattern("prod-web", "prod-web") {
		t.Error("exact match should work")
	}
}

func TestMatchPattern_Glob(t *testing.T) {
	if !matchPattern("prod-web-01", "prod-web-*") {
		t.Error("glob should match")
	}
	if matchPattern("staging-web", "prod-web-*") {
		t.Error("glob should not match different prefix")
	}
}

func TestMatchPattern_CommaList(t *testing.T) {
	if !matchPattern("server-a", "server-a,server-b") {
		t.Error("comma list should match")
	}
	if matchPattern("server-c", "server-a,server-b") {
		t.Error("comma list should not match missing entry")
	}
}

func TestFindTargets_ByGroup(t *testing.T) {
	cfg := &config.Config{
		Profiles: []config.Profile{
			{Name: "web-1", Host: "1.1.1.1", Port: 22, User: "root", Group: "prod"},
			{Name: "web-2", Host: "2.2.2.2", Port: 22, User: "root", Group: "prod"},
			{Name: "dev-1", Host: "3.3.3.3", Port: 22, User: "root", Group: "dev"},
		},
	}
	targets := FindTargets(cfg, "", "prod")
	if len(targets) != 2 {
		t.Errorf("expected 2 prod targets, got %d", len(targets))
	}
}