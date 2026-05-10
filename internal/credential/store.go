package credential

import (
	"fmt"
	"github.com/zalando/go-keyring"
)

const keyPrefix = "sshgo"
const (
	KindPassword     = "password"
	KindKeyPassphrase = "key-passphrase"
)

func Key(kind, name string) string {
	return fmt.Sprintf("%s:%s:%s", keyPrefix, kind, name)
}
func Set(kind, name, value string) error {
	return keyring.Set(keyPrefix, Key(kind, name), value)
}
func Get(kind, name string) (string, error) {
	val, err := keyring.Get(keyPrefix, Key(kind, name))
	if err != nil {
		return "", err
	}
	return val, nil
}
func Delete(kind, name string) error {
	return keyring.Delete(keyPrefix, Key(kind, name))
}