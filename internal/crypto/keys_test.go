package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateSecret(t *testing.T) {
	s1, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}

	s2, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}

	if len(s1) != MasterSecretLength {
		t.Fatalf("expected %d bytes, got %d", MasterSecretLength, len(s1))
	}

	if bytes.Equal(s1, s2) {
		t.Fatal("two secrets should not be equal")
	}
}

func TestDeriveKeys(t *testing.T) {
	secret, _ := GenerateSecret()

	k1, err := DeriveKeys(secret)
	if err != nil {
		t.Fatal(err)
	}

	k2, err := DeriveKeys(secret)
	if err != nil {
		t.Fatal(err)
	}

	// same secret should produce same keys
	if !bytes.Equal(k1.DiscoveryKey, k2.DiscoveryKey) {
		t.Fatal("discovery keys should match")
	}
	if k1.Topic != k2.Topic {
		t.Fatal("topics should match")
	}
	if !bytes.Equal(k1.AESKey, k2.AESKey) {
		t.Fatal("AES keys should match")
	}

	// different keys from each other
	if bytes.Equal(k1.DiscoveryKey, k1.AESKey) {
		t.Fatal("discovery key and AES key should differ")
	}
}

func TestDeriveKeysDifferentSecrets(t *testing.T) {
	s1, _ := GenerateSecret()
	s2, _ := GenerateSecret()

	k1, _ := DeriveKeys(s1)
	k2, _ := DeriveKeys(s2)

	if bytes.Equal(k1.DiscoveryKey, k2.DiscoveryKey) {
		t.Fatal("different secrets should produce different keys")
	}
}
