package executor

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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
	port := p.Port
	if port == 0 {
		port = 22
	}
	addr := fmt.Sprintf("%s:%d", p.Host, port)

	sshConf := &ssh.ClientConfig{
		User:            p.User,
		Auth:            passwordAuthMethods(password),
		HostKeyCallback: knownHostsCallback(),
		Timeout:         10 * time.Second,
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
	defer signal.Stop(sigCh)

	err = session.Shell()
	if err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	session.Wait()
	return nil
}

func connectViaJump(password string, p config.Profile, addr string, sshConf *ssh.ClientConfig) (*ssh.Client, error) {
	var jumpClients []*ssh.Client
	closeJumpClients := func() {
		for i := len(jumpClients) - 1; i >= 0; i-- {
			jumpClients[i].Close()
		}
	}

	for _, j := range p.JumpHosts {
		jConf := &ssh.ClientConfig{
			User:            j.User,
			HostKeyCallback: knownHostsCallback(),
			Timeout:         10 * time.Second,
		}
		// TODO: Jump hosts currently don't have per-host credential storage.
		// Using key-based auth only. If a jump host requires password auth,
		// add per-jump-host credential storage and integrate it here.
		if j.IdentityFile != "" {
			key, err := loadKey(config.ExpandTilde(j.IdentityFile))
			if err == nil && key != nil {
				jConf.Auth = append(jConf.Auth, key)
			}
		}
		jumpPassword := j.Password
		if jumpPassword == "" {
			jumpPassword = password
		}
		if jumpPassword != "" {
			jConf.Auth = append(jConf.Auth, passwordAuthMethods(jumpPassword)...)
		}

		jAddr := fmt.Sprintf("%s:%d", j.Host, j.Port)
		var netConn net.Conn
		var err error
		if len(jumpClients) == 0 {
			netConn, err = net.DialTimeout("tcp", jAddr, 10*time.Second)
		} else {
			netConn, err = jumpClients[len(jumpClients)-1].Dial("tcp", jAddr)
		}
		if err != nil {
			closeJumpClients()
			return nil, fmt.Errorf("jump host connect %s (%s@%s:%d): %w", j.Name, j.User, j.Host, j.Port, err)
		}
		c, chans, reqs, err := ssh.NewClientConn(netConn, jAddr, jConf)
		if err != nil {
			netConn.Close()
			closeJumpClients()
			return nil, fmt.Errorf("jump host handshake %s (%s@%s:%d): %w", j.Name, j.User, j.Host, j.Port, err)
		}
		jumpClients = append(jumpClients, ssh.NewClient(c, chans, reqs))
	}

	if len(jumpClients) == 0 {
		netConn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			return nil, fmt.Errorf("target connect %s: %w", addr, err)
		}
		c, chans, reqs, err := ssh.NewClientConn(netConn, addr, sshConf)
		if err != nil {
			netConn.Close()
			return nil, fmt.Errorf("target handshake %s: %w", addr, err)
		}
		return ssh.NewClient(c, chans, reqs), nil
	}

	netConn, err := jumpClients[len(jumpClients)-1].Dial("tcp", addr)
	if err != nil {
		closeJumpClients()
		return nil, fmt.Errorf("target connect through bastion %s: %w", addr, err)
	}
	c, chans, reqs, err := ssh.NewClientConn(netConn, addr, sshConf)
	if err != nil {
		netConn.Close()
		closeJumpClients()
		return nil, err
	}
	target := ssh.NewClient(c, chans, reqs)
	go func() {
		target.Wait()
		closeJumpClients()
	}()
	return target, nil
}

func passwordAuthMethods(password string) []ssh.AuthMethod {
	return []ssh.AuthMethod{
		ssh.Password(password),
		ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			answers = make([]string, len(questions))
			for i := range questions {
				answers[i] = password
			}
			return answers, nil
		}),
	}
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

// knownHostsCallback reads ~/.ssh/known_hosts and returns a HostKeyCallback.
// It falls back to prompting the user if no known entry is found.
func knownHostsCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		known, err := checkKnownHosts(hostname, remote, key)
		if err != nil {
			return err
		}
		if !known {
			return promptHostKey(hostname, remote, key)
		}
		return nil
	}
}

func checkKnownHosts(hostname string, remote net.Addr, key ssh.PublicKey) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("cannot determine home directory: %w", err)
	}
	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")

	f, err := os.Open(knownHostsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	hostStr := hostname
	if port := extractPort(remote); port != "" && port != "22" {
		hostStr = "[" + hostname + "]:" + port
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			continue
		}
		hostField := parts[0]

		// Check if hostname matches (supports comma-separated hosts)
		hostMatch := false
		for _, h := range strings.Split(hostField, ",") {
			if h == hostStr {
				hostMatch = true
				break
			}
		}
		if !hostMatch {
			continue
		}

		pubKey, _, _, _, errKey := ssh.ParseAuthorizedKey([]byte(line))
		if errKey != nil {
			continue
		}

		if string(pubKey.Marshal()) == string(key.Marshal()) {
			return true, nil
		}
		return false, fmt.Errorf("WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED for %s!", hostStr)
	}
	return false, scanner.Err()
}

func extractPort(remote net.Addr) string {
	_, port, err := net.SplitHostPort(remote.String())
	if err != nil {
		return ""
	}
	return port
}

func promptHostKey(hostname string, remote net.Addr, key ssh.PublicKey) error {
	sum := sha256.Sum256(key.Marshal())
	fingerprint := "SHA256:" + base64.StdEncoding.EncodeToString(sum[:])

	fmt.Printf("\nThe authenticity of host '%s' can't be established.\n", hostname)
	fmt.Printf("%s key fingerprint is %s.\n", key.Type(), fingerprint)
	fmt.Print("Are you sure you want to continue connecting (yes/no)? ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		resp := strings.TrimSpace(scanner.Text())
		if resp != "yes" {
			return fmt.Errorf("host key verification cancelled")
		}
	}

	return addToKnownHosts(hostname, remote, key)
}

func addToKnownHosts(hostname string, remote net.Addr, key ssh.PublicKey) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return err
	}
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	hostStr := hostname
	if port := extractPort(remote); port != "" && port != "22" {
		hostStr = "[" + hostname + "]:" + port
	}

	line := fmt.Sprintf("%s %s %s\n", hostStr, key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))

	f, err := os.OpenFile(knownHostsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(line)
	return err
}
