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

type SessionHandler struct {
	*Handler
}

func NewSessionHandler(a *app.App) *SessionHandler {
	return &SessionHandler{
		Handler: NewHandler(a),
	}
}

func (h *SessionHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

func (h *SessionHandler) Create(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.SessionCreateRequest{}

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("SessionHandler.Create.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	result, err := h.App.SessionService.Create(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "SessionHandler.Create", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(result))
}

func (h *SessionHandler) Join(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.SessionJoinRequest{}

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("SessionHandler.Join.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	result, err := h.App.SessionService.Join(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "SessionHandler.Join", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(result))
}

func (h *SessionHandler) Leave(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	sessionIDStr := c.Param("id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid session id"))
		return
	}

	ctx := c.Request.Context()
	if err := h.App.SessionService.Leave(ctx, uid, sessionID); err != nil {
		h.logError(ctx, "SessionHandler.Leave", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

func (h *SessionHandler) Close(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	sessionIDStr := c.Param("id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid session id"))
		return
	}

	ctx := c.Request.Context()
	if err := h.App.SessionService.Close(ctx, uid, sessionID); err != nil {
		h.logError(ctx, "SessionHandler.Close", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

func (h *SessionHandler) Get(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	sessionIDStr := c.Param("id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid session id"))
		return
	}

	ctx := c.Request.Context()
	result, err := h.App.SessionService.Get(ctx, uid, sessionID)
	if err != nil {
		h.logError(ctx, "SessionHandler.Get", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(result))
}

func (h *SessionHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	list, err := h.App.SessionService.List(ctx, uid)
	if err != nil {
		h.logError(ctx, "SessionHandler.List", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(list))
}
