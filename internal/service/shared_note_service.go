package service

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
)

type sharedNoteService struct {
	sharedNote  domain.SharedNoteRepository
	clientType  string
	clientName  string
	clientVer   string
}

func NewSharedNoteService(sn domain.SharedNoteRepository, ct, cn, cv string) NoteService {
	return &sharedNoteService{sharedNote: sn, clientType: ct, clientName: cn, clientVer: cv}
}

func (s *sharedNoteService) UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *dto.NoteUpdateCheckResponse, error) {
	existing, err := s.sharedNote.GetByPath(ctx, params.Vault, params.Path)
	if err != nil {
		return "Create", nil, nil
	}
	if existing.ContentHash == params.ContentHash {
		mtime := existing.MTime
		if mtime == 0 {
			mtime = time.Now().UnixMilli()
		}
		return "UpdateContent", &dto.NoteUpdateCheckResponse{
			ContentHash:      existing.ContentHash,
			Content:          existing.Content,
			Size:             existing.Size,
			UpdatedTimestamp: mtime,
		}, nil
	}
	return "UpdateContent", &dto.NoteUpdateCheckResponse{
		ContentHash:      existing.ContentHash,
		Content:          existing.Content,
		Size:             existing.Size,
		UpdatedTimestamp: existing.MTime,
	}, nil
}

func (s *sharedNoteService) ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, overwrite bool) (*dto.NoteDTO, error) {
	snap := &domain.SnapFile{
		Path: params.Path, PathHash: params.PathHash,
		Content: params.Content, ContentHash: params.ContentHash,
		Size: params.Size, MTime: params.Mtime, CTime: params.Ctime,
	}
	_, err := s.sharedNote.Upsert(ctx, params.Vault, snap, uid)
	if err != nil {
		return nil, err
	}
	return &dto.NoteDTO{
		Path: params.Path, PathHash: params.PathHash,
		ContentHash: params.ContentHash, Size: params.Size,
		Mtime: params.Mtime, Ctime: params.Ctime,
		ClientType: s.clientType, ClientName: s.clientName, ClientVersion: s.clientVer,
		UpdatedTimestamp: time.Now().UnixMilli(),
	}, nil
}

func (s *sharedNoteService) Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*dto.NoteDTO, error) {
	err := s.sharedNote.Delete(ctx, params.Vault, params.Path)
	if err != nil {
		return nil, err
	}
	return &dto.NoteDTO{Path: params.Path, PathHash: params.PathHash}, nil
}

func (s *sharedNoteService) ListByLastTime(ctx context.Context, uid int64, params *dto.NoteListRequest) ([]*dto.NoteDTO, error) {
	snaps, err := s.sharedNote.ListByVault(ctx, params.Vault, params.LastTime)
	if err != nil {
		return nil, err
	}
	var res []*dto.NoteDTO
	for _, snap := range snaps {
		res = append(res, &dto.NoteDTO{
			Path: snap.Path, PathHash: snap.PathHash,
			ContentHash: snap.ContentHash, Size: snap.Size,
			Mtime: snap.MTime, Ctime: snap.CTime,
		})
	}
	return res, nil
}

func (s *sharedNoteService) Get(ctx context.Context, uid int64, params *dto.NoteGetRequest) (*dto.NoteDTO, error) {
	snap, err := s.sharedNote.GetByPath(ctx, params.Vault, params.Path)
	if err != nil {
		return nil, err
	}
	return &dto.NoteDTO{
		Path: snap.Path, PathHash: snap.PathHash,
		Content: snap.Content, ContentHash: snap.ContentHash,
		Size: snap.Size, Mtime: snap.MTime, Ctime: snap.CTime,
	}, nil
}

// Stub implementations for unused NoteService methods
func (s *sharedNoteService) Create(ctx context.Context, uid int64, params *dto.NoteCreateRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) GetByPath(ctx context.Context, uid int64, vault, path, pathHash string) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) Rename(ctx context.Context, uid int64, params *dto.NoteRenameRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) GetBacklinks(ctx context.Context, uid int64, vault, path, pathHash string) ([]string, error) { return nil, nil }
func (s *sharedNoteService) GetOutlinks(ctx context.Context, uid int64, vault, path, pathHash string) ([]string, error) { return nil, nil }
func (s *sharedNoteService) PatchFrontmatter(ctx context.Context, uid int64, params *dto.NotePatchFrontmatterRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) AppendContent(ctx context.Context, uid int64, params *dto.NoteAppendRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) PrependContent(ctx context.Context, uid int64, params *dto.NotePrependRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) ReplaceContent(ctx context.Context, uid int64, params *dto.NoteReplaceRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) List(ctx context.Context, uid int64, params *dto.NoteListRequest) ([]*dto.NoteDTO, int64, error) { return nil, 0, nil }
func (s *sharedNoteService) GetPathHash(ctx context.Context, uid int64, vault, pathHash string) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) Restore(ctx context.Context, uid int64, params *dto.NoteRestoreRequest) error { return nil }
func (s *sharedNoteService) RecycleClear(ctx context.Context, uid int64, params *dto.NoteRecycleClearRequest) error { return nil }
