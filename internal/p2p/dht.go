package p2p

import (
	"context"
	"fmt"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multihash"
)

func NewDHT(ctx context.Context, h host.Host) (*dht.IpfsDHT, error) {
	kdht, err := dht.New(
		ctx,
		h,
		dht.Mode(dht.ModeServer),
		dht.ProtocolPrefix("/p2p/chat"),
	)

	if err != nil {
		return nil, err
	}

	if err := kdht.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	fmt.Println("DHT bootstrapped")
	return kdht, nil
}

func KeyToMultihash(key []byte) (string, error) {
	mh, err := multihash.Sum(key, multihash.SHA2_256, -1)

	if err != nil {
		return "", err
	}

	return mh.B58String(), nil
}
