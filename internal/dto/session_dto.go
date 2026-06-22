package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

type SessionCreateRequest struct {
	Name    string `json:"name" form:"name" binding:"required"`
}

type SessionJoinRequest struct {
	SessionID int64 `json:"sessionId" form:"sessionId" binding:"required"`
}

type SessionDTO struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	HostUID   int64      `json:"hostUid"`
	Status    string     `json:"status"`
	CreatedAt timex.Time `json:"createdAt"`
	UpdatedAt timex.Time `json:"updatedAt"`
	Members   []*SessionMemberDTO `json:"members,omitempty"`
}

type SessionMemberDTO struct {
	ID        int64      `json:"id"`
	SessionID int64      `json:"sessionId"`
	UID       int64      `json:"uid"`
	Role      string     `json:"role"`
	Online    bool       `json:"online"`
	JoinedAt  timex.Time `json:"joinedAt"`
}

type SessionListDTO struct {
	List  []*SessionDTO `json:"list"`
	Total int64         `json:"total"`
}
