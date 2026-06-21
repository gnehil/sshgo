package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type Config struct {
	Profiles []Profile `yaml:"profiles"`
	Groups   []Group   `yaml:"groups"`
}

type Profile struct {
	Name              string        `yaml:"name"`
	Host              string        `yaml:"host"`
	Port              int           `yaml:"port"`
	User              string        `yaml:"user"`
	Group             string        `yaml:"group,omitempty"`
	IdentityFile      string        `yaml:"identity_file,omitempty"`
	JumpHosts         []JumpHost    `yaml:"jump_hosts,omitempty"`
	ForwardPorts      []ForwardPort `yaml:"forward_ports,omitempty"`
	KeepaliveInterval int           `yaml:"keepalive_interval,omitempty"`
	ServerAliveCount  int           `yaml:"server_alive_count,omitempty"`
}

type JumpHost struct {
	Name         string `yaml:"name"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	IdentityFile string `yaml:"identity_file,omitempty"`
	Password     string `yaml:"-"`
}

type ForwardPort struct {
	LocalPort  int    `yaml:"local_port"`
	RemoteHost string `yaml:"remote_host"`
	RemotePort int    `yaml:"remote_port"`
}

type Group struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

var validNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func (p *Profile) Validate(cfg *Config) error {
	if p.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !validNameRegex.MatchString(p.Name) {
		return fmt.Errorf("invalid name format: %q", p.Name)
	}
	if p.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if p.Port < 1 || p.Port > 65535 {
		return fmt.Errorf("invalid port: %d", p.Port)
	}
	if p.User == "" {
		return fmt.Errorf("user cannot be empty")
	}
	if p.Group != "" {
		found := false
		for _, g := range cfg.Groups {
			if g.Name == p.Group {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("group %q does not exist", p.Group)
		}
	}
	if p.IdentityFile != "" {
		if err := ValidateIdentityFile(ExpandTilde(p.IdentityFile)); err != nil {
			return err
		}
	}
	for i, j := range p.JumpHosts {
		if j.Host == "" {
			return fmt.Errorf("jump_hosts[%d].host cannot be empty", i)
		}
		if j.User == "" {
			return fmt.Errorf("jump_hosts[%d].user cannot be empty", i)
		}
		if j.Port == 0 {
			p.JumpHosts[i].Port = 22
		}
		if j.Port < 1 || j.Port > 65535 {
			return fmt.Errorf("invalid jump_hosts[%d].port: %d", i, j.Port)
		}
		if j.IdentityFile != "" {
			if err := ValidateIdentityFile(ExpandTilde(j.IdentityFile)); err != nil {
				return fmt.Errorf("jump_hosts[%d].%w", i, err)
			}
		}
	}
	return nil
}

func ExpandTilde(path string) string {
	if len(path) == 0 {
		return path
	}
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if len(path) == 1 {
			return home
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// ValidateIdentityFile checks that path points to a readable file whose
// permissions OpenSSH will accept. An empty path is a no-op.
//
// The mask 0o037 mirrors OpenSSH's secure_filename() check for user-owned
// keys (the common case when sshgo is invoked as the same user who owns
// the key): any of S_IWGRP | S_IXGRP | S_IROTH | S_IWOTH | S_IXOTH is
// rejected. Group-read-only (e.g. 0o640) is accepted, matching OpenSSH's
// actual behavior, so we surface the same failure at profile creation
// time rather than as a cryptic "Permissions 0644 ... are too open" at
// connect time.
//
// Known limitations:
//   - For root-owned keys OpenSSH applies a stricter check (mask 0o077,
//     rejecting any group/other access). This function does not replicate
//     that branch, so a root-owned key with 0o040 group read will pass
//     here but fail at connect time. This is an exotic edge case (sshgo
//     runs as an unprivileged user in essentially all real scenarios).
//   - There is an inherent TOCTOU window between this stat and the
//     eventual ssh invocation. Acceptable for a local CLI; an attacker
//     would need both filesystem access and millisecond timing.
//   - The returned error includes the absolute path. On a shared system
//     this could leak home-directory structure; we accept that trade-off
//     because the user invoking sshgo already knows their own paths.
func ValidateIdentityFile(path string) error {
	if path == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("identity file %s: %w", path, err)
	}
	if mode := info.Mode().Perm(); mode&0o037 != 0 {
		return fmt.Errorf("identity file %s has insecure permissions (%#o); run: chmod 600 %s", path, mode, path)
	}
	return nil
}
