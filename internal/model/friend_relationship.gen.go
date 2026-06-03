package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameFriendRelationship = "friend_relationship"

type FriendRelationship struct {
	ID        int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	UID       int64      `gorm:"column:uid;not null;index;default:0" json:"uid" form:"uid"`
	FriendUID int64     `gorm:"column:friend_uid;not null;index;default:0" json:"friendUid" form:"friendUid"`
	Status    string     `gorm:"column:status;not null;default:'pending'" json:"status" form:"status"`
	CreatedAt timex.Time `gorm:"column:created_at;default:NULL;autoCreateTime:false" json:"createdAt" form:"createdAt"`
	UpdatedAt timex.Time `gorm:"column:updated_at;default:NULL;autoUpdateTime:false" json:"updatedAt" form:"updatedAt"`
	DeletedAt timex.Time `gorm:"column:deleted_at;default:NULL" json:"deletedAt" form:"deletedAt"`
}

func (*FriendRelationship) TableName() string {
	return TableNameFriendRelationship
}
