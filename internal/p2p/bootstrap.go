package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func ConnectToBootstrapPeers(ctx context.Context, h host.Host, addrs []string) {
	for _, s := range addrs {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			fmt.Printf("Invalid bootstrap addr %s: %s\n", s, err)
			continue
		}

		pi, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			fmt.Printf("Invalid bootstrap peer %s: %s\n", s, err)
			continue
		}

		if err := h.Connect(ctx, *pi); err != nil {
			fmt.Printf("Failed to connect to bootstrap %s: %s\n", pi.ID, err)
		} else {
			fmt.Printf("Connected to bootstrap %s\n", pi.ID)
		}
	}
}
