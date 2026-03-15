package app

import (
	"context"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Node interface {
	Provide(ctx context.Context, key []byte) error
	FindProviders(ctx context.Context, key []byte) ([]peer.AddrInfo, error)
	ConnectToPeer(ctx context.Context, pi peer.AddrInfo) error

	Subscribe(ctx context.Context, topic string, handler func(senderID string, data []byte)) error
	Publish(ctx context.Context, topic string, data []byte) error
}
