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

type FriendRelationshipService interface {
	AddFriend(ctx context.Context, uid int64, params *dto.FriendRequestAdd) (*dto.FriendRelationshipDTO, error)
	RespondToRequest(ctx context.Context, uid int64, params *dto.FriendRequestRespond) (*dto.FriendRelationshipDTO, error)
	RemoveFriend(ctx context.Context, uid, friendUID int64) error
	ListFriends(ctx context.Context, uid int64) ([]*dto.FriendRelationshipDTO, error)
	ListPendingRequests(ctx context.Context, uid int64) ([]*dto.FriendRelationshipDTO, error)
}

type friendRelationshipService struct {
	friendRepo domain.FriendRelationshipRepository
	userRepo   domain.UserRepository
	logger     *zap.Logger
	config     *ServiceConfig
}

func NewFriendRelationshipService(
	friendRepo domain.FriendRelationshipRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
	config *ServiceConfig,
) FriendRelationshipService {
	return &friendRelationshipService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
		logger:     logger,
		config:     config,
	}
}

func (s *friendRelationshipService) domainToDTO(d *domain.FriendRelationship) *dto.FriendRelationshipDTO {
	if d == nil {
		return nil
	}
	return &dto.FriendRelationshipDTO{
		ID:        d.ID,
		UID:       d.UID,
		FriendUID: d.FriendUID,
		Status:    string(d.Status),
		UpdatedAt: timex.Time(d.UpdatedAt),
		CreatedAt: timex.Time(d.CreatedAt),
	}
}

func (s *friendRelationshipService) AddFriend(ctx context.Context, uid int64, params *dto.FriendRequestAdd) (*dto.FriendRelationshipDTO, error) {
	if uid == params.FriendUID {
		return nil, code.ErrorInvalidParams.WithDetails("Cannot add yourself as friend")
	}

	existing, err := s.friendRepo.GetByUIDAndFriend(ctx, uid, params.FriendUID)
	if err == nil && existing != nil {
		if existing.Status == domain.FriendStatusBlocked {
			existing.Status = domain.FriendStatusPending
			existing.UpdatedAt = time.Now()
			if err := s.friendRepo.Update(ctx, existing); err != nil {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
			return s.domainToDTO(existing), nil
		}
		return nil, code.ErrorFriendAlreadyExists
	}

	reverse, err := s.friendRepo.GetReverse(ctx, uid, params.FriendUID)
	if err == nil && reverse != nil {
		if reverse.Status == domain.FriendStatusBlocked {
			reverse.Status = domain.FriendStatusPending
			reverse.UpdatedAt = time.Now()
			if err := s.friendRepo.Update(ctx, reverse); err != nil {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
			return s.domainToDTO(reverse), nil
		}
		return nil, code.ErrorFriendAlreadyExists
	}

	rel := &domain.FriendRelationship{
		UID:       uid,
		FriendUID: params.FriendUID,
		Status:    domain.FriendStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	created, err := s.friendRepo.Create(ctx, rel)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(created), nil
}

func (s *friendRelationshipService) RespondToRequest(ctx context.Context, uid int64, params *dto.FriendRequestRespond) (*dto.FriendRelationshipDTO, error) {
	rel, err := s.friendRepo.GetReverse(ctx, uid, params.FriendUID)
	if err != nil {
		return nil, code.ErrorFriendRequestNotFound
	}
	if rel == nil || rel.Status != domain.FriendStatusPending {
		return nil, code.ErrorFriendRequestNotFound
	}

	if params.Accept {
		rel.Status = domain.FriendStatusAccepted
	} else {
		rel.Status = domain.FriendStatusBlocked
	}
	rel.UpdatedAt = time.Now()

	if err := s.friendRepo.Update(ctx, rel); err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(rel), nil
}

func (s *friendRelationshipService) RemoveFriend(ctx context.Context, uid, friendUID int64) error {
	if err := s.friendRepo.Delete(ctx, uid, friendUID); err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	_ = s.friendRepo.Delete(ctx, friendUID, uid)
	return nil
}

func (s *friendRelationshipService) ListFriends(ctx context.Context, uid int64) ([]*dto.FriendRelationshipDTO, error) {
	rels, err := s.friendRepo.ListAcceptedByUID(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	var result []*dto.FriendRelationshipDTO
	for _, r := range rels {
		// Normalize: if the current user is the friendUID, swap to show friendUid as the friend
		dto := s.domainToDTO(r)
		if dto.FriendUID == uid {
			dto.UID, dto.FriendUID = dto.FriendUID, dto.UID
		}
		result = append(result, dto)
	}
	return result, nil
}

func (s *friendRelationshipService) ListPendingRequests(ctx context.Context, uid int64) ([]*dto.FriendRelationshipDTO, error) {
	rels, err := s.friendRepo.ListPendingByUID(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	var result []*dto.FriendRelationshipDTO
	for _, r := range rels {
		result = append(result, s.domainToDTO(r))
	}
	return result, nil
}

var _ FriendRelationshipService = (*friendRelationshipService)(nil)
