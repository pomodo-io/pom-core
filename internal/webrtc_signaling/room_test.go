package webrtc_signaling

import (
	"testing"
	"time"
)

func TestNewRoom(t *testing.T) {
	hub := &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	roomID := "test-room"
	room := newRoom(roomID, hub)

	if room.ID != roomID {
		t.Errorf("Expected room ID %s, got %s", roomID, room.ID)
	}
	if room.hub != hub {
		t.Error("Room hub not properly set")
	}
	if len(room.clients) != 0 {
		t.Error("Expected empty clients map")
	}
}

func TestRoomRegistration(t *testing.T) {
	hub := &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	room := newRoom("test-room", hub)
	go room.run()

	// Create a test client
	client := &Client{
		UserID: "test-user",
		send:   make(chan []byte, 256),
	}

	// Register client
	room.register <- client
	time.Sleep(100 * time.Millisecond) // Give time for goroutine to process

	if len(room.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(room.clients))
	}

	if !room.clients[client] {
		t.Error("Client not properly registered")
	}

	// Test duplicate registration
	room.register <- client
	time.Sleep(100 * time.Millisecond)

	if len(room.clients) != 1 {
		t.Errorf("Expected still 1 client after duplicate registration, got %d", len(room.clients))
	}
}

func TestRoomUnregistration(t *testing.T) {
	hub := &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		emptyRooms: make(chan string),
	}
	room := newRoom("test-room", hub)
	go room.run()

	client := &Client{
		UserID: "test-user",
		send:   make(chan []byte, 256),
	}

	// Register then unregister
	room.register <- client
	time.Sleep(100 * time.Millisecond)

	room.unregister <- client
	time.Sleep(100 * time.Millisecond)

	if len(room.clients) != 0 {
		t.Errorf("Expected 0 clients after unregistration, got %d", len(room.clients))
	}
}

func TestRoomBroadcast(t *testing.T) {
	hub := &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	room := newRoom("test-room", hub)
	go room.run()

	// Create two test clients
	client1 := &Client{
		UserID: "user1",
		send:   make(chan []byte, 256),
	}
	client2 := &Client{
		UserID: "user2",
		send:   make(chan []byte, 256),
	}

	// Register both clients
	room.register <- client1
	room.register <- client2
	time.Sleep(100 * time.Millisecond)

	// Create a test message
	message := &WebSocketMessage{
		Type:         "test_message",
		Payload:      "Hello, World!",
		RoomID:       room.ID,
		SenderUserID: "user1",
		Timestamp:    time.Now().UnixMilli(),
	}

	// Broadcast message
	room.broadcast <- message
	time.Sleep(100 * time.Millisecond)

	// Check if both clients received the message
	select {
	case msg1 := <-client1.send:
		t.Logf("Client 1 received message: %s", string(msg1))
	default:
		t.Error("Client 1 did not receive message")
	}

	select {
	case msg2 := <-client2.send:
		t.Logf("Client 2 received message: %s", string(msg2))
	default:
		t.Error("Client 2 did not receive message")
	}
}

func TestRoomBroadcastExclude(t *testing.T) {
	hub := &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	room := newRoom("test-room", hub)
	go room.run()

	// Create two test clients
	client1 := &Client{
		UserID: "user1",
		send:   make(chan []byte, 256),
	}
	client2 := &Client{
		UserID: "user2",
		send:   make(chan []byte, 256),
	}

	// Register both clients
	room.register <- client1
	room.register <- client2
	time.Sleep(100 * time.Millisecond)

	// Drain any system messages from the channels
	drainMessages := func(ch chan []byte) {
		for {
			select {
			case <-ch:
			default:
				return
			}
		}
	}
	drainMessages(client1.send)
	drainMessages(client2.send)

	// Create a test message
	message := &WebSocketMessage{
		Type:         "test_message",
		Payload:      "Hello, World!",
		RoomID:       room.ID,
		SenderUserID: "user1",
		Timestamp:    time.Now().UnixMilli(),
	}

	// Broadcast message excluding client1
	room.broadcastToRoom(message, client1)
	time.Sleep(100 * time.Millisecond)

	// Check if client1 did NOT receive the message
	select {
	case msg := <-client1.send:
		t.Errorf("Client 1 should not have received message, but got: %s", string(msg))
	default:
		// This is good - client1 should not receive the message
	}

	// Check if client2 received the message
	select {
	case msg2 := <-client2.send:
		t.Logf("Client 2 received message: %s", string(msg2))
	default:
		t.Error("Client 2 did not receive message")
	}
}
