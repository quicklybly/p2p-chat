package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/hkdf"
	"io"
)

type Keys struct {
	DiscoveryKey []byte
	Topic        string
	AESKey       []byte
}

func GenerateSecret() ([]byte, error) {
	secret := make([]byte, MasterSecretLength)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}
	return secret, nil
}

func DeriveKeys(secret []byte) (*Keys, error) {
	discoveryKey, err := deriveKey(secret, LabelDiscovery, MasterSecretLength)
	if err != nil {
		return nil, fmt.Errorf("failed to derive discovery key: %w", err)
	}

	topicBytes, err := deriveKey(secret, LabelTopic, MasterSecretLength)
	if err != nil {
		return nil, fmt.Errorf("failed to derive topic key: %w", err)
	}

	aesKey, err := deriveKey(secret, LabelAES, AESKeyLength)
	if err != nil {
		return nil, fmt.Errorf("failed to derive AES key: %w", err)
	}

	return &Keys{
		DiscoveryKey: discoveryKey,
		Topic:        hex.EncodeToString(topicBytes),
		AESKey:       aesKey,
	}, nil
}

func deriveKey(secret []byte, label string, length int) ([]byte, error) {
	hkdfReader := hkdf.New(sha256.New, secret, nil, []byte(label))

	key := make([]byte, length)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, err
	}

	return key, nil
}
