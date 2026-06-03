package service

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
)

type sharedFolderService struct {
	repo domain.SharedFolderRepository
}

func NewSharedFolderService(repo domain.SharedFolderRepository, ct, cn, cv string) FolderService {
	return &sharedFolderService{repo: repo}
}

func (s *sharedFolderService) WithClient(ct, cn, cv string) FolderService { return s }

func (s *sharedFolderService) UpdateOrCreate(ctx context.Context, uid int64, params *dto.FolderCreateRequest) (*dto.FolderDTO, error) {
	err := s.repo.Upsert(ctx, params.Vault, params.Path, params.PathHash, 0, uid)
	if err != nil {
		return nil, err
	}
	return &dto.FolderDTO{Path: params.Path, PathHash: params.PathHash}, nil
}

func (s *sharedFolderService) Delete(ctx context.Context, uid int64, params *dto.FolderDeleteRequest) (*dto.FolderDTO, error) {
	err := s.repo.Delete(ctx, params.Vault, params.Path)
	if err != nil {
		return nil, err
	}
	return &dto.FolderDTO{Path: params.Path}, nil
}

func (s *sharedFolderService) ListByUpdatedTimestamp(ctx context.Context, uid int64, vault string, lastTime int64) ([]*dto.FolderDTO, error) {
	paths, err := s.repo.ListByVault(ctx, vault)
	if err != nil {
		return nil, err
	}
	var res []*dto.FolderDTO
	for _, p := range paths {
		res = append(res, &dto.FolderDTO{Path: *p})
	}
	return res, nil
}

func (s *sharedFolderService) Get(ctx context.Context, uid int64, params *dto.FolderGetRequest) (*dto.FolderDTO, error) { return nil, nil }
func (s *sharedFolderService) List(ctx context.Context, uid int64, params *dto.FolderListRequest) ([]*dto.FolderDTO, error) { return nil, nil }
func (s *sharedFolderService) Rename(ctx context.Context, uid int64, params *dto.FolderRenameRequest) (*dto.FolderDTO, *dto.FolderDTO, error) { return nil, nil, nil }
func (s *sharedFolderService) ListNotes(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *pkgapp.Pager) ([]*dto.NoteNoContentDTO, int, error) { return nil, 0, nil }
func (s *sharedFolderService) ListFiles(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *pkgapp.Pager) ([]*dto.FileDTO, int, error) { return nil, 0, nil }
func (s *sharedFolderService) EnsurePathFID(ctx context.Context, uid int64, vaultID int64, path string) (int64, error) { return 0, nil }
func (s *sharedFolderService) SyncResourceFID(ctx context.Context, uid int64, vaultID int64, noteIDs []int64, fileIDs []int64) error { return nil }
func (s *sharedFolderService) GetTree(ctx context.Context, uid int64, params *dto.FolderTreeRequest) (*dto.FolderTreeResponse, error) { return nil, nil }
func (s *sharedFolderService) CleanDuplicateFolders(ctx context.Context, uid int64, vaultID int64) error { return nil }
