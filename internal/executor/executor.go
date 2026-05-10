package executor

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"strings"
	"github.com/sshgo/sshgo/internal/config"
)

func BuildSSHCommand(p config.Profile) (bin string, args []string) {
	target := p.User + "@" + p.Host
	if p.Port != 0 && p.Port != 22 {
		target = fmt.Sprintf("%s@%s:%d", p.User, p.Host, p.Port)
		args = append(args, "-p", fmt.Sprintf("%d", p.Port))
	}
	if p.IdentityFile != "" {
		args = append(args, "-i", config.ExpandTilde(p.IdentityFile))
	}
	if len(p.JumpHosts) > 0 {
		var jumps []string
		for _, j := range p.JumpHosts {
			jumpAddr := j.User + "@" + j.Host
			if j.Port != 0 && j.Port != 22 {
				jumpAddr = fmt.Sprintf("%s@%s:%d", j.User, j.Host, j.Port)
			}
			jumps = append(jumps, jumpAddr)
		}
		args = append(args, "-J", strings.Join(jumps, ","))
	}
	for _, f := range p.ForwardPorts {
		local := fmt.Sprintf("%d:%s:%d", f.LocalPort, f.RemoteHost, f.RemotePort)
		args = append(args, "-L", local)
	}
	if p.KeepaliveInterval > 0 {
		args = append(args, "-o", fmt.Sprintf("ServerAliveInterval=%d", p.KeepaliveInterval))
	}
	if p.ServerAliveCount > 0 {
		args = append(args, "-o", fmt.Sprintf("ServerAliveCountMax=%d", p.ServerAliveCount))
	}
	args = append(args, target)
	return "ssh", args
}

func HasIdentityKey(p config.Profile) bool {
	return p.IdentityFile != ""
}

func ExecSSH(p config.Profile, extraArgs ...string) error {
	bin, args := BuildSSHCommand(p)
	args = append(args, extraArgs...)
	path, err := exec.LookPath(bin)
	if err != nil {
		return fmt.Errorf("找不到 ssh 命令: %w", err)
	}
	return syscall.Exec(path, append([]string{bin}, args...), os.Environ())
}

func ExecWithPassword(password string, p config.Profile, extraArgs ...string) error {
	_, args := BuildSSHCommand(p)
	args = append(args, extraArgs...)

	if sshpassPath, err := exec.LookPath("sshpass"); err == nil {
		sshArgs := append([]string{"sshpass", "-p", password, "ssh"}, args...)
		return syscall.Exec(sshpassPath, sshArgs, os.Environ())
	}

	return ExecSSH(p, extraArgs...)
}