package domain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Invite struct {
	Room   string `json:"room"`
	Secret string `json:"secret"`
}

func NewInvite(room string, secret []byte) *Invite {
	return &Invite{
		Room:   room,
		Secret: base64.StdEncoding.EncodeToString(secret),
	}
}

func (i *Invite) Encode() (string, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return "", fmt.Errorf("failed to encode invite: %w", err)
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

func (i *Invite) DecodeSecret() ([]byte, error) {
	secret, err := base64.StdEncoding.DecodeString(i.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode secret: %w", err)
	}
	return secret, nil
}

func DecodeInvite(encoded string) (*Invite, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid invite: %w", err)
	}

	var invite Invite
	if err := json.Unmarshal(data, &invite); err != nil {
		return nil, fmt.Errorf("invalid invite format: %w", err)
	}

	if invite.Room == "" || invite.Secret == "" {
		return nil, fmt.Errorf("invite missing room or secret")
	}

	return &invite, nil
}
