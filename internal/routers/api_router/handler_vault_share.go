package api_router

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/websocket_router"
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

func (h *VaultShareHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
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

	// Notify the target user about the vault share invitation
	if h.App.GetWSS() != nil {
		h.App.GetWSS().BroadcastToUser(params.FriendUID,
			code.Success.WithData(map[string]interface{}{
				"id":        result.ID,
				"vaultName": result.VaultName,
				"fromUid":   uid,
			}),
			websocket_router.ShareSyncRefresh)
	}

	// Persist inbox item for the recipient
	if h.App.InboxItemService != nil {
		h.createVaultShareInboxItem(ctx, params.FriendUID,
			"vault_share_"+strconv.FormatInt(result.ID, 10),
			result.VaultName, uid)
	}
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

	if result == nil {
		return
	}

	// If rejected, notify the owner
	if h.App.GetWSS() != nil && !params.Accept {
		h.App.GetWSS().BroadcastToUser(result.OwnerUID,
			code.Success.WithData(map[string]interface{}{
				"id":        result.ID,
				"vaultName": result.VaultName,
			}),
			websocket_router.ShareSyncRejected)
	}

	// Persist inbox item
	if h.App.InboxItemService != nil {
		if params.Accept {
			h.createVaultShareInboxItem(ctx, result.OwnerUID,
				"vault_share_accepted_"+strconv.FormatInt(result.ID, 10),
				result.VaultName, uid)
		} else {
			h.createVaultShareInboxItem(ctx, result.OwnerUID,
				"vault_share_rejected_"+strconv.FormatInt(result.ID, 10),
				result.VaultName, uid)
		}
	}
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

func (h *VaultShareHandler) createVaultShareInboxItem(ctx context.Context, targetUID int64, itemID, vaultName string, fromUID int64) {
	if h.App.InboxItemService == nil {
		return
	}
	// Extract share ID from itemID (e.g. "vault_share_5")
	shareID := int64(0)
	if parts := strings.Split(itemID, "_"); len(parts) > 0 {
		if id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64); err == nil {
			shareID = id
		}
	}
	_, _ = h.App.InboxItemService.AddItem(ctx, targetUID, &dto.InboxItemCreateRequest{
		ItemID:   itemID,
		Type:     "vaultShare",
		Title:    "仓库共享邀请",
		Subtitle: "用户 #" + strconv.FormatInt(fromUID, 10) + " 邀请你共享 \"" + vaultName + "\"",
		Payload:  fmt.Sprintf(`{"id":%d,"vaultName":"%s","fromUid":%d}`, shareID, vaultName, fromUID),
	})
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
