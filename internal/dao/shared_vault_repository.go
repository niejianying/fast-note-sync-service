package dao

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type sharedVaultRepository struct {
	dao *Dao
}

func NewSharedVaultRepository(dao *Dao) domain.SharedVaultRepository {
	return &sharedVaultRepository{dao: dao}
}

func (r *sharedVaultRepository) GetKey(uid int64) string {
	return ""
}

func init() {
	RegisterModel(ModelConfig{
		Name:     "SharedVault",
		IsMainDB: true,
	})
}

func (r *sharedVaultRepository) query() *query.Query {
	return r.dao.QueryWithOnceInit(func(g *gorm.DB) {
		model.AutoMigrate(g, "SharedVault")
	}, "shared_vault#default")
}

func (r *sharedVaultRepository) toDomain(m *model.SharedVault) *domain.SharedVault {
	if m == nil {
		return nil
	}
	return &domain.SharedVault{
		ID:        m.ID,
		VaultName: m.VaultName,
		OwnerUID:  m.OwnerUID,
		TargetUID: m.TargetUID,
		VaultKey:  m.VaultKey,
		Status:    domain.SharedVaultStatus(m.Status),
		CreatedAt: time.Time(m.CreatedAt),
		UpdatedAt: time.Time(m.UpdatedAt),
	}
}

func (r *sharedVaultRepository) toModel(d *domain.SharedVault) *model.SharedVault {
	if d == nil {
		return nil
	}
	return &model.SharedVault{
		ID:        d.ID,
		VaultName: d.VaultName,
		OwnerUID:  d.OwnerUID,
		TargetUID: d.TargetUID,
		VaultKey:  d.VaultKey,
		Status:    string(d.Status),
		CreatedAt: timex.Time(d.CreatedAt),
		UpdatedAt: timex.Time(d.UpdatedAt),
	}
}

func (r *sharedVaultRepository) Create(ctx context.Context, sv *domain.SharedVault) (*domain.SharedVault, error) {
	q := r.query().SharedVault
	m := r.toModel(sv)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := q.WithContext(ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *sharedVaultRepository) GetByID(ctx context.Context, id int64) (*domain.SharedVault, error) {
	q := r.query().SharedVault
	m, err := q.WithContext(ctx).Where(q.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *sharedVaultRepository) Update(ctx context.Context, sv *domain.SharedVault) error {
	q := r.query().SharedVault
	m := r.toModel(sv)
	m.UpdatedAt = timex.Now()
	_, err := q.WithContext(ctx).Where(q.ID.Eq(sv.ID)).Updates(m)
	return err
}

func (r *sharedVaultRepository) Delete(ctx context.Context, id int64) error {
	q := r.query().SharedVault
	_, err := q.WithContext(ctx).Where(q.ID.Eq(id)).Delete()
	return err
}

func (r *sharedVaultRepository) ListByTarget(ctx context.Context, targetUID int64) ([]*domain.SharedVault, error) {
	q := r.query().SharedVault
	ms, err := q.WithContext(ctx).Where(q.TargetUID.Eq(targetUID)).Order(q.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.SharedVault
	for _, m := range ms {
		res = append(res, r.toDomain(m))
	}
	return res, nil
}

func (r *sharedVaultRepository) ListByOwner(ctx context.Context, ownerUID int64) ([]*domain.SharedVault, error) {
	q := r.query().SharedVault
	ms, err := q.WithContext(ctx).Where(q.OwnerUID.Eq(ownerUID)).Order(q.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.SharedVault
	for _, m := range ms {
		res = append(res, r.toDomain(m))
	}
	return res, nil
}

func (r *sharedVaultRepository) GetByOwnerAndTarget(ctx context.Context, ownerUID, targetUID int64, vaultName string) (*domain.SharedVault, error) {
	q := r.query().SharedVault
	m, err := q.WithContext(ctx).Where(q.OwnerUID.Eq(ownerUID), q.TargetUID.Eq(targetUID), q.VaultName.Eq(vaultName)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

var _ domain.SharedVaultRepository = (*sharedVaultRepository)(nil)
