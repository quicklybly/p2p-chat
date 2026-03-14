package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const mdnsServiceTag = "p2p-chat"

type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return
	}
	fmt.Printf("Discovered peer: %s\n", pi.ID)
	if err := n.h.Connect(context.Background(), pi); err != nil {
		fmt.Printf("Failed to connect to %s: %s\n", pi.ID, err)
	} else {
		fmt.Printf("Connected to %s\n", pi.ID)
	}
}

func SetupMDNS(h host.Host) error {
	s := mdns.NewMdnsService(h, mdnsServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}
