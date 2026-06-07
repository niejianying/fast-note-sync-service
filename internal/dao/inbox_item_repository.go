package dao

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type inboxItemRepository struct {
	dao *Dao
}

func NewInboxItemRepository(dao *Dao) domain.InboxItemRepository {
	return &inboxItemRepository{dao: dao}
}

func init() {
	RegisterModel(ModelConfig{
		Name:     "InboxItem",
		IsMainDB: true,
	})
}

func (r *inboxItemRepository) db() *gorm.DB {
	return r.dao.Db
}

func (r *inboxItemRepository) toDomain(m *model.InboxItem) *domain.InboxItem {
	if m == nil {
		return nil
	}
	return &domain.InboxItem{
		ID:             m.ID,
		ItemID:         m.ItemID,
		UID:            m.UID,
		Type:           m.Type,
		Title:          m.Title,
		Subtitle:       m.Subtitle,
		Payload:        m.Payload,
		SourceNotePath: m.SourceNotePath,
		SourceLine:     m.SourceLine,
		IsRead:         m.IsRead,
		CreatedAt:      time.Time(m.CreatedAt),
		UpdatedAt:      time.Time(m.UpdatedAt),
		DeletedAt:      time.Time(m.DeletedAt),
	}
}

func (r *inboxItemRepository) toModel(d *domain.InboxItem) *model.InboxItem {
	if d == nil {
		return nil
	}
	return &model.InboxItem{
		ID:             d.ID,
		ItemID:         d.ItemID,
		UID:            d.UID,
		Type:           d.Type,
		Title:          d.Title,
		Subtitle:       d.Subtitle,
		Payload:        d.Payload,
		SourceNotePath: d.SourceNotePath,
		SourceLine:     d.SourceLine,
		IsRead:         d.IsRead,
		CreatedAt:      timex.Time(d.CreatedAt),
		UpdatedAt:      timex.Time(d.UpdatedAt),
		DeletedAt:      timex.Time(d.DeletedAt),
	}
}

func (r *inboxItemRepository) Create(ctx context.Context, item *domain.InboxItem) (*domain.InboxItem, error) {
	m := r.toModel(item)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	if err := r.db().WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *inboxItemRepository) GetByItemID(ctx context.Context, itemID string, uid int64) (*domain.InboxItem, error) {
	var m model.InboxItem
	err := r.db().WithContext(ctx).Where("item_id = ? AND uid = ?", itemID, uid).First(&m).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&m), nil
}

func (r *inboxItemRepository) GetByID(ctx context.Context, id int64, uid int64) (*domain.InboxItem, error) {
	var m model.InboxItem
	err := r.db().WithContext(ctx).Where("id = ? AND uid = ?", id, uid).First(&m).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&m), nil
}

func (r *inboxItemRepository) ListByUID(ctx context.Context, uid int64, page, pageSize int) ([]*domain.InboxItem, int64, error) {
	var ms []model.InboxItem
	var total int64

	db := r.db().WithContext(ctx).Where("uid = ?", uid)

	if err := db.Model(&model.InboxItem{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&ms).Error; err != nil {
		return nil, 0, err
	}

	var res []*domain.InboxItem
	for i := range ms {
		res = append(res, r.toDomain(&ms[i]))
	}
	return res, total, nil
}

func (r *inboxItemRepository) MarkRead(ctx context.Context, id int64, uid int64) error {
	return r.db().WithContext(ctx).Model(&model.InboxItem{}).
		Where("id = ? AND uid = ?", id, uid).
		Update("is_read", true).Error
}

func (r *inboxItemRepository) MarkAllRead(ctx context.Context, uid int64) error {
	return r.db().WithContext(ctx).Model(&model.InboxItem{}).
		Where("uid = ? AND is_read = ?", uid, false).
		Update("is_read", true).Error
}

func (r *inboxItemRepository) Delete(ctx context.Context, id int64, uid int64) error {
	return r.db().WithContext(ctx).Where("id = ? AND uid = ?", id, uid).
		Delete(&model.InboxItem{}).Error
}
