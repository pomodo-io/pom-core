package webrtc_signaling

// Generic structure for messages exchanged over Web
type WebSocketMessage struct {
	Type string `json:"type"`
	Payload interface{} `json:"payload"`
}

// Structure for messages exchanged over WebRTC Signaling Channel
type SignalMessagePayload struct {
	SignalType string `json:"signal_type"`
	SDP string `json:"sdp,omitempty"`
	Candidate *ICECandidate `json:"candidate,omitempty"`
	TargetUserID string `json:"targetUserID"`
}

// Structure for ICE Candidates
type ICECandidate struct {
	Candidate string `json:"candidate"`
	SDMPid string `json:"sdpMid"`
	SDMPLineIndex uint16 `json:"sdpMLineIndex"`
}