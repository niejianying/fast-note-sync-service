package domain

import "context"

type VaultMember struct {
	ID        int64
	VaultName string
	OwnerUID  int64
	MemberUID int64
}

type VaultMemberRepository interface {
	Add(ctx context.Context, m *VaultMember) (*VaultMember, error)
	Remove(ctx context.Context, vaultName string, memberUID int64) error
	ListByMember(ctx context.Context, memberUID int64) ([]*VaultMember, error)
	ListByVault(ctx context.Context, vaultName string) ([]*VaultMember, error)
	IsMember(ctx context.Context, vaultName string, uid int64) (bool, error)
}
