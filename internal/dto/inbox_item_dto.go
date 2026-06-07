package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

type InboxItemDTO struct {
	ID             int64      `json:"id"`
	ItemID         string     `json:"itemId"`
	UID            int64      `json:"uid"`
	Type           string     `json:"type"`
	Title          string     `json:"title"`
	Subtitle       string     `json:"subtitle,omitempty"`
	Payload        string     `json:"payload,omitempty"`
	SourceNotePath string     `json:"sourceNotePath,omitempty"`
	SourceLine     int64      `json:"sourceLine,omitempty"`
	IsRead         bool       `json:"isRead"`
	CreatedAt      timex.Time `json:"createdAt"`
	UpdatedAt      timex.Time `json:"updatedAt"`
}

type InboxItemCreateRequest struct {
	ItemID         string `json:"itemId" form:"itemId" binding:"required"`
	Type           string `json:"type" form:"type" binding:"required"`
	Title          string `json:"title" form:"title" binding:"required"`
	Subtitle       string `json:"subtitle" form:"subtitle"`
	Payload        string `json:"payload" form:"payload"`
	SourceNotePath string `json:"sourceNotePath" form:"sourceNotePath"`
	SourceLine     int64  `json:"sourceLine" form:"sourceLine"`
}

type InboxItemListDTO struct {
	List  []*InboxItemDTO `json:"list"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}
