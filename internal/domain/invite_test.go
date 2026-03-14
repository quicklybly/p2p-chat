package domain

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestInviteEncodeDecode(t *testing.T) {
	secret := make([]byte, 32)
	_, _ = rand.Read(secret)

	invite := NewInvite("dev", secret)

	encoded, err := invite.Encode()
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeInvite(encoded)
	if err != nil {
		t.Fatal(err)
	}

	if decoded.Room != "dev" {
		t.Fatalf("expected room 'dev', got '%s'", decoded.Room)
	}

	decodedSecret, err := decoded.DecodeSecret()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(secret, decodedSecret) {
		t.Fatal("secrets should match")
	}
}

func TestDecodeInvalidInvite(t *testing.T) {
	_, err := DecodeInvite("not-valid-base64!")
	if err == nil {
		t.Fatal("should fail on invalid input")
	}
}

func TestDecodeEmptyFields(t *testing.T) {
	_, err := DecodeInvite("e30=") // base64 of "{}"
	if err == nil {
		t.Fatal("should fail on empty fields")
	}
}
