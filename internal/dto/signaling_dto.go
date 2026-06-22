package dto

type SignalingMessage struct {
	TargetUID int64  `json:"targetUid" validate:"required"`
	SessionID int64  `json:"sessionId"`
	SDP       string `json:"sdp,omitempty"`
	ICE       string `json:"ice,omitempty"`
	Type      string `json:"type,omitempty"` // "offer" | "answer" | "ice"
}
