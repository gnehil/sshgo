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
	Name              string       `yaml:"name"`
	Host              string       `yaml:"host"`
	Port              int          `yaml:"port"`
	User              string       `yaml:"user"`
	Group             string       `yaml:"group,omitempty"`
	IdentityFile      string       `yaml:"identity_file,omitempty"`
	JumpHosts         []JumpHost   `yaml:"jump_hosts,omitempty"`
	ForwardPorts      []ForwardPort `yaml:"forward_ports,omitempty"`
	KeepaliveInterval int          `yaml:"keepalive_interval,omitempty"`
	ServerAliveCount  int          `yaml:"server_alive_count,omitempty"`
}

type JumpHost struct {
	Name         string `yaml:"name"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	IdentityFile string `yaml:"identity_file,omitempty"`
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
		return fmt.Errorf("name 不能为空")
	}
	if !validNameRegex.MatchString(p.Name) {
		return fmt.Errorf("无效的 name 格式: %q", p.Name)
	}
	if p.Host == "" {
		return fmt.Errorf("host 不能为空")
	}
	if p.Port < 1 || p.Port > 65535 {
		return fmt.Errorf("端口号无效: %d", p.Port)
	}
	if p.User == "" {
		return fmt.Errorf("user 不能为空")
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
			return fmt.Errorf("分组 %q 不存在", p.Group)
		}
	}
	if p.IdentityFile != "" {
		expanded := ExpandTilde(p.IdentityFile)
		if _, err := os.Stat(expanded); os.IsNotExist(err) {
			return fmt.Errorf("密钥文件不存在: %s", expanded)
		}
	}
	for i, j := range p.JumpHosts {
		if j.Host == "" {
			return fmt.Errorf("jump_hosts[%d].host 不能为空", i)
		}
		if j.User == "" {
			return fmt.Errorf("jump_hosts[%d].user 不能为空", i)
		}
		if j.Port == 0 {
			j.Port = 22
		}
		if j.Port < 1 || j.Port > 65535 {
			return fmt.Errorf("jump_hosts[%d].port 无效: %d", i, j.Port)
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