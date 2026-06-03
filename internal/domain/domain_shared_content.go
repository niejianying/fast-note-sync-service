package domain

import (
	"context"
)

type SnapFile struct {
	Path      string
	PathHash  string
	Content   string
	ContentHash string
	Size      int64
	MTime     int64
	CTime     int64
	BaseHash  string
}

type SharedNoteRepository interface {
	GetByPath(ctx context.Context, vaultName, path string) (*SnapFile, error)
	Upsert(ctx context.Context, vaultName string, s *SnapFile, uid int64) (*SnapFile, error)
	Delete(ctx context.Context, vaultName, path string) error
	ListByVault(ctx context.Context, vaultName string, lastTime int64) ([]*SnapFile, error)
	GetLastTime(ctx context.Context, vaultName string) (int64, error)
}

type SharedFolderRepository interface {
	Upsert(ctx context.Context, vaultName, path, pathHash string, mtime int64, uid int64) error
	Delete(ctx context.Context, vaultName, path string) error
	ListByVault(ctx context.Context, vaultName string) ([]*string, error)
}

type SharedFileRepository interface {
	Upsert(ctx context.Context, vaultName string, s *SnapFile, uid int64) error
	Delete(ctx context.Context, vaultName, path string) error
	ListByVault(ctx context.Context, vaultName string, lastTime int64) ([]*SnapFile, error)
}
