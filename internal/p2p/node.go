package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/quicklybly/p2p-chat/internal/config"
)

type Node struct {
	Host host.Host
}

func NewNode(ctx context.Context, cfg config.P2PConfig) (*Node, error) {
	listenAddrs := make([]multiaddr.Multiaddr, 0, len(cfg.ListenAddrs))
	for _, s := range cfg.ListenAddrs {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			return nil, fmt.Errorf("invalid listen addr %s: %w", s, err)
		}
		listenAddrs = append(listenAddrs, ma)
	}

	h, err := libp2p.New(
		libp2p.ListenAddrs(listenAddrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// mDNS
	if cfg.EnableMDNS {
		if err := SetupMDNS(h); err != nil {
			return nil, fmt.Errorf("failed to setup mDNS: %w", err)
		}
		fmt.Println("mDNS discovery enabled")
	}

	// manual bootstrap
	if len(cfg.BootstrapPeers) > 0 {
		ConnectToBootstrapPeers(ctx, h, cfg.BootstrapPeers)
	}

	return &Node{Host: h}, nil
}

func (n *Node) ID() peer.ID {
	return n.Host.ID()
}

func (n *Node) Addrs() []multiaddr.Multiaddr {
	return n.Host.Addrs()
}

func (n *Node) Stop() error {
	return n.Host.Close()
}

func (n *Node) PrintInfo() {
	fmt.Printf("PeerID: %s\n", n.Host.ID())
	fmt.Println("Listening on:")
	for _, addr := range n.Host.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, n.Host.ID())
	}
}
