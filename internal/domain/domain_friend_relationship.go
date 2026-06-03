package domain

import (
	"context"
	"time"
)

type FriendStatus string

const (
	FriendStatusPending  FriendStatus = "pending"
	FriendStatusAccepted FriendStatus = "accepted"
	FriendStatusBlocked  FriendStatus = "blocked"
)

type FriendRelationship struct {
	ID        int64
	UID       int64
	FriendUID int64
	Status    FriendStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (f *FriendRelationship) IsActive() bool {
	return f.Status == FriendStatusAccepted
}

type FriendRelationshipRepository interface {
	GetByID(ctx context.Context, id int64) (*FriendRelationship, error)
	GetByUIDAndFriend(ctx context.Context, uid, friendUID int64) (*FriendRelationship, error)
	Create(ctx context.Context, rel *FriendRelationship) (*FriendRelationship, error)
	Update(ctx context.Context, rel *FriendRelationship) error
	Delete(ctx context.Context, uid, friendUID int64) error
	ListByUID(ctx context.Context, uid int64) ([]*FriendRelationship, error)
	ListAcceptedByUID(ctx context.Context, uid int64) ([]*FriendRelationship, error)
	ListPendingByUID(ctx context.Context, uid int64) ([]*FriendRelationship, error)
	GetReverse(ctx context.Context, uid, friendUID int64) (*FriendRelationship, error)
}
