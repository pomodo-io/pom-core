package webrtc_signaling

import (
	"log"
	"sync" // For protecting the rooms map if accessed outside the run goroutine
)

// Hub maintains the set of active rooms and handles client registrations to rooms.
type Hub struct {
	rooms      map[string]*Room
	register   chan *Client // Client wants to join/be placed in a room
	unregister chan *Client // Client is disconnecting globally
	emptyRooms chan string  // Channel to signal a room is empty
	mu         sync.RWMutex // To protect the rooms map
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client), // For global client disconnections
		emptyRooms: make(chan string),
	}
}

func (h *Hub) Run() {
	log.Println("Hub started")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			room, ok := h.rooms[client.RoomID]
			if !ok {
				log.Printf("Room %s does not exist. Creating.", client.RoomID)
				room = newRoom(client.RoomID, h)
				h.rooms[client.RoomID] = room
				go room.run() // Start the new room's goroutine
			}
			h.mu.Unlock()
			room.register <- client // Register client to the specific room

		case client := <-h.unregister: // A client disconnected entirely
			h.mu.RLock()
			if room, ok := h.rooms[client.RoomID]; ok {
				room.unregister <- client // Forward unregistration to the room
			}
			h.mu.RUnlock()
			log.Printf("Hub noted client %s (room %s) disconnected.", client.UserID, client.RoomID)

		case roomID := <-h.emptyRooms:
			h.mu.Lock()
			if room, ok := h.rooms[roomID]; ok {
				if len(room.clients) == 0 { // Double check, though room.run() should ensure this
					delete(h.rooms, roomID)
					log.Printf("Hub removed empty room %s.", roomID)
				}
			}
			h.mu.Unlock()
		}
	}
}

// GetRoom (optional helper, ensure thread-safe access if called from HTTP handlers)
func (h *Hub) GetRoom(roomID string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, ok := h.rooms[roomID]
	return room, ok
}
