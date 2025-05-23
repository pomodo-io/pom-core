package webrtc_signaling

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 // Adjusted from 512, can be tuned
)

// Client is a middleman between the WebSocket connection and the hub/room.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte // Buffered channel of outbound messages (marshalled JSON)
	UserID string      // Authenticated User ID
	RoomID string      // Room the client is currently in
	room   *Room       // Reference to the Room object
}

// readPump pumps messages from the WebSocket connection to the room or hub.
func (c *Client) readPump() {
	defer func() {
		if c.room != nil {
			c.room.unregister <- c
		} else {
			// If client was never fully registered to a room but hub knows it
			c.hub.unregister <- c
		}
		c.conn.Close()
		log.Printf("Client %s disconnected from room %s", c.UserID, c.RoomID)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Client %s read error: %v", c.UserID, err)
			}
			break
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Printf("Client %s error unmarshalling message: %v. Raw: %s", c.UserID, err, string(rawMessage))
			continue
		}

		// Crucially, set server-side information
		msg.SenderUserID = c.UserID
		msg.RoomID = c.RoomID // Ensure roomID is consistent
		msg.Timestamp = time.Now().UnixMilli()

		if c.room == nil {
			log.Printf("Client %s in room %s has no room assigned. Dropping message type %s", c.UserID, c.RoomID, msg.Type)
			continue
		}

		// Route message based on type
		switch msg.Type {
		case "chat":
			// Re-marshal with server-set fields before broadcasting
			// The payload should already be ChatMessagePayload if client sent correctly
			// No further unmarshalling of payload needed here if client sends it correctly structured.
			// The room's broadcast will send this enriched `msg`.
			c.room.broadcast <- &msg

		case "webrtc_offer", "webrtc_answer", "webrtc_candidate":
			// The payload should be WebRTCSignalPayload
			// Handle WebRTC signaling: forward to target or broadcast to room (excluding self)
			var signalPayload WebRTCSignalPayload
			payloadBytes, _ := json.Marshal(msg.Payload) // Get payload back to bytes
			if err := json.Unmarshal(payloadBytes, &signalPayload); err != nil {
				log.Printf("Client %s error unmarshalling WebRTC payload: %v", c.UserID, err)
				continue
			}

			// Forwarding logic (simplified example)
			for peerClient := range c.room.clients {
				if msg.Type == "webrtc_candidate" && peerClient == c {
					continue // Don't send own candidates back to self unless needed
				}
				// If TargetUserID is specified and matches, send only to them
				if signalPayload.TargetUserID != "" && peerClient.UserID != signalPayload.TargetUserID {
					continue
				}
				// If it's an offer/answer, it's usually targeted.
				// If it's a candidate, it might be broadcast or targeted.
				if peerClient != c || signalPayload.TargetUserID == c.UserID { // Send to others, or if targeted to self (unlikely for offer/answer)
					// Re-marshal the full WebSocketMessage for sending
					fullMsgBytes, _ := json.Marshal(msg)
					select {
					case peerClient.send <- fullMsgBytes:
					default:
						log.Printf("Client %s send channel full for WebRTC to %s", c.UserID, peerClient.UserID)
						// Consider closing peerClient.send and removing them
					}
				}
			}

		default:
			log.Printf("Client %s sent unknown message type: %s", c.UserID, msg.Type)
		}
	}
}

// writePump pumps messages from the send channel to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case messageBytes, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				log.Printf("Client %s write error: %v", c.UserID, err)
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Client %s ping error: %v", c.UserID, err)
				return
			}
		}
	}
}
