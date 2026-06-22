package websocket_router

import (
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
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
