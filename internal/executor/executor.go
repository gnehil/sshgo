package executor

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"strings"
	"github.com/sshgo/sshgo/internal/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
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
		return fmt.Errorf("ssh command not found: %w", err)
	}
	return syscall.Exec(path, append([]string{bin}, args...), os.Environ())
}


func ExecWithPassword(password string, p config.Profile, extraArgs ...string) error {
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)

	sshConf := &ssh.ClientConfig{
		User: p.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
				answers = make([]string, len(questions))
				for i := range questions {
					answers[i] = password
				}
				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var client *ssh.Client
	var err error

	if len(p.JumpHosts) > 0 {
		client, err = connectViaJump(password, p, addr, sshConf)
	} else {
		client, err = ssh.Dial("tcp", addr, sshConf)
	}
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal setup failed: %w", err)
	}
	defer term.Restore(fd, oldState)

	w, h, _ := term.GetSize(fd)
	if err := session.RequestPty("xterm-256color", h, w, modes); err != nil {
		return fmt.Errorf("PTY request failed: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go func() {
		for range sigCh {
			w, h, _ = term.GetSize(fd)
			session.WindowChange(h, w)
		}
	}()

	err = session.Shell()
	if err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	session.Wait()
	return nil
}

func connectViaJump(password string, p config.Profile, addr string, sshConf *ssh.ClientConfig) (*ssh.Client, error) {
	var last *ssh.Client
	for _, j := range p.JumpHosts {
		jAddr := fmt.Sprintf("%s:%d", j.Host, j.Port)
		jConf := &ssh.ClientConfig{
			User: j.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
				ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
					answers = make([]string, len(questions))
					for i := range questions {
						answers[i] = password
					}
					return answers, nil
				}),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		if j.IdentityFile != "" {
			key, err := loadKey(config.ExpandTilde(j.IdentityFile))
			if err == nil && key != nil {
				jConf.Auth = append(jConf.Auth, key)
			}
		}

		conn, err := dialWithBastion(last, "tcp", jAddr, jConf)
		if err != nil {
			return nil, fmt.Errorf("jump host %s (%s@%s:%d): %w", j.Name, j.User, j.Host, j.Port, err)
		}
		last = conn
	}

	if last == nil {
		return ssh.Dial("tcp", addr, sshConf)
	}

	conn, err := dialWithBastion(last, "tcp", addr, sshConf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func dialWithBastion(bastion *ssh.Client, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := bastion.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func loadKey(path string) (ssh.AuthMethod, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	key, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}