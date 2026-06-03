package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

type FriendRequestAdd struct {
	FriendUID int64  `json:"friendUid" form:"friendUid" binding:"required"`
	Message   string `json:"message" form:"message"`
}

type FriendRequestRespond struct {
	FriendUID int64 `json:"friendUid" form:"friendUid" binding:"required"`
	Accept    bool  `json:"accept" form:"accept" binding:"required"`
}

type FriendRelationshipDTO struct {
	ID        int64      `json:"id"`
	UID       int64      `json:"uid"`
	FriendUID int64      `json:"friendUid"`
	Status    string     `json:"status"`
	CreatedAt timex.Time `json:"createdAt"`
	UpdatedAt timex.Time `json:"updatedAt"`
}

type UserSearchDTO struct {
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
