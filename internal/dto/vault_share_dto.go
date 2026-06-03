package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

type VaultShareRequest struct {
	FriendUID int64  `json:"friendUid" form:"friendUid" binding:"required"`
	VaultName string `json:"vaultName" form:"vaultName" binding:"required"`
}

type VaultShareRespondRequest struct {
	Accept bool `json:"accept" form:"accept" binding:"required"`
}

type SharedVaultDTO struct {
	ID        int64      `json:"id"`
	VaultName string     `json:"vaultName"`
	OwnerUID  int64      `json:"ownerUid"`
	TargetUID int64      `json:"targetUid"`
	VaultKey  string     `json:"vaultKey,omitempty"`
	Status    string     `json:"status"`
	CreatedAt timex.Time `json:"createdAt"`
	UpdatedAt timex.Time `json:"updatedAt"`
}
