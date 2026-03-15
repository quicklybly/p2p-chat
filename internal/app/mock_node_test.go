package app

import (
	"context"
	"github.com/libp2p/go-libp2p/core/peer"
	"sync"
)

type mockNode struct {
	mu            sync.Mutex
	provided      [][]byte
	published     map[string][][]byte
	subscriptions map[string]func(senderID string, data []byte)
	providers     map[string][]peer.AddrInfo
	provideErr    error
}

func newMockNode() *mockNode {
	return &mockNode{
		published:     make(map[string][][]byte),
		subscriptions: make(map[string]func(senderID string, data []byte)),
		providers:     make(map[string][]peer.AddrInfo),
	}
}

func (m *mockNode) Provide(_ context.Context, key []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.provideErr != nil {
		return m.provideErr
	}

	m.provided = append(m.provided, key)
	return nil
}

func (m *mockNode) FindProviders(_ context.Context, key []byte) ([]peer.AddrInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	k := string(key)
	return m.providers[k], nil
}

func (m *mockNode) ConnectToPeer(_ context.Context, _ peer.AddrInfo) error {
	return nil
}

func (m *mockNode) Subscribe(_ context.Context, topic string, handler func(senderID string, data []byte)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscriptions[topic] = handler
	return nil
}

func (m *mockNode) Publish(_ context.Context, topic string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.published[topic] = append(m.published[topic], data)
	return nil
}

func (m *mockNode) simulateMessage(topic string, senderID string, data []byte) {
	m.mu.Lock()
	handler, ok := m.subscriptions[topic]
	m.mu.Unlock()

	if ok {
		handler(senderID, data)
	}
}

func (m *mockNode) getProvided() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.provided
}

func (m *mockNode) getPublished(topic string) [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.published[topic]
}
