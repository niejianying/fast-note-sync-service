package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameSharedFile = "shared_file"

type SharedFile struct {
	ID         int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	VaultName  string     `gorm:"column:vault_name;not null;index;default:''" json:"vaultName" form:"vaultName"`
	Action     string     `gorm:"column:action;not null;default:'create'" json:"action" form:"action"`
	Path       string     `gorm:"column:path;not null;default:''" json:"path" form:"path"`
	PathHash   string     `gorm:"column:path_hash;not null;index;default:''" json:"pathHash" form:"pathHash"`
	ContentHash string    `gorm:"column:content_hash;not null;default:''" json:"contentHash" form:"contentHash"`
	SavePath   string     `gorm:"column:save_path;not null;default:''" json:"savePath" form:"savePath"`
	Size       int64      `gorm:"column:size;not null;default:0" json:"size" form:"size"`
	CTime      int64      `gorm:"column:ctime;not null;default:0" json:"ctime" form:"ctime"`
	MTime      int64      `gorm:"column:mtime;not null;default:0" json:"mtime" form:"mtime"`
	UpdatedUID int64      `gorm:"column:updated_uid;not null;default:0" json:"updatedUid" form:"updatedUid"`
	UpdatedAt  timex.Time `gorm:"column:updated_at;default:NULL;autoUpdateTime:false" json:"updatedAt" form:"updatedAt"`
}

func (*SharedFile) TableName() string { return TableNameSharedFile }
