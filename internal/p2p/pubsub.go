package p2p

import (
	"context"
	"fmt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

type PubSubService struct {
	ps     *pubsub.PubSub
	host   host.Host
	topics map[string]*pubsub.Topic
}

func NewPubSub(ctx context.Context, h host.Host) (*PubSubService, error) {
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	fmt.Println("PubSub (Gossipsub) enabled")
	return &PubSubService{
		ps:     ps,
		host:   h,
		topics: make(map[string]*pubsub.Topic),
	}, nil
}

func (p *PubSubService) Join(topic string) (*pubsub.Topic, error) {
	return p.ps.Join(topic)
}

func (p *PubSubService) Subscribe(ctx context.Context, topicName string, handler func(senderID string, data []byte)) error {
	topic, err := p.ps.Join(topicName)
	if err != nil {
		return fmt.Errorf("failed to join topic: %w", err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	p.topics[topicName] = topic

	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				return
			}
			if msg.ReceivedFrom == p.host.ID() {
				continue
			}
			handler(msg.ReceivedFrom.String(), msg.Data)
		}
	}()

	return nil
}

func (p *PubSubService) Publish(ctx context.Context, topicName string, data []byte) error {
	topic, ok := p.topics[topicName]
	if !ok {
		return fmt.Errorf("not subscribed to topic")
	}
	return topic.Publish(ctx, data)
}
