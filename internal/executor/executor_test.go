package executor

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sshgo/sshgo/internal/config"
	"golang.org/x/crypto/ssh"
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

func TestConnectViaJumpDialsTargetThroughJumpHost(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	jumpClientKey, jumpClientSigner := newTestKey(t)
	jumpIdentityFile := writePrivateKey(t, t.TempDir(), jumpClientKey)

	target := startTestSSHServer(t, testSSHServerConfig{
		passwordUser: "deploy",
		password:     "secret",
	})
	jump := startTestSSHServer(t, testSSHServerConfig{
		publicKeyUser: "jumpuser",
		publicKey:     jumpClientSigner.PublicKey(),
		forwardTarget: target.addr,
	})
	writeKnownHost(t, home, jump.knownHostsName, jump.hostKey.PublicKey())

	targetAddr := fmt.Sprintf("127.0.0.1:%d", unusedTCPPort(t))
	jumpHost, jumpPort := splitHostPort(t, jump.addr)
	profile := config.Profile{
		Name: "target",
		Host: "127.0.0.1",
		Port: jumpPortForTarget(t, targetAddr),
		User: "deploy",
		JumpHosts: []config.JumpHost{{
			Name:         "bastion",
			Host:         jumpHost,
			Port:         jumpPort,
			User:         "jumpuser",
			IdentityFile: jumpIdentityFile,
		}},
	}

	sshConf := &ssh.ClientConfig{
		User:            "deploy",
		Auth:            []ssh.AuthMethod{ssh.Password("secret")},
		HostKeyCallback: ssh.FixedHostKey(target.hostKey.PublicKey()),
		Timeout:         time.Second,
	}

	client, err := connectViaJump("secret", profile, targetAddr, sshConf)
	if err != nil {
		t.Fatalf("connectViaJump failed: %v", err)
	}
	client.Close()

	if got := jump.directTCPIPCount.Load(); got != 1 {
		t.Fatalf("expected one direct-tcpip request through jump host, got %d", got)
	}
	if got := jump.lastForwardAddr.Load(); got != targetAddr {
		t.Fatalf("expected jump host to receive target addr %q, got %v", targetAddr, got)
	}
}

func TestConnectViaJumpAuthenticatesJumpHostWithPassword(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	target := startTestSSHServer(t, testSSHServerConfig{
		passwordUser: "deploy",
		password:     "target-secret",
	})
	jump := startTestSSHServer(t, testSSHServerConfig{
		passwordUser:  "liheng",
		password:      "jump-secret",
		forwardTarget: target.addr,
	})
	writeKnownHost(t, home, jump.knownHostsName, jump.hostKey.PublicKey())

	targetAddr := fmt.Sprintf("127.0.0.1:%d", unusedTCPPort(t))
	jumpHost, jumpPort := splitHostPort(t, jump.addr)
	profile := config.Profile{
		Name: "target",
		Host: "127.0.0.1",
		Port: jumpPortForTarget(t, targetAddr),
		User: "deploy",
		JumpHosts: []config.JumpHost{{
			Name:     "dev-jumper-hk",
			Host:     jumpHost,
			Port:     jumpPort,
			User:     "liheng",
			Password: "jump-secret",
		}},
	}

	sshConf := &ssh.ClientConfig{
		User:            "deploy",
		Auth:            []ssh.AuthMethod{ssh.Password("target-secret")},
		HostKeyCallback: ssh.FixedHostKey(target.hostKey.PublicKey()),
		Timeout:         time.Second,
	}

	client, err := connectViaJump("target-secret", profile, targetAddr, sshConf)
	if err != nil {
		t.Fatalf("connectViaJump failed: %v", err)
	}
	client.Close()

	if got := jump.directTCPIPCount.Load(); got != 1 {
		t.Fatalf("expected one direct-tcpip request through jump host, got %d", got)
	}
}

type testSSHServerConfig struct {
	passwordUser  string
	password      string
	publicKeyUser string
	publicKey     ssh.PublicKey
	forwardTarget string
}

type testSSHServer struct {
	addr             string
	hostKey          ssh.Signer
	knownHostsName   string
	directTCPIPCount atomic.Int32
	lastForwardAddr  atomic.Value
}

type directTCPIPMsg struct {
	Raddr string
	Rport uint32
	Laddr string
	Lport uint32
}

func startTestSSHServer(t *testing.T, cfg testSSHServerConfig) *testSSHServer {
	t.Helper()

	hostKey := newTestSigner(t)
	server := &testSSHServer{hostKey: hostKey}

	sshConfig := &ssh.ServerConfig{}
	sshConfig.AddHostKey(hostKey)
	if cfg.passwordUser != "" {
		sshConfig.PasswordCallback = func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if conn.User() == cfg.passwordUser && string(password) == cfg.password {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected")
		}
	}
	if cfg.publicKeyUser != "" {
		sshConfig.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if conn.User() == cfg.publicKeyUser && bytes.Equal(key.Marshal(), cfg.publicKey.Marshal()) {
				return nil, nil
			}
			return nil, fmt.Errorf("public key rejected")
		}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen test ssh server: %v", err)
	}
	t.Cleanup(func() { listener.Close() })

	server.addr = listener.Addr().String()
	server.knownHostsName = knownHostsNameForTest(server.addr)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConn(conn, sshConfig, cfg.forwardTarget)
		}
	}()

	return server
}

func (s *testSSHServer) handleConn(conn net.Conn, cfg *ssh.ServerConfig, forwardTarget string) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, cfg)
	if err != nil {
		conn.Close()
		return
	}
	go ssh.DiscardRequests(reqs)
	go func() {
		sshConn.Wait()
		conn.Close()
	}()

	for newChannel := range chans {
		if newChannel.ChannelType() != "direct-tcpip" || forwardTarget == "" {
			newChannel.Reject(ssh.UnknownChannelType, "unsupported channel")
			continue
		}

		var msg directTCPIPMsg
		if err := ssh.Unmarshal(newChannel.ExtraData(), &msg); err != nil {
			newChannel.Reject(ssh.ConnectionFailed, err.Error())
			continue
		}
		s.directTCPIPCount.Add(1)
		requestedAddr := net.JoinHostPort(msg.Raddr, strconv.Itoa(int(msg.Rport)))
		s.lastForwardAddr.Store(requestedAddr)

		upstream, err := net.Dial("tcp", forwardTarget)
		if err != nil {
			newChannel.Reject(ssh.ConnectionFailed, err.Error())
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			upstream.Close()
			continue
		}
		go ssh.DiscardRequests(requests)
		go proxyConn(channel, upstream)
		go proxyConn(upstream, channel)
	}
}

func proxyConn(dst io.WriteCloser, src io.ReadCloser) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

func newTestKey(t *testing.T) (*rsa.PrivateKey, ssh.Signer) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatalf("create ssh signer: %v", err)
	}
	return key, signer
}

func newTestSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, signer := newTestKey(t)
	return signer
}

func writePrivateKey(t *testing.T, dir string, privateKey *rsa.PrivateKey) string {
	t.Helper()

	path := filepath.Join(dir, "id_rsa")
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
	return path
}

func writeKnownHost(t *testing.T, home, host string, key ssh.PublicKey) {
	t.Helper()

	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("mkdir known_hosts dir: %v", err)
	}
	line := fmt.Sprintf("%s %s", host, ssh.MarshalAuthorizedKey(key))
	if err := os.WriteFile(filepath.Join(sshDir, "known_hosts"), []byte(line), 0600); err != nil {
		t.Fatalf("write known_hosts: %v", err)
	}
}

func knownHostsNameForTest(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return fmt.Sprintf("[%s:%s]:%s", host, port, port)
}

func unusedTCPPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate unused port: %v", err)
	}
	_, port := splitHostPort(t, listener.Addr().String())
	listener.Close()
	return port
}

func splitHostPort(t *testing.T, addr string) (string, int) {
	t.Helper()

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host port %q: %v", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse port %q: %v", portStr, err)
	}
	return host, port
}

func jumpPortForTarget(t *testing.T, addr string) int {
	t.Helper()
	_, port := splitHostPort(t, addr)
	return port
}
