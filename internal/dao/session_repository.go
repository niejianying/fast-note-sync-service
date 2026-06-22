package dao

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type sessionRepository struct {
	dao *Dao
}

func NewSessionRepository(dao *Dao) domain.SessionRepository {
	return &sessionRepository{dao: dao}
}

func init() {
	RegisterModel(ModelConfig{
		Name:     "CollabSession",
		IsMainDB: true,
	})
	RegisterModel(ModelConfig{
		Name:     "SessionMember",
		IsMainDB: true,
	})
}

func (r *sessionRepository) db() *gorm.DB {
	return r.dao.Db
}

func (r *sessionRepository) sessionToDomain(m *model.CollabSession) *domain.CollabSession {
	if m == nil {
		return nil
	}
	return &domain.CollabSession{
		ID:        m.ID,
		Name:      m.Name,
		HostUID:   m.HostUID,
		Status:    domain.SessionStatus(m.Status),
		CreatedAt: time.Time(m.CreatedAt),
		UpdatedAt: time.Time(m.UpdatedAt),
	}
}

func (r *sessionRepository) sessionToModel(d *domain.CollabSession) *model.CollabSession {
	if d == nil {
		return nil
	}
	return &model.CollabSession{
		ID:        d.ID,
		Name:      d.Name,
		HostUID:   d.HostUID,
		Status:    string(d.Status),
		CreatedAt: timex.Time(d.CreatedAt),
		UpdatedAt: timex.Time(d.UpdatedAt),
	}
}

func (r *sessionRepository) memberToDomain(m *model.SessionMember) *domain.SessionMember {
	if m == nil {
		return nil
	}
	return &domain.SessionMember{
		ID:        m.ID,
		SessionID: m.SessionID,
		UID:       m.UID,
		Role:      domain.SessionMemberRole(m.Role),
		Online:    m.Online,
		JoinedAt:  time.Time(m.JoinedAt),
	}
}

func (r *sessionRepository) memberToModel(d *domain.SessionMember) *model.SessionMember {
	if d == nil {
		return nil
	}
	return &model.SessionMember{
		ID:        d.ID,
		SessionID: d.SessionID,
		UID:       d.UID,
		Role:      string(d.Role),
		Online:    d.Online,
		JoinedAt:  timex.Time(d.JoinedAt),
	}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.CollabSession) (*domain.CollabSession, error) {
	m := r.sessionToModel(session)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	if err := r.db().WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return r.sessionToDomain(m), nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id int64) (*domain.CollabSession, error) {
	var m model.CollabSession
	if err := r.db().WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return r.sessionToDomain(&m), nil
}

func (r *sessionRepository) Update(ctx context.Context, session *domain.CollabSession) error {
	m := r.sessionToModel(session)
	m.UpdatedAt = timex.Now()
	return r.db().WithContext(ctx).Model(&model.CollabSession{}).Where("id = ?", session.ID).Updates(m).Error
}

func (r *sessionRepository) ListByUID(ctx context.Context, uid int64) ([]*domain.CollabSession, error) {
	var sessions []*model.CollabSession
	err := r.db().WithContext(ctx).
		Joins("JOIN session_member ON session_member.session_id = collab_session.id").
		Where("session_member.uid = ?", uid).
		Where("collab_session.status = ?", domain.SessionStatusActive).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	var result []*domain.CollabSession
	for _, s := range sessions {
		result = append(result, r.sessionToDomain(s))
	}
	return result, nil
}

func (r *sessionRepository) AddMember(ctx context.Context, member *domain.SessionMember) (*domain.SessionMember, error) {
	m := r.memberToModel(member)
	m.JoinedAt = timex.Now()
	if err := r.db().WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return r.memberToDomain(m), nil
}

func (r *sessionRepository) RemoveMember(ctx context.Context, sessionID, uid int64) error {
	return r.db().WithContext(ctx).
		Where("session_id = ? AND uid = ?", sessionID, uid).
		Delete(&model.SessionMember{}).Error
}

func (r *sessionRepository) GetMember(ctx context.Context, sessionID, uid int64) (*domain.SessionMember, error) {
	var m model.SessionMember
	if err := r.db().WithContext(ctx).
		Where("session_id = ? AND uid = ?", sessionID, uid).
		First(&m).Error; err != nil {
		return nil, err
	}
	return r.memberToDomain(&m), nil
}

func (r *sessionRepository) ListMembers(ctx context.Context, sessionID int64) ([]*domain.SessionMember, error) {
	var members []*model.SessionMember
	if err := r.db().WithContext(ctx).
		Where("session_id = ?", sessionID).
		Find(&members).Error; err != nil {
		return nil, err
	}
	var result []*domain.SessionMember
	for _, m := range members {
		result = append(result, r.memberToDomain(m))
	}
	return result, nil
}

func (r *sessionRepository) UpdateMemberOnline(ctx context.Context, sessionID, uid int64, online bool) error {
	return r.db().WithContext(ctx).
		Model(&model.SessionMember{}).
		Where("session_id = ? AND uid = ?", sessionID, uid).
		Update("online", online).Error
}

func (r *sessionRepository) SetAllMembersOffline(ctx context.Context, sessionID int64) error {
	return r.db().WithContext(ctx).
		Model(&model.SessionMember{}).
		Where("session_id = ?", sessionID).
		Update("online", false).Error
}

func (r *sessionRepository) CountMembers(ctx context.Context, sessionID int64) (int64, error) {
	var count int64
	err := r.db().WithContext(ctx).
		Model(&model.SessionMember{}).
		Where("session_id = ?", sessionID).
		Count(&count).Error
	return count, err
}
