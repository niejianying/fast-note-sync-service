package api_router

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"go.uber.org/zap"
)

type VaultMemberHandler struct {
	*Handler
}

func NewVaultMemberHandler(a *app.App) *VaultMemberHandler {
	return &VaultMemberHandler{Handler: NewHandler(a)}
}

func (h *VaultMemberHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

// ListMembers GET /api/vault/:name/members
func (h *VaultMemberHandler) ListMembers(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	vaultName := c.Param("name")
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()

	isMember, err := h.App.VaultMemberRepo.IsMember(ctx, vaultName, uid)
	if err != nil {
		h.logError(ctx, "VaultMemberHandler.ListMembers.IsMember", err)
		apperrors.ErrorResponse(c, err)
		return
	}
	if !isMember {
		response.ToResponse(code.ErrorAuthTokenScopeRestricted)
		return
	}

	members, err := h.App.VaultMemberRepo.ListByVault(ctx, vaultName)
	if err != nil {
		h.logError(ctx, "VaultMemberHandler.ListMembers.ListByVault", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	var result []*dto.VaultMemberDTO
	for _, m := range members {
		user, err := h.App.UserRepo.GetByUID(ctx, m.MemberUID, true)
		if err != nil {
			h.logError(ctx, "VaultMemberHandler.ListMembers.GetByUID", err)
			apperrors.ErrorResponse(c, err)
			return
		}

		role := "member"
		if m.OwnerUID == m.MemberUID {
			role = "owner"
		}

		result = append(result, &dto.VaultMemberDTO{
			UID:      m.MemberUID,
			Nickname: user.Username,
			Role:     role,
		})
	}

	response.ToResponse(code.Success.WithData(result))
}

// RemoveMember DELETE /api/vault/:name/members/:uid
func (h *VaultMemberHandler) RemoveMember(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	vaultName := c.Param("name")
	targetUIDStr := c.Param("uid")
	targetUID, err := strconv.ParseInt(targetUIDStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid uid"))
		return
	}

	currentUID := pkgapp.GetUID(c)
	if currentUID == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	ctx := c.Request.Context()

	members, err := h.App.VaultMemberRepo.ListByVault(ctx, vaultName)
	if err != nil {
		h.logError(ctx, "VaultMemberHandler.RemoveMember.ListByVault", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	var currentMember *domain.VaultMember
	var targetMember *domain.VaultMember
	for _, m := range members {
		if m.MemberUID == currentUID {
			currentMember = m
		}
		if m.MemberUID == targetUID {
			targetMember = m
		}
	}

	if currentMember == nil {
		response.ToResponse(code.ErrorAuthTokenScopeRestricted)
		return
	}

	isOwner := currentMember.OwnerUID == currentMember.MemberUID

	if currentUID == targetUID {
		if isOwner {
			response.ToResponse(code.ErrorAuthTokenScopeRestricted)
			return
		}
	} else {
		if !isOwner {
			response.ToResponse(code.ErrorAuthTokenScopeRestricted)
			return
		}
	}

	if err := h.App.VaultMemberRepo.Remove(ctx, vaultName, targetUID); err != nil {
		h.logError(ctx, "VaultMemberHandler.RemoveMember.Remove", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	if targetMember != nil {
		shares, err := h.App.SharedVaultRepo.ListByOwner(ctx, targetMember.OwnerUID)
		if err != nil {
			h.logError(ctx, "VaultMemberHandler.RemoveMember.ListByOwner", err)
			apperrors.ErrorResponse(c, err)
			return
		}
		for _, s := range shares {
			if s.VaultName == vaultName && s.TargetUID == targetUID {
				if err := h.App.SharedVaultRepo.Delete(ctx, s.ID); err != nil {
					h.logError(ctx, "VaultMemberHandler.RemoveMember.Delete", err)
					apperrors.ErrorResponse(c, err)
					return
				}
				break
			}
		}
	}

	response.ToResponse(code.Success)
}
