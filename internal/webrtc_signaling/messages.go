package webrtc_signaling

// Generic structure for messages exchanged over Web
type WebSocketMessage struct {
	Type         string      `json:"type"`
	Payload      interface{} `json:"payload"`
	RoomID       string      `json:"roomID,omitempty"`       // Optional: Client can send, or server can infer/add
	SenderUserID string      `json:"senderUserID,omitempty"` // Server will set this reliably
	Timestamp    int64       `json:"timestamp,omitempty"`    // Server can set this
}

// ChatMessagePayload defines the content of a chat message.
type ChatMessagePayload struct {
	Content         string `json:"content"`
	UserDisplayName string `json:"userDisplayName,omitempty"`
}

// WebRTCSignalPayload (renamed from SignalMessagePayload for clarity)
// This is the payload when WebSocketMessage.Type is "webrtc_offer", "webrtc_answer", "webrtc_candidate"
type WebRTCSignalPayload struct {
	SignalType   string        `json:"signalType"` // "offer", "answer", "candidate"
	SDP          string        `json:"sdp,omitempty"`
	Candidate    *ICECandidate `json:"candidate,omitempty"`
	TargetUserID string        `json:"targetUserID,omitempty"` // For direct signals
}

// Structure for ICE Candidates
type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDMPid        string `json:"sdpMid"`
	SDMPLineIndex uint16 `json:"sdpMLineIndex"`
}

// SystemMessagePayload for server-generated notifications within a room.
type SystemMessagePayload struct {
	Event   string                 `json:"event"`             // e.g., "user_joined_room", "user_left_room", "pomodoro_started"
	Message string                 `json:"message"`           // Descriptive message
	UserID  string                 `json:"userID,omitempty"`  // User related to the event
	Details map[string]interface{} `json:"details,omitempty"` // Any extra details
}
