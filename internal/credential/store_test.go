package credential

import "testing"

func TestKeyFormat(t *testing.T) {
	tests := []struct {
		kind, name, want string
	}{
		{KindPassword, "my-server", "sshgo:password:my-server"},
		{KindKeyPassphrase, "bastion", "sshgo:key-passphrase:bastion"},
	}
	for _, tt := range tests {
		got := Key(tt.kind, tt.name)
		if got != tt.want {
			t.Errorf("Key(%q, %q) = %q, want %q", tt.kind, tt.name, got, tt.want)
		}
	}
}