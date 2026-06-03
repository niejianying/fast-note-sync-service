package dao

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type vaultMemberRepository struct {
	dao *Dao
}

func NewVaultMemberRepository(dao *Dao) domain.VaultMemberRepository {
	return &vaultMemberRepository{dao: dao}
}

func init() {
	RegisterModel(ModelConfig{Name: "VaultMember", IsMainDB: true})
}

func (r *vaultMemberRepository) db() *gorm.DB { return r.dao.Db }

func (r *vaultMemberRepository) autoMigrate() {
	if r.db() != nil {
		_ = r.db().AutoMigrate(&model.VaultMember{})
	}
}

func (r *vaultMemberRepository) Add(ctx context.Context, m *domain.VaultMember) (*domain.VaultMember, error) {
	r.autoMigrate()
	rec := &model.VaultMember{VaultName: m.VaultName, OwnerUID: m.OwnerUID, MemberUID: m.MemberUID, CreatedAt: timex.Now()}
	err := r.db().WithContext(ctx).Create(rec).Error
	if err != nil {
		return nil, err
	}
	return &domain.VaultMember{ID: rec.ID, VaultName: rec.VaultName, OwnerUID: rec.OwnerUID, MemberUID: rec.MemberUID}, nil
}

func (r *vaultMemberRepository) Remove(ctx context.Context, vaultName string, memberUID int64) error {
	r.autoMigrate()
	return r.db().WithContext(ctx).Where("vault_name = ? AND member_uid = ?", vaultName, memberUID).Delete(&model.VaultMember{}).Error
}

func (r *vaultMemberRepository) ListByMember(ctx context.Context, memberUID int64) ([]*domain.VaultMember, error) {
	r.autoMigrate()
	var ms []model.VaultMember
	err := r.db().WithContext(ctx).Where("member_uid = ?", memberUID).Find(&ms).Error
	if err != nil {
		return nil, err
	}
	var res []*domain.VaultMember
	for _, m := range ms {
		v := m
		res = append(res, &domain.VaultMember{ID: v.ID, VaultName: v.VaultName, OwnerUID: v.OwnerUID, MemberUID: v.MemberUID})
	}
	return res, nil
}

func (r *vaultMemberRepository) ListByVault(ctx context.Context, vaultName string) ([]*domain.VaultMember, error) {
	r.autoMigrate()
	var ms []model.VaultMember
	err := r.db().WithContext(ctx).Where("vault_name = ?", vaultName).Find(&ms).Error
	if err != nil {
		return nil, err
	}
	var res []*domain.VaultMember
	for _, m := range ms {
		v := m
		res = append(res, &domain.VaultMember{ID: v.ID, VaultName: v.VaultName, OwnerUID: v.OwnerUID, MemberUID: v.MemberUID})
	}
	return res, nil
}

func (r *vaultMemberRepository) IsMember(ctx context.Context, vaultName string, uid int64) (bool, error) {
	r.autoMigrate()
	var count int64
	err := r.db().WithContext(ctx).Model(&model.VaultMember{}).Where("vault_name = ? AND (owner_uid = ? OR member_uid = ?)", vaultName, uid, uid).Count(&count).Error
	return count > 0, err
}
