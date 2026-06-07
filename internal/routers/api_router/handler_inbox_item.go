package api_router

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"go.uber.org/zap"
)

type InboxItemHandler struct {
	*Handler
}

func NewInboxItemHandler(a *app.App) *InboxItemHandler {
	return &InboxItemHandler{
		Handler: NewHandler(a),
	}
}

func (h *InboxItemHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

// ListItems GET /api/inbox/items
func (h *InboxItemHandler) ListItems(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	ctx := c.Request.Context()
	result, err := h.App.InboxItemService.ListItems(ctx, uid, page, pageSize)
	if err != nil {
		h.logError(ctx, "InboxItemHandler.ListItems", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, result.List, int(result.Total))
}

// CreateItem POST /api/inbox/item
func (h *InboxItemHandler) CreateItem(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.InboxItemCreateRequest{}

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	result, err := h.App.InboxItemService.AddItem(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "InboxItemHandler.CreateItem", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(result))
}

// MarkRead POST /api/inbox/item/:id/read
func (h *InboxItemHandler) MarkRead(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("item id is required"))
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid item id"))
		return
	}

	ctx := c.Request.Context()
	if err := h.App.InboxItemService.MarkRead(ctx, id, uid); err != nil {
		h.logError(ctx, "InboxItemHandler.MarkRead", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

// MarkAllRead POST /api/inbox/items/read-all
func (h *InboxItemHandler) MarkAllRead(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	if err := h.App.InboxItemService.MarkAllRead(ctx, uid); err != nil {
		h.logError(ctx, "InboxItemHandler.MarkAllRead", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

// DeleteItem DELETE /api/inbox/item/:id
func (h *InboxItemHandler) DeleteItem(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("item id is required"))
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid item id"))
		return
	}

	ctx := c.Request.Context()
	if err := h.App.InboxItemService.DeleteItem(ctx, id, uid); err != nil {
		h.logError(ctx, "InboxItemHandler.DeleteItem", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}
