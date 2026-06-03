package service

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
)

type sharedFileService struct {
	repo domain.SharedFileRepository
}

func NewSharedFileService(repo domain.SharedFileRepository, ct, cn, cv string) FileService {
	return &sharedFileService{repo: repo}
}

func (s *sharedFileService) Upload(ctx context.Context, uid int64, params *dto.FileUploadRequest) (*dto.FileDTO, error) {
	snap := &domain.SnapFile{
		Path: params.Path, PathHash: params.PathHash,
		ContentHash: params.ContentHash, Size: params.Size,
		MTime: params.Mtime, CTime: params.Ctime, Content: "",
	}
	err := s.repo.Upsert(ctx, params.Vault, snap, uid)
	if err != nil {
		return nil, err
	}
	return &dto.FileDTO{Path: params.Path, PathHash: params.PathHash, ContentHash: params.ContentHash, Size: params.Size}, nil
}

func (s *sharedFileService) Delete(ctx context.Context, uid int64, params *dto.FileDeleteRequest) error {
	return s.repo.Delete(ctx, params.Vault, params.Path)
}

func (s *sharedFileService) ListByUpdatedTimestamp(ctx context.Context, uid int64, params *dto.FileListRequest) ([]*dto.FileDTO, error) {
	snaps, err := s.repo.ListByVault(ctx, params.Vault, params.LastTime)
	if err != nil {
		return nil, err
	}
	var res []*dto.FileDTO
	for _, snap := range snaps {
		res = append(res, &dto.FileDTO{
			Path: snap.Path, PathHash: snap.PathHash,
			ContentHash: snap.ContentHash, Size: snap.Size,
			Mtime: snap.MTime, Ctime: snap.CTime,
		})
	}
	return res, nil
}

func (s *sharedFileService) WithClient(ct, cn, cv string) FileService { return s }
func (s *sharedFileService) GetInfo(ctx context.Context, uid int64, params *dto.FileGetRequest) (*dto.FileDTO, error) { return nil, nil }
func (s *sharedFileService) Get(ctx context.Context, uid int64, params *dto.FileGetRequest) (*dto.FileDTO, error) { return nil, nil }
func (s *sharedFileService) Rename(ctx context.Context, uid int64, params *dto.FileRenameRequest) (*dto.FileDTO, error) { return nil, nil }
func (s *sharedFileService) Restore(ctx context.Context, uid int64, params *dto.FileRestoreRequest) error { return nil }
func (s *sharedFileService) List(ctx context.Context, uid int64, params *dto.FileListRequest) ([]*dto.FileDTO, int64, error) { return nil, 0, nil }
func (s *sharedFileService) RecycleClear(ctx context.Context, uid int64, params *dto.FileRecycleClearRequest) error { return nil }
func (s *sharedFileService) GetByPath(ctx context.Context, uid int64, vaultName, path, pathHash string) (*dto.FileDTO, error) { return nil, nil }
func (s *sharedFileService) GetPathHash(ctx context.Context, uid int64, vault, pathHash string) (*dto.FileDTO, error) { return nil, nil }
