package service

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
)

type sharedNoteService struct {
	sharedNote domain.SharedNoteRepository
	clientType string
	clientName string
	clientVer  string
}

func NewSharedNoteService(sn domain.SharedNoteRepository, ct, cn, cv string) NoteService {
	return &sharedNoteService{sharedNote: sn, clientType: ct, clientName: cn, clientVer: cv}
}

func (s *sharedNoteService) WithClient(ct, cn, cv string) NoteService { return s }

func (s *sharedNoteService) UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *dto.NoteDTO, error) {
	existing, err := s.sharedNote.GetByPath(ctx, params.Vault, params.Path)
	if err != nil {
		return "Create", nil, nil
	}
	return "UpdateContent", &dto.NoteDTO{
		ContentHash:      existing.ContentHash,
		Content:          existing.Content,
		Size:             existing.Size,
		Mtime:            existing.MTime,
		UpdatedTimestamp: existing.MTime,
	}, nil
}

func (s *sharedNoteService) ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.NoteDTO, error) {
	snap := &domain.SnapFile{
		Path: params.Path, PathHash: params.PathHash,
		Content: params.Content, ContentHash: params.ContentHash,
		Size: params.Size, MTime: params.Mtime, CTime: params.Ctime,
	}
	_, err := s.sharedNote.Upsert(ctx, params.Vault, snap, uid)
	if err != nil {
		return false, nil, err
	}
	return true, &dto.NoteDTO{
		Path: params.Path, PathHash: params.PathHash,
		Content: params.Content, ContentHash: params.ContentHash, Size: params.Size,
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

func (s *sharedNoteService) ListByLastTime(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error) {
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

func (s *sharedNoteService) Sync(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error) {
	return s.ListByLastTime(ctx, uid, params)
}

func (s *sharedNoteService) Rename(ctx context.Context, uid int64, params *dto.NoteRenameRequest) (*dto.NoteDTO, *dto.NoteDTO, error) { return nil, nil, nil }
func (s *sharedNoteService) List(ctx context.Context, uid int64, params *dto.NoteListRequest, pager *pkgapp.Pager) ([]*dto.NoteNoContentDTO, int, error) { return nil, 0, nil }
func (s *sharedNoteService) Restore(ctx context.Context, uid int64, params *dto.NoteRestoreRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) RecycleClear(ctx context.Context, uid int64, params *dto.NoteRecycleClearRequest) error { return nil }
func (s *sharedNoteService) CountSizeSum(ctx context.Context, vaultID int64, uid int64) error { return nil }
func (s *sharedNoteService) Cleanup(ctx context.Context, uid int64) error { return nil }
func (s *sharedNoteService) CleanupByTime(ctx context.Context, cutoffTime int64) error { return nil }
func (s *sharedNoteService) ListNeedSnapshot(ctx context.Context, uid int64) ([]*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error { return nil }
func (s *sharedNoteService) MigratePush(oldNoteID, newNoteID int64, uid int64) {}
func (s *sharedNoteService) PatchFrontmatter(ctx context.Context, uid int64, params *dto.NotePatchFrontmatterRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) AppendContent(ctx context.Context, uid int64, params *dto.NoteAppendRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) PrependContent(ctx context.Context, uid int64, params *dto.NotePrependRequest) (*dto.NoteDTO, error) { return nil, nil }
func (s *sharedNoteService) ReplaceContent(ctx context.Context, uid int64, params *dto.NoteReplaceRequest) (*dto.NoteReplaceResponse, error) { return nil, nil }
func (s *sharedNoteService) UpdateNoteLinks(ctx context.Context, noteID int64, content string, vaultID, uid int64) {}
func (s *sharedNoteService) CleanDuplicateNotes(ctx context.Context, uid int64, vaultID int64) error { return nil }
func (s *sharedNoteService) CleanDuplicateNotesAll(ctx context.Context) error { return nil }
