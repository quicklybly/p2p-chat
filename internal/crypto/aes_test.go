package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	secret, _ := GenerateSecret()
	keys, _ := DeriveKeys(secret)

	plaintext := []byte("hello p2p world")

	ciphertext, err := Encrypt(keys.AESKey, plaintext)
	if err != nil {
		t.Fatal(err)
	}

	// ciphertext should differ from plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext should not equal plaintext")
	}

	// decrypt
	result, err := Decrypt(keys.AESKey, ciphertext)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, plaintext) {
		t.Fatalf("expected %s, got %s", plaintext, result)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	s1, _ := GenerateSecret()
	s2, _ := GenerateSecret()
	k1, _ := DeriveKeys(s1)
	k2, _ := DeriveKeys(s2)

	plaintext := []byte("secret message")

	ciphertext, _ := Encrypt(k1.AESKey, plaintext)

	_, err := Decrypt(k2.AESKey, ciphertext)
	if err == nil {
		t.Fatal("should fail with wrong key")
	}
}

func TestEncryptDifferentNonce(t *testing.T) {
	secret, _ := GenerateSecret()
	keys, _ := DeriveKeys(secret)

	plaintext := []byte("same message")

	c1, _ := Encrypt(keys.AESKey, plaintext)
	c2, _ := Encrypt(keys.AESKey, plaintext)

	if bytes.Equal(c1, c2) {
		t.Fatal("same plaintext should produce different ciphertext")
	}
}
