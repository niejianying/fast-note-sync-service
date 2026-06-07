package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameInboxItem = "inbox_item"

type InboxItem struct {
	ID             int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	ItemID         string     `gorm:"column:item_id;uniqueIndex;size:255;not null" json:"itemId" form:"itemId"`
	UID            int64      `gorm:"column:uid;index:idx_uid_created;not null;default:0" json:"uid" form:"uid"`
	Type           string     `gorm:"column:type;size:50;not null" json:"type" form:"type"`
	Title          string     `gorm:"column:title;size:255;not null" json:"title" form:"title"`
	Subtitle       string     `gorm:"column:subtitle;size:500" json:"subtitle" form:"subtitle"`
	Payload        string     `gorm:"column:payload;type:text" json:"payload" form:"payload"`
	SourceNotePath string     `gorm:"column:source_note_path;size:500" json:"sourceNotePath" form:"sourceNotePath"`
	SourceLine     int64      `gorm:"column:source_line;default:0" json:"sourceLine" form:"sourceLine"`
	IsRead         bool       `gorm:"column:is_read;default:false" json:"isRead" form:"isRead"`
	CreatedAt      timex.Time `gorm:"column:created_at;default:NULL;autoCreateTime:false" json:"createdAt" form:"createdAt"`
	UpdatedAt      timex.Time `gorm:"column:updated_at;default:NULL;autoUpdateTime:false" json:"updatedAt" form:"updatedAt"`
	DeletedAt      timex.Time `gorm:"column:deleted_at;default:NULL;index" json:"deletedAt" form:"deletedAt"`
}

func (*InboxItem) TableName() string {
	return TableNameInboxItem
}
