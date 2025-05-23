package webrtc_signaling

import (
	"reflect" // Useful for comparing slices or structs
	"sort"    // Good for comparing slices regardless of order
	"testing"
)

func TestSignaler_AddUserToRoom(t *testing.T) {
	s := NewSignaler() // Create a new signaler instance for each test

	roomID := "room123"
	userID1 := "userA"
	userID2 := "userB"

	// Test adding the first user
	s.AddUserToRoom(roomID, userID1)
	users, exists := s.GetUsersInRoom(roomID)
	if !exists {
		t.Errorf("Expected room %s to exist after adding user, but it doesn't", roomID)
	}
	expectedUsers1 := []string{userID1}
	// Sort slices before comparison if order doesn't matter
	sort.Strings(users)
	sort.Strings(expectedUsers1)
	if !reflect.DeepEqual(users, expectedUsers1) {
		t.Errorf("Expected users in room %s to be %v, but got %v", roomID, expectedUsers1, users)
	}

	// Test adding a second user to the same room
	s.AddUserToRoom(roomID, userID2)
	users, exists = s.GetUsersInRoom(roomID)
	if !exists {
		t.Errorf("Expected room %s to still exist after adding second user, but it doesn't", roomID)
	}
	expectedUsers2 := []string{userID1, userID2}
	sort.Strings(users)
	sort.Strings(expectedUsers2)
	if !reflect.DeepEqual(users, expectedUsers2) {
		t.Errorf("Expected users in room %s to be %v, but got %v", roomID, expectedUsers2, users)
	}

	// Test adding the same user again (should not change anything)
	s.AddUserToRoom(roomID, userID1)
	users, exists = s.GetUsersInRoom(roomID)
	if !exists {
		t.Errorf("Expected room %s to still exist after adding same user, but it doesn't", roomID)
	}
	sort.Strings(users)
	if !reflect.DeepEqual(users, expectedUsers2) { // Still expect the same users
		t.Errorf("Expected users in room %s to still be %v after adding same user, but got %v", roomID, expectedUsers2, users)
	}
}

func TestSignaler_RemoveUserFromRoom(t *testing.T) {
	s := NewSignaler()

	roomID := "room456"
	userID1 := "userX"
	userID2 := "userY"
	userID3 := "userZ"

	// Set up the room with multiple users
	s.AddUserToRoom(roomID, userID1)
	s.AddUserToRoom(roomID, userID2)
	s.AddUserToRoom(roomID, userID3)

	// Test removing a user
	s.RemoveUserFromRoom(roomID, userID2)
	users, exists := s.GetUsersInRoom(roomID)
	if !exists {
		t.Errorf("Expected room %s to exist after removing one user, but it doesn't", roomID)
	}
	expectedUsers1 := []string{userID1, userID3}
	sort.Strings(users)
	sort.Strings(expectedUsers1)
	if !reflect.DeepEqual(users, expectedUsers1) {
		t.Errorf("Expected users in room %s to be %v after removing %s, but got %v", roomID, expectedUsers1, userID2, users)
	}

	// Test removing the last user
	s.RemoveUserFromRoom(roomID, userID1)
	s.RemoveUserFromRoom(roomID, userID3)
	_, exists = s.GetUsersInRoom(roomID)
	if exists {
		t.Errorf("Expected room %s to be removed after removing all users, but it still exists", roomID)
	}

	// Test removing a user from a non-existent room (should do nothing and not panic)
	s.RemoveUserFromRoom("nonExistentRoom", "someUser")

	// Test removing a user who is not in the room (should not change anything)
	s = NewSignaler() // Reset for this specific test
	s.AddUserToRoom(roomID, userID1)
	s.RemoveUserFromRoom(roomID, "nonExistentUser")
	users, exists = s.GetUsersInRoom(roomID)
	if !exists || len(users) != 1 || users[0] != userID1 {
		t.Errorf("Removing non-existent user changed the room state unexpectedly. Expected [%s], got %v, exists: %t", userID1, users, exists)
	}
}

func TestSignaler_GetUsersInRoom(t *testing.T) {
	s := NewSignaler()
	roomID := "room789"
	userIDs := []string{"userM", "userN", "userO"}

	// Test getting users from a non-existent room
	users, exists := s.GetUsersInRoom("nonExistentRoom")
	if exists {
		t.Errorf("Expected non-existent room to not exist, but it does")
	}
	if len(users) != 0 {
		t.Errorf("Expected empty slice for non-existent room, but got %v", users)
	}

	// Add users to the room
	for _, userID := range userIDs {
		s.AddUserToRoom(roomID, userID)
	}

	// Test getting users from an existing room
	users, exists = s.GetUsersInRoom(roomID)
	if !exists {
		t.Errorf("Expected room %s to exist, but it doesn't", roomID)
	}
	expectedUsers := userIDs
	sort.Strings(users)
	sort.Strings(expectedUsers)
	if !reflect.DeepEqual(users, expectedUsers) {
		t.Errorf("Expected users %v for room %s, but got %v", expectedUsers, roomID, users)
	}
}
