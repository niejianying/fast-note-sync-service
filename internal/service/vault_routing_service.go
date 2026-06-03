package service

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
)

type VaultRoutingService interface {
	IsSharedVault(ctx context.Context, vaultName string) (bool, error)
	GetNoteService(ctx context.Context, vaultName string, clientType, clientName, clientVersion string) NoteService
	GetFolderService(ctx context.Context, vaultName string, clientType, clientName, clientVersion string) FolderService
	GetFileService(ctx context.Context, vaultName string, clientType, clientName, clientVersion string) FileService
}

type vaultRoutingService struct {
	memberRepo  domain.VaultMemberRepository
	sharedNote  domain.SharedNoteRepository
	sharedFolder domain.SharedFolderRepository
	sharedFile  domain.SharedFileRepository
	noteSvc     NoteService
	folderSvc   FolderService
	fileSvc     FileService
}

func NewVaultRoutingService(
	memberRepo domain.VaultMemberRepository,
	sharedNote domain.SharedNoteRepository,
	sharedFolder domain.SharedFolderRepository,
	sharedFile domain.SharedFileRepository,
	noteSvc NoteService,
	folderSvc FolderService,
	fileSvc FileService,
) VaultRoutingService {
	return &vaultRoutingService{
		memberRepo:   memberRepo,
		sharedNote:   sharedNote,
		sharedFolder: sharedFolder,
		sharedFile:   sharedFile,
		noteSvc:      noteSvc,
		folderSvc:    folderSvc,
		fileSvc:      fileSvc,
	}
}

func (s *vaultRoutingService) IsSharedVault(ctx context.Context, vaultName string) (bool, error) {
	members, err := s.memberRepo.ListByVault(ctx, vaultName)
	if err != nil {
		return false, err
	}
	return len(members) > 1, nil
}

func (s *vaultRoutingService) GetNoteService(ctx context.Context, vaultName string, clientType, clientName, clientVersion string) NoteService {
	shared, _ := s.IsSharedVault(ctx, vaultName)
	if shared {
		return NewSharedNoteService(s.sharedNote, clientType, clientName, clientVersion)
	}
	return s.noteSvc
}

func (s *vaultRoutingService) GetFolderService(ctx context.Context, vaultName string, clientType, clientName, clientVersion string) FolderService {
	shared, _ := s.IsSharedVault(ctx, vaultName)
	if shared {
		return NewSharedFolderService(s.sharedFolder, clientType, clientName, clientVersion)
	}
	return s.folderSvc
}

func (s *vaultRoutingService) GetFileService(ctx context.Context, vaultName string, clientType, clientName, clientVersion string) FileService {
	shared, _ := s.IsSharedVault(ctx, vaultName)
	if shared {
		return NewSharedFileService(s.sharedFile, clientType, clientName, clientVersion)
	}
	return s.fileSvc
}

var _ VaultRoutingService = (*vaultRoutingService)(nil)
