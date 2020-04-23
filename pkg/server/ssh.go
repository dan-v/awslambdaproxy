package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"os/user"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type sshManager struct {
	privateKey *rsa.PrivateKey
}

func (s *sshManager) getPrivateKeyBytes() []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(s.privateKey),
		},
	)
}

func (s *sshManager) getPublicKeyBytes() []byte {
	publicKey, _ := ssh.NewPublicKey(&s.privateKey.PublicKey)
	return ssh.MarshalAuthorizedKey(publicKey)
}

func (s *sshManager) getPublicKeyString() string {
	return strings.Trim(string(s.getPublicKeyBytes()[:]), "\n")
}

func (s *sshManager) insertAuthorizedKey() error {
	f, err := os.OpenFile(s.getAuthorizedKeysFile(), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "Failed to open authorized_keys file")
	}
	defer f.Close()

	if _, err = f.Write(s.getPublicKeyBytes()); err != nil {
		errors.Wrap(err, "Failed to write authorized_keys file")
	}
	return nil
}

func (s *sshManager) removeAuthorizedKey() error {
	authorizedKeysBytes, err := ioutil.ReadFile(s.getAuthorizedKeysFile())
	if err != nil {
		errors.Wrap(err, "Failed to read authorized_keys file")
	}

	lines := strings.Split(string(authorizedKeysBytes), "\n")

	for i, line := range lines {
		if line == s.getPublicKeyString() {
			log.Println("Removed entry to authorized_keys")
			lines[i] = ""
		}
	}

	output := strings.Join(lines, "\n")
	outputClean := strings.Replace(output, "\n\n", "\n", -1)
	err = ioutil.WriteFile(s.getAuthorizedKeysFile(), []byte(outputClean), 0644)
	if err != nil {
		errors.Wrap(err, "Failed to write authorized_keys file")
	}

	return nil
}

func (s *sshManager) getAuthorizedKeysFile() string {
	usr, _ := user.Current()
	return usr.HomeDir + "/.ssh/authorized_keys"
}

// NewSSHManager generates an ssh key and adds to authorized_keys so Lambda can connect to the host
func NewSSHManager() ([]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrap(err, "Error generating private SSH key")
	}
	s := &sshManager{
		privateKey: privateKey,
	}
	log.Println("Generated SSH key: ", s.getPublicKeyString())

	err = s.insertAuthorizedKey()
	if err != nil {
		return nil, errors.Wrap(err, "Error adding authorized key")
	}
	log.Println("Added entry to authorized keys file ", s.getAuthorizedKeysFile())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Println("Shutting down due to ", sig.String())
			log.Println("Cleaning up authorized_key file ", s.getAuthorizedKeysFile())
			s.removeAuthorizedKey()
			os.Exit(0)
		}
	}()

	return s.getPrivateKeyBytes(), nil
}
