package dto

type SignalingMessage struct {
	TargetUID int64  `json:"targetUid" validate:"required"`
	SessionID int64  `json:"sessionId"`
	SDP       string `json:"sdp,omitempty"`
	ICE       string `json:"ice,omitempty"`
	Type      string `json:"type,omitempty"` // "offer" | "answer" | "ice"
}

type RelayMessage struct {
	SessionID int64  `json:"sessionId" validate:"required"`
	Payload   string `json:"payload,omitempty"`
	Binary    []byte `json:"binary,omitempty"`
	ReqID     int64  `json:"reqId"`
}
