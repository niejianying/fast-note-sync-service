package api_router

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"go.uber.org/zap"
)

type VaultShareHandler struct {
	*Handler
}

func NewVaultShareHandler(a *app.App) *VaultShareHandler {
	return &VaultShareHandler{Handler: NewHandler(a)}
}

func (h *VaultShareHandler) Share(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.VaultShareRequest{}
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	result, err := h.App.VaultShareService.Share(ctx, uid, params, params.VaultKey)
	if err != nil {
		h.logError(ctx, "VaultShareHandler.Share", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(result))
}

func (h *VaultShareHandler) Respond(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.VaultShareRespondRequest{}
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid id"))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	result, err := h.App.VaultShareService.Respond(ctx, uid, id, params)
	if err != nil {
		h.logError(ctx, "VaultShareHandler.Respond", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(result))
}

func (h *VaultShareHandler) Revoke(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid id"))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	if err := h.App.VaultShareService.Revoke(ctx, uid, id); err != nil {
		h.logError(ctx, "VaultShareHandler.Revoke", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

func (h *VaultShareHandler) ListIncoming(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	list, err := h.App.VaultShareService.ListIncoming(ctx, uid)
	if err != nil {
		h.logError(ctx, "VaultShareHandler.ListIncoming", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(list))
}

func (h *VaultShareHandler) ListOutgoing(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	list, err := h.App.VaultShareService.ListOutgoing(ctx, uid)
	if err != nil {
		h.logError(ctx, "VaultShareHandler.ListOutgoing", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(list))
}
