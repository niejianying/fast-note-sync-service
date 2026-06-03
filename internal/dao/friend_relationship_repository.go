package dao

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type friendRelationshipRepository struct {
	dao *Dao
}

func NewFriendRelationshipRepository(dao *Dao) domain.FriendRelationshipRepository {
	return &friendRelationshipRepository{dao: dao}
}

func (r *friendRelationshipRepository) GetKey(uid int64) string {
	return ""
}

func init() {
	RegisterModel(ModelConfig{
		Name:     "FriendRelationship",
		IsMainDB: true,
	})
}

func (r *friendRelationshipRepository) friendQuery() *query.Query {
	return r.dao.QueryWithOnceInit(func(g *gorm.DB) {
		model.AutoMigrate(g, "FriendRelationship")
	}, "friend_relationship#default")
}

func (r *friendRelationshipRepository) toDomain(m *model.FriendRelationship) *domain.FriendRelationship {
	if m == nil {
		return nil
	}
	return &domain.FriendRelationship{
		ID:        m.ID,
		UID:       m.UID,
		FriendUID: m.FriendUID,
		Status:    domain.FriendStatus(m.Status),
		CreatedAt: time.Time(m.CreatedAt),
		UpdatedAt: time.Time(m.UpdatedAt),
		DeletedAt: time.Time(m.DeletedAt),
	}
}

func (r *friendRelationshipRepository) toModel(d *domain.FriendRelationship) *model.FriendRelationship {
	if d == nil {
		return nil
	}
	return &model.FriendRelationship{
		ID:        d.ID,
		UID:       d.UID,
		FriendUID: d.FriendUID,
		Status:    string(d.Status),
		CreatedAt: timex.Time(d.CreatedAt),
		UpdatedAt: timex.Time(d.UpdatedAt),
		DeletedAt: timex.Time(d.DeletedAt),
	}
}

func (r *friendRelationshipRepository) GetByID(ctx context.Context, id int64) (*domain.FriendRelationship, error) {
	q := r.friendQuery().FriendRelationship
	m, err := q.WithContext(ctx).Where(q.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *friendRelationshipRepository) GetByUIDAndFriend(ctx context.Context, uid, friendUID int64) (*domain.FriendRelationship, error) {
	q := r.friendQuery().FriendRelationship
	m, err := q.WithContext(ctx).Where(q.UID.Eq(uid), q.FriendUID.Eq(friendUID)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *friendRelationshipRepository) GetReverse(ctx context.Context, uid, friendUID int64) (*domain.FriendRelationship, error) {
	q := r.friendQuery().FriendRelationship
	m, err := q.WithContext(ctx).Where(q.UID.Eq(friendUID), q.FriendUID.Eq(uid)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *friendRelationshipRepository) Create(ctx context.Context, rel *domain.FriendRelationship) (*domain.FriendRelationship, error) {
	q := r.friendQuery().FriendRelationship
	m := r.toModel(rel)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := q.WithContext(ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *friendRelationshipRepository) Update(ctx context.Context, rel *domain.FriendRelationship) error {
	q := r.friendQuery().FriendRelationship
	m := r.toModel(rel)
	m.UpdatedAt = timex.Now()
	_, err := q.WithContext(ctx).Where(q.ID.Eq(rel.ID)).Updates(m)
	return err
}

func (r *friendRelationshipRepository) Delete(ctx context.Context, uid, friendUID int64) error {
	q := r.friendQuery().FriendRelationship
	_, err := q.WithContext(ctx).Where(q.UID.Eq(uid), q.FriendUID.Eq(friendUID)).Delete()
	return err
}

func (r *friendRelationshipRepository) ListByUID(ctx context.Context, uid int64) ([]*domain.FriendRelationship, error) {
	q := r.friendQuery().FriendRelationship
	ms, err := q.WithContext(ctx).Where(q.UID.Eq(uid)).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.FriendRelationship
	for _, m := range ms {
		res = append(res, r.toDomain(m))
	}
	return res, nil
}

func (r *friendRelationshipRepository) ListPendingByUID(ctx context.Context, uid int64) ([]*domain.FriendRelationship, error) {
	q := r.friendQuery().FriendRelationship
	ms, err := q.WithContext(ctx).Where(q.FriendUID.Eq(uid), q.Status.Eq("pending")).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.FriendRelationship
	for _, m := range ms {
		res = append(res, r.toDomain(m))
	}
	return res, nil
}

var _ domain.FriendRelationshipRepository = (*friendRelationshipRepository)(nil)
