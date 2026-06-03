package domain

import (
	"context"
	"time"
)

type SharedVaultStatus string

const (
	SharedVaultPending  SharedVaultStatus = "pending"
	SharedVaultAccepted SharedVaultStatus = "accepted"
	SharedVaultDeclined SharedVaultStatus = "declined"
)

type SharedVault struct {
	ID        int64
	VaultName string
	OwnerUID  int64
	TargetUID int64
	VaultKey  string
	Status    SharedVaultStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SharedVaultRepository interface {
	Create(ctx context.Context, sv *SharedVault) (*SharedVault, error)
	GetByID(ctx context.Context, id int64) (*SharedVault, error)
	Update(ctx context.Context, sv *SharedVault) error
	Delete(ctx context.Context, id int64) error
	ListByTarget(ctx context.Context, targetUID int64) ([]*SharedVault, error)
	ListByOwner(ctx context.Context, ownerUID int64) ([]*SharedVault, error)
	GetByOwnerAndTarget(ctx context.Context, ownerUID, targetUID int64, vaultName string) (*SharedVault, error)
}
