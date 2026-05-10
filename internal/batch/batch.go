package batch

import (
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"github.com/sshgo/sshgo/internal/config"
	"github.com/sshgo/sshgo/internal/executor"
)

type Target struct {
	Name    string
	Profile config.Profile
	Command string
}

type Result struct {
	Name   string
	Output string
	Err    error
}

func ExecuteAll(targets []Target) []Result {
	var (
		wg      sync.WaitGroup
		results = make([]Result, len(targets))
	)
	for i, t := range targets {
		wg.Add(1)
		go func(idx int, target Target) {
			defer wg.Done()
			bin, args := executor.BuildSSHCommand(target.Profile)
			args = append(args, target.Command)
			cmd := exec.Command(bin, args...)
			out, err := cmd.CombinedOutput()
			results[idx] = Result{Name: target.Name, Output: strings.TrimSpace(string(out)), Err: err}
		}(i, t)
	}
	wg.Wait()
	return results
}

func FindTargets(cfg *config.Config, pattern string, group string) []config.Profile {
	var matched []config.Profile
	for _, p := range cfg.Profiles {
		if group != "" {
			if p.Group == group {
				matched = append(matched, p)
			}
			continue
		}
		if matchPattern(p.Name, pattern) {
			matched = append(matched, p)
		}
	}
	return matched
}

func matchPattern(name, pattern string) bool {
	if strings.Contains(pattern, ",") {
		for _, p := range strings.Split(pattern, ",") {
			if name == strings.TrimSpace(p) {
				return true
			}
		}
		return false
	}
	m, _ := filepath.Match(pattern, name)
	return m
}