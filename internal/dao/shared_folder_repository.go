package dao

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"gorm.io/gorm"
)

type sharedFolderRepository struct {
	dao *Dao
}

func NewSharedFolderRepository(dao *Dao) domain.SharedFolderRepository {
	return &sharedFolderRepository{dao: dao}
}

func init() {
	RegisterModel(ModelConfig{Name: "SharedFolder", IsMainDB: true})
}

func (r *sharedFolderRepository) db() *gorm.DB { return r.dao.Db }

func (r *sharedFolderRepository) autoMigrate() {
	if r.db() != nil {
		_ = r.db().AutoMigrate(&model.SharedFolder{})
	}
}

func (r *sharedFolderRepository) Upsert(ctx context.Context, vaultName, path, pathHash string, mtime int64, uid int64) error {
	r.autoMigrate()
	var existing model.SharedFolder
	err := r.db().WithContext(ctx).Where("vault_name = ? AND path = ?", vaultName, path).First(&existing).Error
	if err == nil {
		return r.db().WithContext(ctx).Model(&existing).Updates(map[string]interface{}{"mtime": mtime, "updated_uid": uid}).Error
	}
	return r.db().WithContext(ctx).Create(&model.SharedFolder{
		VaultName: vaultName, Path: path, PathHash: pathHash,
		Action: "create", MTime: mtime, CTime: mtime, UpdatedUID: uid,
	}).Error
}

func (r *sharedFolderRepository) Delete(ctx context.Context, vaultName, path string) error {
	r.autoMigrate()
	return r.db().WithContext(ctx).Where("vault_name = ? AND path = ?", vaultName, path).Delete(&model.SharedFolder{}).Error
}

func (r *sharedFolderRepository) ListByVault(ctx context.Context, vaultName string) ([]*string, error) {
	r.autoMigrate()
	var ms []model.SharedFolder
	err := r.db().WithContext(ctx).Where("vault_name = ?", vaultName).Find(&ms).Error
	if err != nil {
		return nil, err
	}
	var res []*string
	for _, m := range ms {
		v := m.Path
		res = append(res, &v)
	}
	return res, nil
}
