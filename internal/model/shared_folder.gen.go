package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameSharedFolder = "shared_folder"

type SharedFolder struct {
	ID        int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	VaultName string     `gorm:"column:vault_name;not null;index;default:''" json:"vaultName" form:"vaultName"`
	Action    string     `gorm:"column:action;not null;default:'create'" json:"action" form:"action"`
	Path      string     `gorm:"column:path;not null;default:''" json:"path" form:"path"`
	PathHash  string     `gorm:"column:path_hash;not null;index;default:''" json:"pathHash" form:"pathHash"`
	Level     int64      `gorm:"column:level;not null;default:0" json:"level" form:"level"`
	FID       string     `gorm:"column:fid;not null;default:''" json:"fid" form:"fid"`
	CTime     int64      `gorm:"column:ctime;not null;default:0" json:"ctime" form:"ctime"`
	MTime     int64      `gorm:"column:mtime;not null;default:0" json:"mtime" form:"mtime"`
	UpdatedUID int64     `gorm:"column:updated_uid;not null;default:0" json:"updatedUid" form:"updatedUid"`
	UpdatedAt  timex.Time `gorm:"column:updated_at;default:NULL;autoUpdateTime:false" json:"updatedAt" form:"updatedAt"`
}

func (*SharedFolder) TableName() string { return TableNameSharedFolder }
