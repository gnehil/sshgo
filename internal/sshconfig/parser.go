package sshconfig

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/sshgo/sshgo/internal/config"
)

func ParseSSHConfig(path string) ([]config.Profile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %w", path, err)
	}
	defer f.Close()
	var profiles []config.Profile
	var current *config.Profile
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "host":
			if current != nil {
				profiles = append(profiles, *current)
			}
			current = &config.Profile{Name: value, Port: 22}
		case "hostname":
			if current != nil {
				current.Host = value
			}
		case "user":
			if current != nil {
				current.User = value
			}
		case "port":
			if current != nil {
				p, _ := strconv.Atoi(value)
				current.Port = p
			}
		case "identityfile":
			if current != nil {
				current.IdentityFile = value
			}
		case "proxyjump":
			if current != nil {
				for _, jp := range strings.Split(value, ",") {
					current.JumpHosts = append(current.JumpHosts, parseJumpHost(jp))
				}
			}
		case "include":
			included, err := parseInclude(value, path)
			if err == nil {
				profiles = append(profiles, included...)
			}
		}
	}
	if current != nil {
		profiles = append(profiles, *current)
	}
	return profiles, scanner.Err()
}

func parseJumpHost(addr string) config.JumpHost {
	jh := config.JumpHost{Port: 22}
	addr = strings.TrimSpace(addr)
	if at := strings.Index(addr, "@"); at >= 0 {
		jh.User = addr[:at]
		addr = addr[at+1:]
	}
	if colon := strings.LastIndex(addr, ":"); colon >= 0 {
		p, _ := strconv.Atoi(addr[colon+1:])
		if p > 0 {
			jh.Port = p
		}
		jh.Host = addr[:colon]
	} else {
		jh.Host = addr
	}
	if jh.User == "" {
		jh.User = "root"
	}
	jh.Name = jh.User + "@" + jh.Host
	return jh
}

func parseInclude(pattern, sourcePath string) ([]config.Profile, error) {
	dir := filepath.Dir(sourcePath)
	pattern = config.ExpandTilde(pattern)
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(dir, pattern)
	}
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	var profiles []config.Profile
	for _, m := range matches {
		ps, err := ParseSSHConfig(m)
		if err == nil {
			profiles = append(profiles, ps...)
		}
	}
	return profiles, nil
}