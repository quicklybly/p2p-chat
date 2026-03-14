package p2p

import (
	"context"
	"fmt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

type PubSubService struct {
	ps *pubsub.PubSub
}

func NewPubSub(ctx context.Context, h host.Host) (*PubSubService, error) {
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	fmt.Println("Gossip sub enabled")

	return &PubSubService{ps: ps}, nil
}

func (p *PubSubService) Join(topic string) (*pubsub.Topic, error) {
	return p.ps.Join(topic)
}

func (p *PubSubService) Subscribe(topic *pubsub.Topic) (*pubsub.Subscription, error) {
	return topic.Subscribe()
}

func (p *PubSubService) Publish(ctx context.Context, topic *pubsub.Topic, data []byte) error {
	return topic.Publish(ctx, data)
}
