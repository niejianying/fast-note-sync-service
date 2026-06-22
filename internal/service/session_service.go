package service

import (
	"context"
	"errors"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SessionService interface {
	Create(ctx context.Context, uid int64, params *dto.SessionCreateRequest) (*dto.SessionDTO, error)
	Join(ctx context.Context, uid int64, params *dto.SessionJoinRequest) (*dto.SessionDTO, error)
	Leave(ctx context.Context, uid int64, sessionID int64) error
	Close(ctx context.Context, uid int64, sessionID int64) error
	Get(ctx context.Context, uid int64, sessionID int64) (*dto.SessionDTO, error)
	List(ctx context.Context, uid int64) ([]*dto.SessionDTO, error)
	ListMembers(ctx context.Context, sessionID int64) ([]*dto.SessionMemberDTO, error)
	SetOnline(ctx context.Context, sessionID, uid int64, online bool) error
	SetAllOffline(ctx context.Context, sessionID int64) error
}

type sessionService struct {
	sessionRepo domain.SessionRepository
	userRepo    domain.UserRepository
	logger      *zap.Logger
	config      *ServiceConfig
}

func NewSessionService(
	sessionRepo domain.SessionRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
	config *ServiceConfig,
) SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		logger:      logger,
		config:      config,
	}
}

func (s *sessionService) sessionToDTO(d *domain.CollabSession) *dto.SessionDTO {
	if d == nil {
		return nil
	}
	return &dto.SessionDTO{
		ID:        d.ID,
		Name:      d.Name,
		HostUID:   d.HostUID,
		Status:    string(d.Status),
		CreatedAt: timex.Time(d.CreatedAt),
		UpdatedAt: timex.Time(d.UpdatedAt),
	}
}

func (s *sessionService) memberToDTO(d *domain.SessionMember) *dto.SessionMemberDTO {
	if d == nil {
		return nil
	}
	return &dto.SessionMemberDTO{
		ID:        d.ID,
		SessionID: d.SessionID,
		UID:       d.UID,
		Role:      string(d.Role),
		Online:    d.Online,
		JoinedAt:  timex.Time(d.JoinedAt),
	}
}

func (s *sessionService) Create(ctx context.Context, uid int64, params *dto.SessionCreateRequest) (*dto.SessionDTO, error) {
	session := &domain.CollabSession{
		Name:     params.Name,
		HostUID:  uid,
		Status:   domain.SessionStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	created, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	member := &domain.SessionMember{
		SessionID: created.ID,
		UID:       uid,
		Role:      domain.SessionMemberRoleHost,
		Online:    true,
		JoinedAt:  time.Now(),
	}

	if _, err := s.sessionRepo.AddMember(ctx, member); err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	dto := s.sessionToDTO(created)
	members, _ := s.sessionRepo.ListMembers(ctx, created.ID)
	for _, m := range members {
		dto.Members = append(dto.Members, s.memberToDTO(m))
	}
	return dto, nil
}

func (s *sessionService) Join(ctx context.Context, uid int64, params *dto.SessionJoinRequest) (*dto.SessionDTO, error) {
	session, err := s.sessionRepo.GetByID(ctx, params.SessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorSessionNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	if !session.IsActive() {
		return nil, code.ErrorSessionClosed
	}

	if session.HostUID == uid {
		return nil, code.ErrorSessionCannotJoinOwn
	}

	existing, err := s.sessionRepo.GetMember(ctx, session.ID, uid)
	if err == nil && existing != nil {
		existing.Online = true
		_ = s.sessionRepo.UpdateMemberOnline(ctx, session.ID, uid, true)
		dto := s.sessionToDTO(session)
		members, _ := s.sessionRepo.ListMembers(ctx, session.ID)
		for _, m := range members {
			dto.Members = append(dto.Members, s.memberToDTO(m))
		}
		return dto, nil
	}

	member := &domain.SessionMember{
		SessionID: session.ID,
		UID:       uid,
		Role:      domain.SessionMemberRoleMember,
		Online:    true,
		JoinedAt:  time.Now(),
	}

	if _, err := s.sessionRepo.AddMember(ctx, member); err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	dto := s.sessionToDTO(session)
	members, _ := s.sessionRepo.ListMembers(ctx, session.ID)
	for _, m := range members {
		dto.Members = append(dto.Members, s.memberToDTO(m))
	}
	return dto, nil
}

func (s *sessionService) Leave(ctx context.Context, uid int64, sessionID int64) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.ErrorSessionNotFound
		}
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	if uid == session.HostUID {
		session.Status = domain.SessionStatusClosed
		session.UpdatedAt = time.Now()
		if err := s.sessionRepo.Update(ctx, session); err != nil {
			return code.ErrorDBQuery.WithDetails(err.Error())
		}
		_ = s.sessionRepo.SetAllMembersOffline(ctx, sessionID)
		return nil
	}

	return s.sessionRepo.RemoveMember(ctx, sessionID, uid)
}

func (s *sessionService) Close(ctx context.Context, uid int64, sessionID int64) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.ErrorSessionNotFound
		}
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	if uid != session.HostUID {
		return code.ErrorSessionNotHost
	}

	session.Status = domain.SessionStatusClosed
	session.UpdatedAt = time.Now()
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	_ = s.sessionRepo.SetAllMembersOffline(ctx, sessionID)
	return nil
}

func (s *sessionService) Get(ctx context.Context, uid int64, sessionID int64) (*dto.SessionDTO, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorSessionNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	member, err := s.sessionRepo.GetMember(ctx, sessionID, uid)
	if err != nil {
		return nil, code.ErrorSessionNotMember
	}
	if member == nil {
		return nil, code.ErrorSessionNotMember
	}

	dto := s.sessionToDTO(session)
	members, _ := s.sessionRepo.ListMembers(ctx, sessionID)
	for _, m := range members {
		dto.Members = append(dto.Members, s.memberToDTO(m))
	}
	return dto, nil
}

func (s *sessionService) List(ctx context.Context, uid int64) ([]*dto.SessionDTO, error) {
	sessions, err := s.sessionRepo.ListByUID(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var result []*dto.SessionDTO
	for _, session := range sessions {
		dto := s.sessionToDTO(session)
		members, _ := s.sessionRepo.ListMembers(ctx, session.ID)
		for _, m := range members {
			dto.Members = append(dto.Members, s.memberToDTO(m))
		}
		result = append(result, dto)
	}
	return result, nil
}

func (s *sessionService) SetOnline(ctx context.Context, sessionID, uid int64, online bool) error {
	return s.sessionRepo.UpdateMemberOnline(ctx, sessionID, uid, online)
}

func (s *sessionService) ListMembers(ctx context.Context, sessionID int64) ([]*dto.SessionMemberDTO, error) {
	members, err := s.sessionRepo.ListMembers(ctx, sessionID)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	var result []*dto.SessionMemberDTO
	for _, m := range members {
		result = append(result, s.memberToDTO(m))
	}
	return result, nil
}

func (s *sessionService) SetAllOffline(ctx context.Context, sessionID int64) error {
	return s.sessionRepo.SetAllMembersOffline(ctx, sessionID)
}

var _ SessionService = (*sessionService)(nil)
