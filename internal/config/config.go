package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to get home directory: %w", err)
	}
	return filepath.Join(home, ".sshgo"), nil
}

func DefaultConfigPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func DefaultHistoryPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("malformed config file: %w", err)
	}
	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if _, err := os.Stat(path); err == nil {
		if err := backupConfig(path); err != nil {
			return fmt.Errorf("failed to backup config: %w", err)
		}
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func backupConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.bak.%s", path, timestamp)
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return err
	}
	return cleanupOldBackups(path, 5)
}

func cleanupOldBackups(path string, keep int) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var backups []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), base+".bak.") {
			backups = append(backups, filepath.Join(dir, e.Name()))
		}
	}
	if len(backups) <= keep {
		return nil
	}
	for i := 0; i < len(backups)-keep; i++ {
		os.Remove(backups[i])
	}
	return nil
}

func (c *Config) FindProfile(name string) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i]
		}
	}
	return nil
}

func (c *Config) ResolveJumpHosts(p Profile) Profile {
	for i, jump := range p.JumpHosts {
		ref := c.FindProfile(jump.Host)
		if ref == nil {
			continue
		}

		resolved := jump
		resolved.Host = ref.Host
		resolved.User = ref.User
		resolved.Port = ref.Port
		if resolved.Port == 0 {
			resolved.Port = 22
		}
		if resolved.IdentityFile == "" {
			resolved.IdentityFile = ref.IdentityFile
		}
		if resolved.Name == "" || resolved.Name == jump.User+"@"+jump.Host {
			resolved.Name = ref.Name
		}
		p.JumpHosts[i] = resolved
	}
	return p
}

func (c *Config) AddProfile(p Profile) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == p.Name {
			c.Profiles[i] = p
			return
		}
	}
	c.Profiles = append(c.Profiles, p)
}

func (c *Config) RemoveProfile(name string) bool {
	for i, p := range c.Profiles {
		if p.Name == name {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return true
		}
	}
	return false
}

func (c *Config) FindGroup(name string) *Group {
	for i := range c.Groups {
		if c.Groups[i].Name == name {
			return &c.Groups[i]
		}
	}
	return nil
}
