package webrtc_signaling

import (
	"fmt" 
	"sync" 
)

type Signaler struct {
	// maps room ID to a list of user IDs
	rooms map[string][]string

	// mutex to protect the map from concurrent access
	mu sync.Mutex
}

// NewSignaler for new Signaler instance.
func NewSignaler() *Signaler {
	// Initialize map
	return &Signaler{
		rooms: make(map[string][]string),
	}
}

func (s *Signaler) AddUserToRoom(roomID string, userID string) {
	s.mu.Lock() 
	defer s.mu.Unlock() // Release the lock when the function exits

	users, exists := s.rooms[roomID]
	if !exists {
		// If does not exist, create the room with the first user
		s.rooms[roomID] = []string{userID}
		fmt.Printf("Room %s created, user %s joined.\n", roomID, userID)
		return
	}

	// Check if the user is already in the room
	for _, existingUser := range users {
		if existingUser == userID {
			fmt.Printf("User %s is already in room %s.\n", userID, roomID)
			return
		}
	}

	// Add the user to the existing room
	s.rooms[roomID] = append(users, userID)
	fmt.Printf("User %s joined room %s. Current users: %v\n", userID, roomID, s.rooms[roomID])
}

// RemoveUserFromRoom removes a user from the specified simulated room.
func (s *Signaler) RemoveUserFromRoom(roomID string, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, exists := s.rooms[roomID]
	if !exists {
		// Room doesn't exist, nothing to remove
		fmt.Printf("Attempted to remove user %s from non-existent room %s.\n", userID, roomID)
		return
	}

	// Find and remove the user from the slice
	newUsers := []string{}
	for _, existingUser := range users {
		if existingUser != userID {
			newUsers = append(newUsers, existingUser)
		}
	}

	s.rooms[roomID] = newUsers
	fmt.Printf("User %s left room %s. Remaining users: %v\n", userID, roomID, s.rooms[roomID])

	// Optional: Clean up the room if it's empty
	if len(s.rooms[roomID]) == 0 {
		delete(s.rooms, roomID)
		fmt.Printf("Room %s is now empty and has been removed.\n", roomID)
	}
}

// GetUsersInRoom returns a slice of user IDs in the specified simulated room.
// Returns an empty slice and false if the room does not exist.
func (s *Signaler) GetUsersInRoom(roomID string) ([]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, exists := s.rooms[roomID]
	if !exists {
		return []string{}, false
	}

	// Return a copy of the slice to prevent external modification of the internal map
	usersCopy := make([]string, len(users))
	copy(usersCopy, users)
	return usersCopy, true
}
