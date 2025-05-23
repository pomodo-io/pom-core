package webrtc_signaling

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For testing purposes
	},
}

// setupTestServer creates a test server with WebSocket endpoint
func setupTestServer(t *testing.T) (*httptest.Server, *Hub) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
		}

		// Extract userID and roomID from query parameters for testing
		userID := r.URL.Query().Get("userID")
		roomID := r.URL.Query().Get("roomID")

		client := &Client{
			hub:    hub,
			conn:   conn,
			send:   make(chan []byte, 256),
			UserID: userID,
			RoomID: roomID,
		}

		hub.register <- client

		go client.writePump()
		go client.readPump()
	}))

	return server, hub
}

// connectTestClient creates a WebSocket connection to the test server
func connectTestClient(t *testing.T, server *httptest.Server, userID, roomID string) *websocket.Conn {
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "?userID=" + userID + "&roomID=" + roomID
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
	return conn
}

func TestClientConnection(t *testing.T) {
	server, hub := setupTestServer(t)
	defer server.Close()

	// Test basic connection
	conn := connectTestClient(t, server, "user1", "room1")
	defer conn.Close()

	// Wait for client to be registered
	time.Sleep(100 * time.Millisecond)

	// Verify room was created and client was registered
	room, exists := hub.GetRoom("room1")
	assert.True(t, exists, "Room should exist")
	assert.Equal(t, 1, len(room.clients), "Room should have one client")
}

func TestClientMessageHandling(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Connect two clients to the same room
	client1 := connectTestClient(t, server, "user1", "room1")
	client2 := connectTestClient(t, server, "user2", "room1")
	defer client1.Close()
	defer client2.Close()

	// Wait for clients to be registered
	time.Sleep(100 * time.Millisecond)

	// Test chat message
	chatMsg := WebSocketMessage{
		Type: "chat",
		Payload: ChatMessagePayload{
			Content:         "Hello, World!",
			UserDisplayName: "User1",
		},
	}
	msgBytes, _ := json.Marshal(chatMsg)
	err := client1.WriteMessage(websocket.TextMessage, msgBytes)
	assert.NoError(t, err, "Should send chat message successfully")

	// Read message on client2
	_, message, err := client2.ReadMessage()
	assert.NoError(t, err, "Should receive message")

	var receivedMsg WebSocketMessage
	err = json.Unmarshal(message, &receivedMsg)
	assert.NoError(t, err, "Should unmarshal message")
	assert.Equal(t, "chat", receivedMsg.Type, "Message type should be chat")
	assert.Equal(t, "user1", receivedMsg.SenderUserID, "Sender should be user1")
}

func TestWebRTCSignaling(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Connect two clients
	client1 := connectTestClient(t, server, "user1", "room1")
	client2 := connectTestClient(t, server, "user2", "room1")
	defer client1.Close()
	defer client2.Close()

	// Wait for clients to be registered
	time.Sleep(100 * time.Millisecond)

	// Test WebRTC offer
	offerMsg := WebSocketMessage{
		Type: "webrtc_offer",
		Payload: WebRTCSignalPayload{
			SignalType:   "offer",
			SDP:          "test-sdp",
			TargetUserID: "user2",
		},
	}
	msgBytes, _ := json.Marshal(offerMsg)
	err := client1.WriteMessage(websocket.TextMessage, msgBytes)
	assert.NoError(t, err, "Should send WebRTC offer successfully")

	// Read offer on client2
	_, message, err := client2.ReadMessage()
	assert.NoError(t, err, "Should receive WebRTC offer")

	var receivedMsg WebSocketMessage
	err = json.Unmarshal(message, &receivedMsg)
	assert.NoError(t, err, "Should unmarshal WebRTC offer")
	assert.Equal(t, "webrtc_offer", receivedMsg.Type, "Message type should be webrtc_offer")
	assert.Equal(t, "user1", receivedMsg.SenderUserID, "Sender should be user1")
}

func TestClientDisconnection(t *testing.T) {
	server, hub := setupTestServer(t)
	defer server.Close()

	// Connect a client
	conn := connectTestClient(t, server, "user1", "room1")

	// Wait for client to be registered
	time.Sleep(100 * time.Millisecond)

	// Close the connection
	conn.Close()

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify room is empty
	room, exists := hub.GetRoom("room1")
	if exists {
		assert.Equal(t, 0, len(room.clients), "Room should be empty after client disconnection")
	}
}

func TestInvalidMessageHandling(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Connect a client
	conn := connectTestClient(t, server, "user1", "room1")
	defer conn.Close()

	// Send invalid JSON
	err := conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	assert.NoError(t, err, "Should send invalid message without error")

	// Send message with unknown type
	invalidMsg := WebSocketMessage{
		Type:    "unknown_type",
		Payload: "test",
	}
	msgBytes, _ := json.Marshal(invalidMsg)
	err = conn.WriteMessage(websocket.TextMessage, msgBytes)
	assert.NoError(t, err, "Should send unknown message type without error")
}
