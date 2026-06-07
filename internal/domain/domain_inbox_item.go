package domain

import (
	"context"
	"time"
)

type InboxItem struct {
	ID             int64
	ItemID         string
	UID            int64
	Type           string
	Title          string
	Subtitle       string
	Payload        string
	SourceNotePath string
	SourceLine     int64
	IsRead         bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      time.Time
}

type InboxItemRepository interface {
	Create(ctx context.Context, item *InboxItem) (*InboxItem, error)
	GetByItemID(ctx context.Context, itemID string, uid int64) (*InboxItem, error)
	GetByID(ctx context.Context, id int64, uid int64) (*InboxItem, error)
	ListByUID(ctx context.Context, uid int64, page, pageSize int) ([]*InboxItem, int64, error)
	MarkRead(ctx context.Context, id int64, uid int64) error
	MarkAllRead(ctx context.Context, uid int64) error
	Delete(ctx context.Context, id int64, uid int64) error
}
