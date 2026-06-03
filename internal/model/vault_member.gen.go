package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameVaultMember = "vault_member"

type VaultMember struct {
	ID        int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	VaultName string     `gorm:"column:vault_name;not null;index:idx_vault_user,uniqueIndex:idx_vault_user;default:''" json:"vaultName" form:"vaultName"`
	OwnerUID  int64      `gorm:"column:owner_uid;not null;index;default:0" json:"ownerUid" form:"ownerUid"`
	MemberUID int64      `gorm:"column:member_uid;not null;index:idx_vault_user,uniqueIndex:idx_vault_user;default:0" json:"memberUid" form:"memberUid"`
	CreatedAt timex.Time `gorm:"column:created_at;default:NULL;autoCreateTime:false" json:"createdAt" form:"createdAt"`
}

func (*VaultMember) TableName() string { return TableNameVaultMember }
