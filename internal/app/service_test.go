package app

import (
	"context"
	"github.com/quicklybly/p2p-chat/internal/crypto"
	"github.com/quicklybly/p2p-chat/internal/domain"
	"sync"
	"testing"
	"time"
)

func TestCreateRoom(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	invite, err := svc.CreateRoom(ctx, "dev")
	if err != nil {
		t.Fatal(err)
	}

	// invite is not empty
	if invite == "" {
		t.Fatal("invite should not be empty")
	}

	// invite is decodable
	decoded, err := domain.DecodeInvite(invite)
	if err != nil {
		t.Fatal(err)
	}

	if decoded.Room != "dev" {
		t.Fatalf("expected room 'dev', got '%s'", decoded.Room)
	}

	// room exists
	rooms := svc.Rooms()
	if len(rooms) != 1 || rooms[0] != "dev" {
		t.Fatalf("expected [dev], got %v", rooms)
	}

	// provided to DHT
	if len(node.getProvided()) == 0 {
		t.Fatal("should have provided to DHT")
	}
}

func TestCreateRoomDuplicate(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	_, err := svc.CreateRoom(ctx, "dev")
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.CreateRoom(ctx, "dev")
	if err == nil {
		t.Fatal("should fail on duplicate room")
	}
}

func TestJoinRoom(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	// create invite manually
	secret, _ := crypto.GenerateSecret()
	invite := domain.NewInvite("dev", secret)
	encoded, _ := invite.Encode()

	roomName, err := svc.JoinRoom(ctx, encoded)
	if err != nil {
		t.Fatal(err)
	}

	if roomName != "dev" {
		t.Fatalf("expected 'dev', got '%s'", roomName)
	}

	if !svc.HasRoom("dev") {
		t.Fatal("should have room 'dev'")
	}
}

func TestJoinRoomInvalidInvite(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	_, err := svc.JoinRoom(ctx, "not-valid-invite")
	if err == nil {
		t.Fatal("should fail on invalid invite")
	}
}

func TestJoinRoomDuplicate(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	secret, _ := crypto.GenerateSecret()
	invite := domain.NewInvite("dev", secret)
	encoded, _ := invite.Encode()

	_, err := svc.JoinRoom(ctx, encoded)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.JoinRoom(ctx, encoded)
	if err == nil {
		t.Fatal("should fail on duplicate join")
	}
}

func TestSendMessage(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	_, err := svc.CreateRoom(ctx, "dev")
	if err != nil {
		t.Fatal(err)
	}

	err = svc.SendMessage(ctx, "dev", "hello")
	if err != nil {
		t.Fatal(err)
	}

	// check that something was published
	rooms := svc.Rooms()
	room, _ := svc.getRoom(rooms[0])

	published := node.getPublished(room.Topic)
	if len(published) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(published))
	}

	// published data should be encrypted
	if string(published[0]) == "hello" {
		t.Fatal("message should be encrypted")
	}

	// should be decryptable
	plaintext, err := crypto.Decrypt(room.AESKey, published[0])
	if err != nil {
		t.Fatal(err)
	}

	if string(plaintext) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", plaintext)
	}
}

func TestSendMessageNoRoom(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	err := svc.SendMessage(ctx, "dev", "hello")
	if err == nil {
		t.Fatal("should fail when not in room")
	}
}

func TestReceiveMessage(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	var received domain.Message
	var wg sync.WaitGroup
	wg.Add(1)

	svc.OnMessage(func(msg domain.Message) {
		received = msg
		wg.Done()
	})

	invite, _ := svc.CreateRoom(ctx, "dev")

	decoded, _ := domain.DecodeInvite(invite)
	secret, _ := decoded.DecodeSecret()
	keys, _ := crypto.DeriveKeys(secret)

	// simulate incoming encrypted message
	ciphertext, _ := crypto.Encrypt(keys.AESKey, []byte("hello from peer"))

	node.simulateMessage(keys.Topic, "12D3KooWTestPeer", ciphertext)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
	if received.Text != "hello from peer" {
		t.Fatalf("expected 'hello from peer', got '%s'", received.Text)
	}

	if received.RoomName != "dev" {
		t.Fatalf("expected room 'dev', got '%s'", received.RoomName)
	}

	if received.SenderID != "12D3KooWTestPeer" {
		t.Fatalf("expected sender '12D3KooWTestPeer', got '%s'", received.SenderID)
	}
}

func TestReceiveMessageWrongKey(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	messageReceived := false
	svc.OnMessage(func(msg domain.Message) {
		messageReceived = true
	})

	invite, _ := svc.CreateRoom(ctx, "dev")

	decoded, _ := domain.DecodeInvite(invite)
	secret, _ := decoded.DecodeSecret()
	keys, _ := crypto.DeriveKeys(secret)

	// encrypt with wrong key
	wrongSecret, _ := crypto.GenerateSecret()
	wrongKeys, _ := crypto.DeriveKeys(wrongSecret)
	ciphertext, _ := crypto.Encrypt(wrongKeys.AESKey, []byte("evil message"))

	node.simulateMessage(keys.Topic, "12D3KooWEvil", ciphertext)

	time.Sleep(500 * time.Millisecond)

	if messageReceived {
		t.Fatal("should not deliver message with wrong key")
	}
}

func TestLeaveRoom(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	_, err := svc.CreateRoom(ctx, "dev")
	if err != nil {
		t.Fatal(err)
	}

	if !svc.HasRoom("dev") {
		t.Fatal("should have room")
	}

	err = svc.LeaveRoom("dev")
	if err != nil {
		t.Fatal(err)
	}

	if svc.HasRoom("dev") {
		t.Fatal("should not have room after leave")
	}
}

func TestLeaveRoomNotJoined(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)

	err := svc.LeaveRoom("dev")
	if err == nil {
		t.Fatal("should fail when not in room")
	}
}

func TestMultipleRooms(t *testing.T) {
	node := newMockNode()
	svc := NewService(node)
	ctx := context.Background()

	_, err := svc.CreateRoom(ctx, "dev")
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.CreateRoom(ctx, "general")
	if err != nil {
		t.Fatal(err)
	}

	rooms := svc.Rooms()
	if len(rooms) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(rooms))
	}

	// send to each room
	err = svc.SendMessage(ctx, "dev", "hello dev")
	if err != nil {
		t.Fatal(err)
	}

	err = svc.SendMessage(ctx, "general", "hello general")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateAndJoinSameSecret(t *testing.T) {
	// two services with same secret should derive same keys
	node1 := newMockNode()
	node2 := newMockNode()
	svc1 := NewService(node1)
	svc2 := NewService(node2)
	ctx := context.Background()

	invite, _ := svc1.CreateRoom(ctx, "dev")

	_, err := svc2.JoinRoom(ctx, invite)
	if err != nil {
		t.Fatal(err)
	}

	room1, _ := svc1.getRoom("dev")
	room2, _ := svc2.getRoom("dev")

	if room1.Topic != room2.Topic {
		t.Fatal("topics should match")
	}

	if string(room1.AESKey) != string(room2.AESKey) {
		t.Fatal("AES keys should match")
	}

	if string(room1.DiscoveryKey) != string(room2.DiscoveryKey) {
		t.Fatal("discovery keys should match")
	}
}
