package dao

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
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

func (r *sharedVaultRepository) db() *gorm.DB {
	return r.dao.Db
}

func (r *sharedVaultRepository) autoMigrate() {
	if r.db() != nil {
		_ = r.db().AutoMigrate(&model.SharedVault{})
	}
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
	r.autoMigrate()
	m := r.toModel(sv)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := r.db().WithContext(ctx).Create(m).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *sharedVaultRepository) GetByID(ctx context.Context, id int64) (*domain.SharedVault, error) {
	r.autoMigrate()
	var m model.SharedVault
	err := r.db().WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&m), nil
}

func (r *sharedVaultRepository) Update(ctx context.Context, sv *domain.SharedVault) error {
	r.autoMigrate()
	m := r.toModel(sv)
	m.UpdatedAt = timex.Now()
	return r.db().WithContext(ctx).Model(&model.SharedVault{}).Where("id = ?", sv.ID).Updates(map[string]interface{}{
		"status":     m.Status,
		"vault_key":  m.VaultKey,
		"updated_at": m.UpdatedAt,
	}).Error
}

func (r *sharedVaultRepository) Delete(ctx context.Context, id int64) error {
	r.autoMigrate()
	return r.db().WithContext(ctx).Where("id = ?", id).Delete(&model.SharedVault{}).Error
}

func (r *sharedVaultRepository) ListByTarget(ctx context.Context, targetUID int64) ([]*domain.SharedVault, error) {
	r.autoMigrate()
	var ms []model.SharedVault
	err := r.db().WithContext(ctx).Where("target_uid = ?", targetUID).Order("created_at desc").Find(&ms).Error
	if err != nil {
		return nil, err
	}
	var res []*domain.SharedVault
	for _, m := range ms {
		v := m
		res = append(res, r.toDomain(&v))
	}
	return res, nil
}

func (r *sharedVaultRepository) ListByOwner(ctx context.Context, ownerUID int64) ([]*domain.SharedVault, error) {
	r.autoMigrate()
	var ms []model.SharedVault
	err := r.db().WithContext(ctx).Where("owner_uid = ?", ownerUID).Order("created_at desc").Find(&ms).Error
	if err != nil {
		return nil, err
	}
	var res []*domain.SharedVault
	for _, m := range ms {
		v := m
		res = append(res, r.toDomain(&v))
	}
	return res, nil
}

func (r *sharedVaultRepository) GetByOwnerAndTarget(ctx context.Context, ownerUID, targetUID int64, vaultName string) (*domain.SharedVault, error) {
	r.autoMigrate()
	var m model.SharedVault
	err := r.db().WithContext(ctx).Where("owner_uid = ? AND target_uid = ? AND vault_name = ?", ownerUID, targetUID, vaultName).First(&m).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&m), nil
}

var _ domain.SharedVaultRepository = (*sharedVaultRepository)(nil)
