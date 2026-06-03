package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameSharedNote = "shared_note"

type SharedNote struct {
	ID              int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	VaultName       string     `gorm:"column:vault_name;not null;index;default:''" json:"vaultName" form:"vaultName"`
	Action          string     `gorm:"column:action;not null;default:'create'" json:"action" form:"action"`
	Path            string     `gorm:"column:path;not null;default:''" json:"path" form:"path"`
	PathHash        string     `gorm:"column:path_hash;not null;index;default:''" json:"pathHash" form:"pathHash"`
	Content         string     `gorm:"column:content;type:longtext" json:"content" form:"content"`
	ContentHash     string     `gorm:"column:content_hash;not null;default:''" json:"contentHash" form:"contentHash"`
	Version         int64      `gorm:"column:version;not null;default:0" json:"version" form:"version"`
	Size            int64      `gorm:"column:size;not null;default:0" json:"size" form:"size"`
	CTime           int64      `gorm:"column:ctime;not null;default:0" json:"ctime" form:"ctime"`
	MTime           int64      `gorm:"column:mtime;not null;default:0" json:"mtime" form:"mtime"`
	CreatedUID      int64      `gorm:"column:created_uid;not null;default:0" json:"createdUid" form:"createdUid"`
	UpdatedUID      int64      `gorm:"column:updated_uid;not null;default:0" json:"updatedUid" form:"updatedUid"`
	UpdatedAt       timex.Time `gorm:"column:updated_at;default:NULL;autoUpdateTime:false" json:"updatedAt" form:"updatedAt"`
}

func (*SharedNote) TableName() string { return TableNameSharedNote }
