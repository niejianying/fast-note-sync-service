package dao

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"gorm.io/gorm"
)

type sharedFileRepository struct {
	dao *Dao
}

func NewSharedFileRepository(dao *Dao) domain.SharedFileRepository {
	return &sharedFileRepository{dao: dao}
}

func init() {
	RegisterModel(ModelConfig{Name: "SharedFile", IsMainDB: true})
}

func (r *sharedFileRepository) db() *gorm.DB { return r.dao.Db }

func (r *sharedFileRepository) autoMigrate() {
	if r.db() != nil {
		_ = r.db().AutoMigrate(&model.SharedFile{})
	}
}

func (r *sharedFileRepository) Upsert(ctx context.Context, vaultName string, s *domain.SnapFile, uid int64) error {
	r.autoMigrate()
	var existing model.SharedFile
	err := r.db().WithContext(ctx).Where("vault_name = ? AND path = ?", vaultName, s.Path).First(&existing).Error
	if err == nil {
		return r.db().WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"content_hash": s.ContentHash, "size": s.Size, "mtime": s.MTime,
			"updated_uid": uid, "save_path": s.Content,
		}).Error
	}
	return r.db().WithContext(ctx).Create(&model.SharedFile{
		VaultName: vaultName, Path: s.Path, PathHash: s.PathHash,
		ContentHash: s.ContentHash, Size: s.Size, SavePath: s.Content,
		CTime: s.CTime, MTime: s.MTime, UpdatedUID: uid,
	}).Error
}

func (r *sharedFileRepository) Delete(ctx context.Context, vaultName, path string) error {
	r.autoMigrate()
	return r.db().WithContext(ctx).Where("vault_name = ? AND path = ?", vaultName, path).Delete(&model.SharedFile{}).Error
}

func (r *sharedFileRepository) ListByVault(ctx context.Context, vaultName string, lastTime int64) ([]*domain.SnapFile, error) {
	r.autoMigrate()
	var ms []model.SharedFile
	q := r.db().WithContext(ctx).Where("vault_name = ?", vaultName)
	if lastTime > 0 {
		q = q.Where("mtime > ?", lastTime)
	}
	err := q.Find(&ms).Error
	if err != nil {
		return nil, err
	}
	var res []*domain.SnapFile
	for _, m := range ms {
		v := m
		res = append(res, &domain.SnapFile{Path: v.Path, PathHash: v.PathHash, Content: v.SavePath, ContentHash: v.ContentHash, Size: v.Size, MTime: v.MTime, CTime: v.CTime})
	}
	return res, nil
}
