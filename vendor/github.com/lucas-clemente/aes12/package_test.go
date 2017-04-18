package aes12_test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"

	"github.com/lucas-clemente/aes12"
)

const plaintextLen = 1000

var (
	key       []byte
	nonce     []byte
	aad       []byte
	plaintext []byte
)

func init() {
	key = make([]byte, 32)
	rand.Read(key)
	nonce = make([]byte, 12)
	rand.Read(nonce)
	aad = make([]byte, 42)
	rand.Read(aad)
	plaintext = make([]byte, plaintextLen)
	rand.Read(plaintext)
}

func TestEncryption(t *testing.T) {
	c, err := aes12.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	gcm, err := aes12.NewGCM(c)
	if err != nil {
		t.Fatal(err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, aad)

	if len(ciphertext) != plaintextLen+12 {
		t.Fatal("expected ciphertext to have len(plaintext)+12")
	}

	// Test that it matches the stdlib
	stdC, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	stdGcm, err := cipher.NewGCM(stdC)
	if err != nil {
		t.Fatal(err)
	}
	stdCiphertext := stdGcm.Seal(nil, nonce, plaintext, aad)
	if !bytes.Equal(ciphertext, stdCiphertext[:len(stdCiphertext)-4]) {
		t.Fatal("did not match stdlib's ciphertext")
	}

	decrypted, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatal("decryption yielded unexpected result")
	}
}

func TestInplaceEncryption(t *testing.T) {
	c, err := aes12.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	gcm, err := aes12.NewGCM(c)
	if err != nil {
		t.Fatal(err)
	}

	buffer := make([]byte, len(plaintext), len(plaintext)+12)
	copy(buffer, plaintext)

	ciphertext := gcm.Seal(buffer[:0], nonce, buffer, aad)

	if len(ciphertext) != plaintextLen+12 {
		t.Fatal("expected ciphertext to have len(plaintext)+12")
	}
	buffer = buffer[:len(plaintext)+12]
	if !bytes.Equal(ciphertext, buffer) {
		t.Fatal("ciphertext != buffer")
	}

	// Test that it matches the stdlib
	stdC, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	stdGcm, err := cipher.NewGCM(stdC)
	if err != nil {
		t.Fatal(err)
	}
	stdCiphertext := stdGcm.Seal(nil, nonce, plaintext, aad)
	if !bytes.Equal(ciphertext, stdCiphertext[:len(stdCiphertext)-4]) {
		t.Fatal("did not match stdlib's ciphertext")
	}

	decrypted, err := gcm.Open(buffer[:0], nonce, buffer, aad)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatal("decryption yielded unexpected result")
	}
}
