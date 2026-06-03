package service

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
)

type VaultShareService interface {
	Share(ctx context.Context, uid int64, params *dto.VaultShareRequest, vaultKey string) (*dto.SharedVaultDTO, error)
	Respond(ctx context.Context, uid int64, id int64, params *dto.VaultShareRespondRequest) (*dto.SharedVaultDTO, error)
	Revoke(ctx context.Context, uid int64, id int64) error
	ListIncoming(ctx context.Context, uid int64) ([]*dto.SharedVaultDTO, error)
	ListOutgoing(ctx context.Context, uid int64) ([]*dto.SharedVaultDTO, error)
}

type vaultShareService struct {
	sharedRepo   domain.SharedVaultRepository
	friendRepo   domain.FriendRelationshipRepository
	tokenService TokenService
	logger       *zap.Logger
	config       *ServiceConfig
}

func NewVaultShareService(
	sharedRepo domain.SharedVaultRepository,
	friendRepo domain.FriendRelationshipRepository,
	tokenService TokenService,
	logger *zap.Logger,
	config *ServiceConfig,
) VaultShareService {
	return &vaultShareService{
		sharedRepo:   sharedRepo,
		friendRepo:   friendRepo,
		tokenService: tokenService,
		logger:       logger,
		config:       config,
	}
}

func (s *vaultShareService) domainToDTO(d *domain.SharedVault) *dto.SharedVaultDTO {
	if d == nil {
		return nil
	}
	return &dto.SharedVaultDTO{
		ID:        d.ID,
		VaultName: d.VaultName,
		OwnerUID:  d.OwnerUID,
		TargetUID: d.TargetUID,
		VaultKey:  d.VaultKey,
		Token:     d.Token,
		Status:    string(d.Status),
		CreatedAt: timex.Time(d.CreatedAt),
		UpdatedAt: timex.Time(d.UpdatedAt),
	}
}

func (s *vaultShareService) Share(ctx context.Context, uid int64, params *dto.VaultShareRequest, vaultKey string) (*dto.SharedVaultDTO, error) {
	if uid == params.FriendUID {
		return nil, code.ErrorInvalidParams.WithDetails("Cannot share with yourself")
	}

	rels, err := s.friendRepo.ListByUID(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	isFriend := false
	for _, r := range rels {
		if r.FriendUID == params.FriendUID && r.IsActive() {
			isFriend = true
			break
		}
	}
	if !isFriend {
		return nil, code.ErrorInvalidParams.WithDetails("User is not your friend")
	}

	existing, _ := s.sharedRepo.GetByOwnerAndTarget(ctx, uid, params.FriendUID, params.VaultName)
	if existing != nil {
		return nil, code.ErrorInvalidParams.WithDetails("Vault already shared with this user")
	}

	sv := &domain.SharedVault{
		VaultName: params.VaultName,
		OwnerUID:  uid,
		TargetUID: params.FriendUID,
		VaultKey:  vaultKey,
		Status:    domain.SharedVaultPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create an access token for the target user to access this vault
	tokenResp, err := s.tokenService.Create(ctx, uid, &dto.TokenIssueRequest{
		ClientType:  "WebGui",
		Protocol:    "*",
		ExpiredDays: 365,
		Vaults:      params.VaultName,
	})
	if err != nil {
		return nil, code.ErrorTokenGenerate.WithDetails(err.Error())
	}
	sv.Token = tokenResp.TokenString

	created, err := s.sharedRepo.Create(ctx, sv)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(created), nil
}

func (s *vaultShareService) Respond(ctx context.Context, uid int64, id int64, params *dto.VaultShareRespondRequest) (*dto.SharedVaultDTO, error) {
	sv, err := s.sharedRepo.GetByID(ctx, id)
	if err != nil {
		return nil, code.ErrorInvalidParams.WithDetails("Share record not found")
	}
	if sv.TargetUID != uid {
		return nil, code.ErrorInvalidParams.WithDetails("Not your share record")
	}
	if sv.Status != domain.SharedVaultPending {
		return nil, code.ErrorInvalidParams.WithDetails("Share already responded")
	}

	if params.Accept {
		sv.Status = domain.SharedVaultAccepted
	} else {
		sv.Status = domain.SharedVaultDeclined
	}
	sv.UpdatedAt = time.Now()

	if err := s.sharedRepo.Update(ctx, sv); err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	dto := s.domainToDTO(sv)
	if params.Accept {
		dto.VaultKey = sv.VaultKey
		dto.Token = sv.Token
	}

	return dto, nil
}

func (s *vaultShareService) Revoke(ctx context.Context, uid int64, id int64) error {
	sv, err := s.sharedRepo.GetByID(ctx, id)
	if err != nil {
		return code.ErrorInvalidParams.WithDetails("Share record not found")
	}
	if sv.OwnerUID != uid {
		return code.ErrorInvalidParams.WithDetails("Not your share record")
	}
	return s.sharedRepo.Delete(ctx, id)
}

func (s *vaultShareService) ListIncoming(ctx context.Context, uid int64) ([]*dto.SharedVaultDTO, error) {
	list, err := s.sharedRepo.ListByTarget(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	var result []*dto.SharedVaultDTO
	for _, sv := range list {
		d := s.domainToDTO(sv)
		// Only return vaultKey for accepted shares
		if sv.Status == domain.SharedVaultAccepted {
			d.VaultKey = sv.VaultKey
		}
		result = append(result, d)
	}
	return result, nil
}

func (s *vaultShareService) ListOutgoing(ctx context.Context, uid int64) ([]*dto.SharedVaultDTO, error) {
	list, err := s.sharedRepo.ListByOwner(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	var result []*dto.SharedVaultDTO
	for _, sv := range list {
		result = append(result, s.domainToDTO(sv))
	}
	return result, nil
}

var _ VaultShareService = (*vaultShareService)(nil)
