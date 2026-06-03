package service

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
)

type sharedFolderService struct {
	repo domain.SharedFolderRepository
}

func NewSharedFolderService(repo domain.SharedFolderRepository, ct, cn, cv string) FolderService {
	return &sharedFolderService{repo: repo}
}

func (s *sharedFolderService) UpdateOrCreate(ctx context.Context, uid int64, params *dto.FolderUpdateOrCreateRequest) (*dto.FolderDTO, error) {
	err := s.repo.Upsert(ctx, params.Vault, params.Path, params.PathHash, params.Mtime, uid)
	if err != nil {
		return nil, err
	}
	return &dto.FolderDTO{Path: params.Path, PathHash: params.PathHash, Mtime: params.Mtime}, nil
}

func (s *sharedFolderService) Delete(ctx context.Context, uid int64, params *dto.FolderDeleteRequest) (*dto.FolderDTO, error) {
	err := s.repo.Delete(ctx, params.Vault, params.Path)
	if err != nil {
		return nil, err
	}
	return &dto.FolderDTO{Path: params.Path}, nil
}

func (s *sharedFolderService) ListByUpdatedTimestamp(ctx context.Context, uid int64, params *dto.FolderQueryRequest) ([]*dto.FolderDTO, error) {
	paths, err := s.repo.ListByVault(ctx, params.Vault)
	if err != nil {
		return nil, err
	}
	var res []*dto.FolderDTO
	for _, p := range paths {
		res = append(res, &dto.FolderDTO{Path: *p})
	}
	return res, nil
}

func (s *sharedFolderService) WithClient(ct, cn, cv string) FolderService { return s }
func (s *sharedFolderService) Get(ctx context.Context, uid int64, params *dto.FolderGetRequest) (*dto.FolderDTO, error) { return nil, nil }
func (s *sharedFolderService) List(ctx context.Context, uid int64, params *dto.FolderQueryRequest) ([]*dto.FolderDTO, error) { return nil, nil }
func (s *sharedFolderService) ListNotes(ctx context.Context, uid int64, params *dto.FolderQueryRequest) ([]*dto.FolderDTO, error) { return nil, nil }
func (s *sharedFolderService) ListFiles(ctx context.Context, uid int64, params *dto.FolderQueryRequest) ([]*dto.FolderDTO, error) { return nil, nil }
func (s *sharedFolderService) Tree(ctx context.Context, uid int64, vault string) ([]*dto.FolderDTO, error) { return nil, nil }
