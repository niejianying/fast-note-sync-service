package service

import (
	"context"
	"io"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
)

type sharedFileService struct {
	repo domain.SharedFileRepository
}

func NewSharedFileService(repo domain.SharedFileRepository, ct, cn, cv string) FileService {
	return &sharedFileService{repo: repo}
}

func (s *sharedFileService) WithClient(ct, cn, cv string) FileService { return s }

func (s *sharedFileService) UpdateOrCreate(ctx context.Context, uid int64, params *dto.FileUpdateRequest, mtimeCheck bool) (bool, *dto.FileDTO, error) {
	snap := &domain.SnapFile{Path: params.Path, PathHash: params.PathHash, ContentHash: params.ContentHash, Size: params.Size, MTime: params.Mtime, CTime: params.Ctime}
	err := s.repo.Upsert(ctx, params.Vault, snap, uid)
	return true, &dto.FileDTO{Path: params.Path, PathHash: params.PathHash, ContentHash: params.ContentHash}, err
}

func (s *sharedFileService) Delete(ctx context.Context, uid int64, params *dto.FileDeleteRequest) (*dto.FileDTO, error) {
	err := s.repo.Delete(ctx, params.Vault, params.Path)
	return &dto.FileDTO{Path: params.Path}, err
}

func (s *sharedFileService) ListByLastTime(ctx context.Context, uid int64, params *dto.FileSyncRequest) ([]*dto.FileDTO, error) {
	snaps, err := s.repo.ListByVault(ctx, params.Vault, params.LastTime)
	if err != nil {
		return nil, err
	}
	var res []*dto.FileDTO
	for _, snap := range snaps {
		res = append(res, &dto.FileDTO{Path: snap.Path, PathHash: snap.PathHash, ContentHash: snap.ContentHash, Size: snap.Size, Mtime: snap.MTime, Ctime: snap.CTime})
	}
	return res, nil
}

func (s *sharedFileService) Get(ctx context.Context, uid int64, params *dto.FileGetRequest) (*dto.FileDTO, error) { return nil, nil }
func (s *sharedFileService) UpdateCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error) { return "", nil, nil }
func (s *sharedFileService) UploadCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error) { return "", nil, nil }
func (s *sharedFileService) UploadComplete(ctx context.Context, uid int64, params *dto.FileUpdateRequest) (bool, *dto.FileDTO, error) { return false, nil, nil }
func (s *sharedFileService) List(ctx context.Context, uid int64, params *dto.FileListRequest, pager *pkgapp.Pager) ([]*dto.FileDTO, int, error) { return nil, 0, nil }
func (s *sharedFileService) CountSizeSum(ctx context.Context, vaultID int64, uid int64) error { return nil }
func (s *sharedFileService) Cleanup(ctx context.Context, uid int64) error { return nil }
func (s *sharedFileService) CleanupByTime(ctx context.Context, cutoffTime int64) error { return nil }
func (s *sharedFileService) ResolveEmbedLinks(ctx context.Context, uid int64, vaultName string, notePath string, content string) (map[string]string, error) { return nil, nil }
func (s *sharedFileService) GetContent(ctx context.Context, uid int64, params *dto.FileGetRequest) (io.ReadCloser, string, int64, string, error) { return nil, "", 0, "", nil }
func (s *sharedFileService) GetContentInfo(ctx context.Context, uid int64, params *dto.FileGetRequest) (string, string, int64, string, string, error) { return "", "", 0, "", "", nil }
func (s *sharedFileService) Restore(ctx context.Context, uid int64, params *dto.FileRestoreRequest) (*dto.FileDTO, error) { return nil, nil }
func (s *sharedFileService) Rename(ctx context.Context, uid int64, params *dto.FileRenameRequest) (*dto.FileDTO, *dto.FileDTO, error) { return nil, nil, nil }
func (s *sharedFileService) RecycleClear(ctx context.Context, uid int64, params *dto.FileRecycleClearRequest) error { return nil }
func (s *sharedFileService) CleanDuplicateFiles(ctx context.Context, uid int64, vaultID int64) error { return nil }
func (s *sharedFileService) CleanDuplicateFilesAll(ctx context.Context) error { return nil }
