package webrtc_signaling

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
	if hub.rooms == nil {
		t.Error("rooms map is nil")
	}
	if hub.register == nil {
		t.Error("register channel is nil")
	}
	if hub.unregister == nil {
		t.Error("unregister channel is nil")
	}
	if hub.emptyRooms == nil {
		t.Error("emptyRooms channel is nil")
	}
}

func TestHubRegistrationAndRoomCreation(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create a test client with initialized send channel
	client := &Client{
		UserID: "test-user",
		RoomID: "test-room",
		send:   make(chan []byte, 256),
	}

	// Register the client
	hub.register <- client

	// Give some time for the goroutine to process
	time.Sleep(100 * time.Millisecond)

	// Check if room was created
	room, exists := hub.GetRoom("test-room")
	if !exists {
		t.Error("Room was not created")
	}
	if room == nil {
		t.Error("Room is nil")
	}
	if room.ID != "test-room" {
		t.Errorf("Room ID mismatch, got %s, want %s", room.ID, "test-room")
	}
}

func TestHubUnregistration(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create and register a test client with initialized send channel
	client := &Client{
		UserID: "test-user",
		RoomID: "test-room",
		send:   make(chan []byte, 256),
	}

	hub.register <- client
	time.Sleep(100 * time.Millisecond)

	// Verify room exists before unregistering
	_, exists := hub.GetRoom("test-room")
	if !exists {
		t.Fatal("Room not found before unregistration")
	}

	// Unregister the client
	hub.unregister <- client
	time.Sleep(100 * time.Millisecond)

	// Check if room was removed
	_, exists = hub.GetRoom("test-room")
	if exists {
		t.Error("Room was not removed after becoming empty")
	}
}

func TestEmptyRoomCleanup(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create and register a test client with initialized send channel
	client := &Client{
		UserID: "test-user",
		RoomID: "test-room",
		send:   make(chan []byte, 256),
	}

	hub.register <- client
	time.Sleep(100 * time.Millisecond)

	// Unregister the client first
	hub.unregister <- client
	time.Sleep(100 * time.Millisecond)

	// Signal that the room is empty
	hub.emptyRooms <- "test-room"
	time.Sleep(100 * time.Millisecond)

	// Check if room was removed
	_, exists := hub.GetRoom("test-room")
	if exists {
		t.Error("Empty room was not removed")
	}
}

func TestGetRoom(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Test non-existent room
	_, exists := hub.GetRoom("non-existent")
	if exists {
		t.Error("GetRoom returned true for non-existent room")
	}

	// Create a room
	client := &Client{
		UserID: "test-user",
		RoomID: "test-room",
		send:   make(chan []byte, 256),
	}
	hub.register <- client
	time.Sleep(100 * time.Millisecond)

	// Test existing room
	room, exists := hub.GetRoom("test-room")
	if !exists {
		t.Error("GetRoom returned false for existing room")
	}
	if room == nil {
		t.Error("GetRoom returned nil room for existing room")
	}
	if room.ID != "test-room" {
		t.Errorf("Room ID mismatch, got %s, want %s", room.ID, "test-room")
	}
}

func TestConcurrentRoomAccess(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create multiple clients for different rooms with initialized send channels
	clients := []*Client{
		{
			UserID: "user1",
			RoomID: "room1",
			send:   make(chan []byte, 256),
		},
		{
			UserID: "user2",
			RoomID: "room2",
			send:   make(chan []byte, 256),
		},
		{
			UserID: "user3",
			RoomID: "room1",
			send:   make(chan []byte, 256),
		},
	}

	// Register clients concurrently
	for _, client := range clients {
		go func(c *Client) {
			hub.register <- c
		}(client)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify rooms were created correctly
	room1, exists := hub.GetRoom("room1")
	if !exists {
		t.Error("Room1 was not created")
	}
	if len(room1.clients) != 2 {
		t.Errorf("Room1 has wrong number of clients, got %d, want %d", len(room1.clients), 2)
	}

	room2, exists := hub.GetRoom("room2")
	if !exists {
		t.Error("Room2 was not created")
	}
	if len(room2.clients) != 1 {
		t.Errorf("Room2 has wrong number of clients, got %d, want %d", len(room2.clients), 1)
	}
}
