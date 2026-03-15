package app

import (
	"context"
	"fmt"
	"github.com/quicklybly/p2p-chat/internal/crypto"
	"github.com/quicklybly/p2p-chat/internal/domain"
	"sync"
	"time"
)

type MessageHandler func(msg domain.Message)

type Service struct {
	node    Node
	rooms   map[string]*domain.Room
	names   map[string]string
	handler MessageHandler
	mutex   sync.RWMutex
}

func NewService(node Node) *Service {
	return &Service{
		node:  node,
		rooms: make(map[string]*domain.Room),
		names: make(map[string]string),
	}
}

func (s *Service) OnMessage(handler MessageHandler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.handler = handler
}

func (s *Service) CreateRoom(ctx context.Context, name string) (string, error) {
	if _, exists := s.getRoom(name); exists {
		return "", fmt.Errorf("room '%s' already exists", name)
	}

	secret, err := crypto.GenerateSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}

	room, err := s.setupRoom(ctx, name, secret)
	if err != nil {
		return "", err
	}

	if err := s.addRoom(room); err != nil {
		return "", err
	}

	invite := domain.NewInvite(name, secret)
	encodedInvite, err := invite.Encode()
	if err != nil {
		return "", fmt.Errorf("failed to encode invite: %w", err)
	}

	return encodedInvite, nil
}

func (s *Service) JoinRoom(ctx context.Context, inviteStr string) (string, error) {
	invite, err := domain.DecodeInvite(inviteStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode invite: %w", err)
	}

	if _, exists := s.getRoom(invite.Room); exists {
		return "", fmt.Errorf("room '%s' already exists", invite.Room)
	}

	secret, err := invite.DecodeSecret()
	if err != nil {
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}

	room, err := s.setupRoom(ctx, invite.Room, secret)
	if err != nil {
		return "", err
	}

	peers, err := s.node.FindProviders(ctx, room.DiscoveryKey)
	if err != nil {
		fmt.Printf("Warning: find providers: %s\n", err)
	}

	connected := 0
	for _, p := range peers {
		if err := s.node.ConnectToPeer(ctx, p); err != nil {
			fmt.Printf("Warning: connect to %s: %s\n", p.ID, err)
		} else {
			connected++
		}
	}

	if err := s.addRoom(room); err != nil {
		return "", err
	}

	fmt.Printf("Joined room '%s' (found %d peers, connected to %d)\n", invite.Room, len(peers), connected)

	return invite.Room, nil
}

func (s *Service) SendMessage(ctx context.Context, roomName string, text string) error {
	room, ok := s.getRoom(roomName)
	if !ok {
		return fmt.Errorf("room '%s' not found", roomName)
	}

	ciphertext, err := crypto.Encrypt(room.AESKey, []byte(text))
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %w", err)
	}

	return s.node.Publish(ctx, room.Topic, ciphertext)
}

func (s *Service) LeaveRoom(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	topic, ok := s.names[name]
	if !ok {
		return fmt.Errorf("not in room '%s'", name)
	}

	delete(s.rooms, topic)
	delete(s.names, name)
	return nil
}

func (s *Service) Rooms() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	names := make([]string, 0, len(s.names))
	for name := range s.names {
		names = append(names, name)
	}
	return names
}

func (s *Service) HasRoom(name string) bool {
	_, ok := s.getRoom(name)
	return ok
}

func (s *Service) setupRoom(ctx context.Context, name string, secret []byte) (*domain.Room, error) {
	keys, err := crypto.DeriveKeys(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to derive keys: %w", err)
	}

	room := &domain.Room{
		Name:         name,
		Secret:       secret,
		DiscoveryKey: keys.DiscoveryKey,
		Topic:        keys.Topic,
		AESKey:       keys.AESKey,
	}

	// provide to dht in background
	go s.provideLoop(ctx, room)

	// subscribe to topic
	if err := s.node.Subscribe(ctx, room.Topic, s.makeTopicHandler(room)); err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	return room, nil
}

func (s *Service) provideLoop(ctx context.Context, room *domain.Room) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := s.node.Provide(ctx, room.DiscoveryKey)
			if err == nil {
				fmt.Printf("Room '%s' provided to DHT\n", room.Name)
				return
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func (s *Service) makeTopicHandler(room *domain.Room) func(senderID string, data []byte) {
	return func(senderID string, data []byte) {
		plaintext, err := crypto.Decrypt(room.AESKey, data)
		if err != nil {
			return // wrong key
		}

		s.mutex.RLock()
		handler := s.handler
		s.mutex.RUnlock()

		if handler != nil {
			handler(domain.Message{
				SenderID:  senderID,
				RoomName:  room.Name,
				Text:      string(plaintext),
				Timestamp: time.Now(),
			})
		}
	}
}

func (s *Service) addRoom(room *domain.Room) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.names[room.Name]; exists {
		return fmt.Errorf("room '%s' already exists", room.Name)
	}

	s.rooms[room.Topic] = room
	s.names[room.Name] = room.Topic
	return nil
}

func (s *Service) getRoom(name string) (*domain.Room, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	topic, ok := s.names[name]
	if !ok {
		return nil, false
	}

	room, ok := s.rooms[topic]
	return room, ok
}
