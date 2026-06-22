package websocket_router

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

type SessionWSHandler struct {
	*WSHandler
}

func NewSessionWSHandler(a *app.App) *SessionWSHandler {
	return &SessionWSHandler{
		WSHandler: NewWSHandler(a),
	}
}

func (h *SessionWSHandler) HandleSignaling(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SignalingMessage{}

	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.session.handleSignaling")
		return
	}

	if params.TargetUID == 0 {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("targetUid is required"))
		return
	}

	if c.User == nil || c.User.UID == 0 {
		c.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	payload := map[string]any{
		"fromUid":   c.User.UID,
		"sessionId": params.SessionID,
		"sdp":       params.SDP,
		"ice":       params.ICE,
		"type":      params.Type,
	}

	if h.App.GetWSS() != nil {
		h.App.GetWSS().BroadcastToUser(params.TargetUID,
			code.Success.WithData(payload),
			msg.Type)
	}

	c.ToResponse(code.Success)
}

func (h *SessionWSHandler) HandleRelay(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.RelayMessage{}

	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.session.handleRelay")
		return
	}

	if c.User == nil || c.User.UID == 0 {
		c.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// Get all session members
	ctx := context.Background()
	members, err := h.App.SessionService.ListMembers(ctx, params.SessionID)
	if err != nil {
		h.App.Logger().Error("HandleRelay.ListMembers", zap.Error(err))
		c.ToResponse(code.ErrorDBQuery.WithDetails(err.Error()))
		return
	}

	// Forward payload to all other members
	if h.App.GetWSS() != nil {
		payload := map[string]any{
			"fromUid":   c.User.UID,
			"sessionId": params.SessionID,
			"reqId":     params.ReqID,
			"payload":   params.Payload,
		}
		for _, member := range members {
			if member.UID == c.User.UID {
				continue
			}
			h.App.GetWSS().BroadcastToUser(member.UID,
				code.Success.WithData(payload),
				msg.Type)
		}
	}

	c.ToResponse(code.Success)
}
