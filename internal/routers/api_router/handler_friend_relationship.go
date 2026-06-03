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

type FriendRelationshipHandler struct {
	*Handler
}

func NewFriendRelationshipHandler(a *app.App) *FriendRelationshipHandler {
	return &FriendRelationshipHandler{
		Handler: NewHandler(a),
	}
}

func (h *FriendRelationshipHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

func (h *FriendRelationshipHandler) AddFriend(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FriendRequestAdd{}

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("FriendRelationshipHandler.AddFriend.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	result, err := h.App.FriendRelationshipService.AddFriend(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "FriendRelationshipHandler.AddFriend", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(result))
}

func (h *FriendRelationshipHandler) RespondToRequest(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FriendRequestRespond{}

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("FriendRelationshipHandler.RespondToRequest.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	result, err := h.App.FriendRelationshipService.RespondToRequest(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "FriendRelationshipHandler.RespondToRequest", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(result))
}

func (h *FriendRelationshipHandler) RemoveFriend(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	friendUIDStr := c.Param("uid")
	if friendUIDStr == "" {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("friend uid is required"))
		return
	}
	friendUID, err := strconv.ParseInt(friendUIDStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid friend uid"))
		return
	}

	ctx := c.Request.Context()
	if err := h.App.FriendRelationshipService.RemoveFriend(ctx, uid, friendUID); err != nil {
		h.logError(ctx, "FriendRelationshipHandler.RemoveFriend", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

func (h *FriendRelationshipHandler) ListFriends(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	list, err := h.App.FriendRelationshipService.ListFriends(ctx, uid)
	if err != nil {
		h.logError(ctx, "FriendRelationshipHandler.ListFriends", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(list))
}

func (h *FriendRelationshipHandler) ListPendingRequests(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	list, err := h.App.FriendRelationshipService.ListPendingRequests(ctx, uid)
	if err != nil {
		h.logError(ctx, "FriendRelationshipHandler.ListPendingRequests", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(list))
}
