package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameCollabSession = "collab_session"
const TableNameSessionMember = "session_member"

type CollabSession struct {
	ID        int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	Name      string     `gorm:"column:name;not null;default:''" json:"name" form:"name"`
	HostUID   int64      `gorm:"column:host_uid;not null;index;default:0" json:"hostUid" form:"hostUid"`
	Status    string     `gorm:"column:status;not null;default:'active'" json:"status" form:"status"`
	CreatedAt timex.Time `gorm:"column:created_at;default:NULL;autoCreateTime:false" json:"createdAt" form:"createdAt"`
	UpdatedAt timex.Time `gorm:"column:updated_at;default:NULL;autoUpdateTime:false" json:"updatedAt" form:"updatedAt"`
}

func (*CollabSession) TableName() string {
	return TableNameCollabSession
}

type SessionMember struct {
	ID        int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	SessionID int64      `gorm:"column:session_id;not null;index;default:0" json:"sessionId" form:"sessionId"`
	UID       int64      `gorm:"column:uid;not null;index;default:0" json:"uid" form:"uid"`
	Role      string     `gorm:"column:role;not null;default:'member'" json:"role" form:"role"`
	Online    bool       `gorm:"column:online;not null;default:false" json:"online" form:"online"`
	JoinedAt  timex.Time `gorm:"column:joined_at;default:NULL;autoCreateTime:false" json:"joinedAt" form:"joinedAt"`
}

func (*SessionMember) TableName() string {
	return TableNameSessionMember
}
