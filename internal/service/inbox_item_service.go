package service

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
)

type InboxItemService interface {
	AddItem(ctx context.Context, uid int64, params *dto.InboxItemCreateRequest) (*dto.InboxItemDTO, error)
	ListItems(ctx context.Context, uid int64, page, pageSize int) (*dto.InboxItemListDTO, error)
	MarkRead(ctx context.Context, id int64, uid int64) error
	MarkAllRead(ctx context.Context, uid int64) error
	DeleteItem(ctx context.Context, id int64, uid int64) error
}

type inboxItemService struct {
	itemRepo domain.InboxItemRepository
	logger   *zap.Logger
}

func NewInboxItemService(itemRepo domain.InboxItemRepository, logger *zap.Logger) InboxItemService {
	return &inboxItemService{
		itemRepo: itemRepo,
		logger:   logger,
	}
}

func (s *inboxItemService) domainToDTO(d *domain.InboxItem) *dto.InboxItemDTO {
	if d == nil {
		return nil
	}
	return &dto.InboxItemDTO{
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
	}
}

func (s *inboxItemService) AddItem(ctx context.Context, uid int64, params *dto.InboxItemCreateRequest) (*dto.InboxItemDTO, error) {
	// Deduplicate by itemID + uid
	existing, err := s.itemRepo.GetByItemID(ctx, params.ItemID, uid)
	if err == nil && existing != nil {
		return s.domainToDTO(existing), nil
	}

	item := &domain.InboxItem{
		ItemID:         params.ItemID,
		UID:            uid,
		Type:           params.Type,
		Title:          params.Title,
		Subtitle:       params.Subtitle,
		Payload:        params.Payload,
		SourceNotePath: params.SourceNotePath,
		SourceLine:     params.SourceLine,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	created, err := s.itemRepo.Create(ctx, item)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(created), nil
}

func (s *inboxItemService) ListItems(ctx context.Context, uid int64, page, pageSize int) (*dto.InboxItemListDTO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := s.itemRepo.ListByUID(ctx, uid, page, pageSize)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var list []*dto.InboxItemDTO
	for _, item := range items {
		list = append(list, s.domainToDTO(item))
	}

	return &dto.InboxItemListDTO{
		List:  list,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

func (s *inboxItemService) MarkRead(ctx context.Context, id int64, uid int64) error {
	if err := s.itemRepo.MarkRead(ctx, id, uid); err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

func (s *inboxItemService) MarkAllRead(ctx context.Context, uid int64) error {
	if err := s.itemRepo.MarkAllRead(ctx, uid); err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

func (s *inboxItemService) DeleteItem(ctx context.Context, id int64, uid int64) error {
	if err := s.itemRepo.Delete(ctx, id, uid); err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}
