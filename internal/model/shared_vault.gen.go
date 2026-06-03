package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameSharedVault = "shared_vault"

type SharedVault struct {
	ID        int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	VaultName string     `gorm:"column:vault_name;not null;default:''" json:"vaultName" form:"vaultName"`
	OwnerUID  int64      `gorm:"column:owner_uid;not null;index;default:0" json:"ownerUid" form:"ownerUid"`
	TargetUID int64      `gorm:"column:target_uid;not null;index;default:0" json:"targetUid" form:"targetUid"`
	VaultKey  string     `gorm:"column:vault_key;not null;default:''" json:"vaultKey" form:"vaultKey"`
	Status    string     `gorm:"column:status;not null;default:'pending'" json:"status" form:"status"`
	CreatedAt timex.Time `gorm:"column:created_at;default:NULL;autoCreateTime:false" json:"createdAt" form:"createdAt"`
	UpdatedAt timex.Time `gorm:"column:updated_at;default:NULL;autoUpdateTime:false" json:"updatedAt" form:"updatedAt"`
}

func (*SharedVault) TableName() string {
	return TableNameSharedVault
}
