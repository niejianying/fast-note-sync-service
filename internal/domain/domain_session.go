package domain

import (
	"context"
	"time"
)

type SessionStatus string

const (
	SessionStatusActive SessionStatus = "active"
	SessionStatusClosed SessionStatus = "closed"
)

type SessionMemberRole string

const (
	SessionMemberRoleHost   SessionMemberRole = "host"
	SessionMemberRoleMember SessionMemberRole = "member"
)

type CollabSession struct {
	ID         int64
	Name       string
	HostUID    int64
	Status     SessionStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SessionMember struct {
	ID        int64
	SessionID int64
	UID       int64
	Role      SessionMemberRole
	Online    bool
	JoinedAt  time.Time
}

func (s *CollabSession) IsActive() bool {
	return s.Status == SessionStatusActive
}

type SessionRepository interface {
	Create(ctx context.Context, session *CollabSession) (*CollabSession, error)
	GetByID(ctx context.Context, id int64) (*CollabSession, error)
	Update(ctx context.Context, session *CollabSession) error
	ListByUID(ctx context.Context, uid int64) ([]*CollabSession, error)

	AddMember(ctx context.Context, member *SessionMember) (*SessionMember, error)
	RemoveMember(ctx context.Context, sessionID, uid int64) error
	GetMember(ctx context.Context, sessionID, uid int64) (*SessionMember, error)
	ListMembers(ctx context.Context, sessionID int64) ([]*SessionMember, error)
	UpdateMemberOnline(ctx context.Context, sessionID, uid int64, online bool) error
	SetAllMembersOffline(ctx context.Context, sessionID int64) error
	CountMembers(ctx context.Context, sessionID int64) (int64, error)
}
