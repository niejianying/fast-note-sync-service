package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"gorm.io/gorm"
)

type sharedNoteRepository struct {
	dao *Dao
}

func NewSharedNoteRepository(dao *Dao) domain.SharedNoteRepository {
	return &sharedNoteRepository{dao: dao}
}

func init() {
	RegisterModel(ModelConfig{Name: "SharedNote", IsMainDB: true})
}

func (r *sharedNoteRepository) db() *gorm.DB { return r.dao.Db }

func (r *sharedNoteRepository) autoMigrate() {
	if r.db() != nil {
		_ = r.db().AutoMigrate(&model.SharedNote{})
	}
}

func (r *sharedNoteRepository) GetByPath(ctx context.Context, vaultName, path string) (*domain.SnapFile, error) {
	r.autoMigrate()
	var m model.SharedNote
	err := r.db().WithContext(ctx).Where("vault_name = ? AND path = ?", vaultName, path).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &domain.SnapFile{Path: m.Path, PathHash: m.PathHash, Content: m.Content, ContentHash: m.ContentHash, Size: m.Size, MTime: m.MTime, CTime: m.CTime}, nil
}

func (r *sharedNoteRepository) Upsert(ctx context.Context, vaultName string, s *domain.SnapFile, uid int64) (*domain.SnapFile, error) {
	r.autoMigrate()
	now := time.Now().UnixMilli()
	m := &model.SharedNote{
		VaultName: vaultName, Path: s.Path, PathHash: s.PathHash,
		Content: s.Content, ContentHash: s.ContentHash, Size: s.Size,
		CTime: s.CTime, MTime: now, UpdatedUID: uid,
	}
	var existing model.SharedNote
	err := r.db().WithContext(ctx).Where("vault_name = ? AND path = ?", vaultName, s.Path).First(&existing).Error
	if err == nil {
		m.ID = existing.ID
		m.Version = existing.Version + 1
		m.CTime = existing.CTime
		m.CreatedUID = existing.CreatedUID
		// Preserve content/version if content hasn't changed
		if m.ContentHash == existing.ContentHash {
			m.Version = existing.Version
			m.Content = existing.Content
		}
		err = r.db().WithContext(ctx).Save(m).Error
	} else {
		m.CreatedUID = uid
		err = r.db().WithContext(ctx).Create(m).Error
	}
	if err != nil {
		return nil, err
	}
	return &domain.SnapFile{Path: m.Path, PathHash: m.PathHash, Content: m.Content, ContentHash: m.ContentHash, Size: m.Size, MTime: m.MTime, CTime: m.CTime}, nil
}

func (r *sharedNoteRepository) Delete(ctx context.Context, vaultName, path string) error {
	r.autoMigrate()
	return r.db().WithContext(ctx).Where("vault_name = ? AND path = ?", vaultName, path).Delete(&model.SharedNote{}).Error
}

func (r *sharedNoteRepository) ListByVault(ctx context.Context, vaultName string, lastTime int64) ([]*domain.SnapFile, error) {
	r.autoMigrate()
	var ms []model.SharedNote
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
		res = append(res, &domain.SnapFile{Path: v.Path, PathHash: v.PathHash, Content: v.Content, ContentHash: v.ContentHash, Size: v.Size, MTime: v.MTime, CTime: v.CTime})
	}
	return res, nil
}

func (r *sharedNoteRepository) GetLastTime(ctx context.Context, vaultName string) (int64, error) {
	r.autoMigrate()
	var m model.SharedNote
	err := r.db().WithContext(ctx).Where("vault_name = ?", vaultName).Order("mtime desc").First(&m).Error
	if err != nil {
		return 0, err
	}
	return m.MTime, nil
}

func (r *sharedNoteRepository) AddVersionMigrationColumnIfNeeded(ctx context.Context) error {
	r.autoMigrate()
	return r.db().WithContext(ctx).Exec("ALTER TABLE shared_note ADD COLUMN IF NOT EXISTS version integer DEFAULT 0").Error
}

func (r *sharedNoteRepository) GetVersionByPath(ctx context.Context, vaultName, path string) (string, error) {
	r.autoMigrate()
	var m model.SharedNote
	err := r.db().WithContext(ctx).Where("vault_name = ? AND path = ?", vaultName, path).Select("version").First(&m).Error
	if err != nil {
		return "0", err
	}
	return strconv.FormatInt(m.Version, 10), nil
}
