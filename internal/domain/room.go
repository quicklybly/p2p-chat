package domain

import "encoding/hex"

type Room struct {
	Name         string
	Secret       []byte
	DiscoveryKey []byte
	Topic        string
	AESKey       []byte
}

func (r *Room) ID() string {
	return hex.EncodeToString(r.DiscoveryKey[:8])
}
