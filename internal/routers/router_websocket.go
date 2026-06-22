package routers

import (
	"context"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/websocket_router"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

func initWebSocketRoutes(wss *pkgapp.WebsocketServer, appContainer *app.App) {
	// Register Protobuf Hooks
	// 注册 Protobuf 编解码钩子
	websocket_router.RegisterProtobufHooks(wss)

	// Create WebSocket Handlers (injected App Container)
	// 创建 WebSocket Handlers（注入 App Container）
	noteWSHandler := websocket_router.NewNoteWSHandler(appContainer)
	folderWSHandler := websocket_router.NewFolderWSHandler(appContainer)
	fileWSHandler := websocket_router.NewFileWSHandler(appContainer)
	settingWSHandler := websocket_router.NewSettingWSHandler(appContainer)
	sessionWSHandler := websocket_router.NewSessionWSHandler(appContainer)

	// Note
	wss.Use(websocket_router.NoteReceiveModify, noteWSHandler.NoteModify)
	wss.Use(websocket_router.NoteReceiveDelete, noteWSHandler.NoteDelete)
	wss.Use(websocket_router.NoteReceiveRename, noteWSHandler.NoteRename)
	wss.Use(websocket_router.NoteReceiveRePush, noteWSHandler.NoteRePush)
	wss.Use(websocket_router.NoteReceiveCheck, noteWSHandler.NoteModifyCheck)
	wss.Use(websocket_router.NoteReceiveSync, noteWSHandler.NoteSync)

	// Folder
	wss.Use(websocket_router.FolderReceiveSync, folderWSHandler.FolderSync)
	wss.Use(websocket_router.FolderReceiveModify, folderWSHandler.FolderModify)
	wss.Use(websocket_router.FolderReceiveDelete, folderWSHandler.FolderDelete)
	wss.Use(websocket_router.FolderReceiveRename, folderWSHandler.FolderRename)

	// Setting
	wss.Use(websocket_router.SettingReceiveModify, settingWSHandler.SettingModify)
	wss.Use(websocket_router.SettingReceiveDelete, settingWSHandler.SettingDelete)
	wss.Use(websocket_router.SettingReceiveCheck, settingWSHandler.SettingModifyCheck)
	wss.Use(websocket_router.SettingReceiveSync, settingWSHandler.SettingSync)
	wss.Use(websocket_router.SettingReceiveClear, settingWSHandler.SettingClear)
	wss.Use(websocket_router.SettingReceiveRePush, settingWSHandler.SettingRePush)

	// Attachment
	wss.Use(websocket_router.FileReceiveSync, fileWSHandler.FileSync)
	wss.Use(websocket_router.FileReceiveUploadCheck, fileWSHandler.FileUploadCheck)
	wss.Use(websocket_router.FileReceiveRename, fileWSHandler.FileRename)
	wss.Use(websocket_router.FileReceiveDelete, fileWSHandler.FileDelete)
	wss.Use(websocket_router.FileReceiveChunkDownload, fileWSHandler.FileChunkDownload)
	wss.Use(websocket_router.FileReceiveRePush, fileWSHandler.FileRePush)

	// Attachment chunk upload
	wss.UseBinary(websocket_router.VaultFileMsgType, fileWSHandler.FileUploadChunkBinary)

	// Collaboration Session Signaling
	wss.Use(websocket_router.SessionReceiveWebRTCOffer, sessionWSHandler.HandleSignaling)
	wss.Use(websocket_router.SessionReceiveWebRTCAnswer, sessionWSHandler.HandleSignaling)
	wss.Use(websocket_router.SessionReceiveWebRTCICECandidate, sessionWSHandler.HandleSignaling)

	// Inject Message Interceptor to handle unauthenticated checks, Vault restrictions, RBAC checks, and error rollbacks
	// 注入消息拦截器，处理未登录验证、Vault笔记库限制校验、RBAC权限检查以及写失败回滚机制
	wss.UseInterceptor(websocket_router.NewMessageInterceptor(appContainer))

	wss.UseUserVerify(noteWSHandler.UserInfo)

	// Inject Token Verification to decouple pkg/app from internal/service
	wss.UseTokenVerify(func(ctx context.Context, uid, tokenID int64, nonce string, reqClientType, reqClientName, reqClientVersion, reqUserAgent, reqIP string) (string, string, error) {
		dbToken, err := appContainer.TokenService.GetActiveToken(ctx, uid, tokenID)
		if err != nil || dbToken == nil {
			fmt.Printf("[WSDebug] Token not found or invalid in DB: uid=%d, tokenId=%d, err=%v\n", uid, tokenID, err)
			if err != nil {
				return "", "", err
			}
			return "", "", code.ErrorInvalidUserAuthToken
		}

		// 0. Verify Nonce (Generation Check)
		// 校验 Nonce（世代校验），如果数据库中有记录且不匹配，说明该令牌已被轮换或失效
		if dbToken.TokenString != "" && nonce != dbToken.TokenString {
			fmt.Printf("[WSDebug] Token rotated: req_nonce=%s, db_nonce=%s\n", nonce, dbToken.TokenString)
			return "", "", code.ErrorInvalidUserAuthToken.WithDetails("Token has been rotated")
		}

		// 1. Verify Scope Permissions (Protocol: ws)
		if !pkgapp.VerifyPermissions(dbToken.Scope, "ws", reqClientType, "") {
			fmt.Printf("[WSDebug] Permission denied: scope=%s, protocol=%s, client=%s\n", dbToken.Scope, "ws", reqClientType)
			return "", "", code.ErrorAuthTokenScopeRestricted.WithDetails("Permission denied: Handshake")
		}

		// 2. Verify Client Type (Only for login tokens where ClientType is used for restriction)
		// 仅对登录令牌执行严格客户端匹配，手动令牌通过 Scope 校验
		if dbToken.IssueType == 1 && dbToken.ClientType != "" && !pkgapp.MatchWildcard(dbToken.ClientType, reqClientType) {
			fmt.Printf("[WSDebug] ClientType mismatch: req=%s, db=%s\n", reqClientType, dbToken.ClientType)
			return "", "", code.ErrorAuthTokenClientRestricted.WithDetails("Client mismatch")
		}

		// 3. Verify User-Agent (Only if bound)
		if dbToken.UserAgent != "" && !pkgapp.MatchWildcard(dbToken.UserAgent, reqUserAgent) {
			fmt.Printf("[WSDebug] User-Agent mismatch: req=%s, db=%s\n", reqUserAgent, dbToken.UserAgent)
			return "", "", code.ErrorAuthTokenUARestricted
		}

		// 4. Verify IP (Only if bound)
		if dbToken.BoundIP != "" && !pkgapp.MatchWildcard(dbToken.BoundIP, reqIP) {
			fmt.Printf("[WSDebug] IP mismatch: req=%s, db=%s\n", reqIP, dbToken.BoundIP)
			return "", "", code.ErrorAuthTokenIPRestricted
		}

		_ = appContainer.TokenService.RecordAccessLog(ctx, &domain.AuthTokenLog{
			TokenID:       tokenID,
			UID:           uid,
			Protocol:      "ws",
			Client:        reqClientType,
			ClientName:    reqClientName,
			ClientVersion: reqClientVersion,
			IP:            reqIP,
			UA:            reqUserAgent,
			StatusCode:    101, // Switching Protocols
		})

		return dbToken.Scope, dbToken.Vaults, nil
	})

	// Register online status hooks
	wss.UseUserConnect(func(uid int64) {
		broadcastMemberStatus(context.Background(), appContainer, uid, true)
		broadcastFriendStatus(context.Background(), appContainer, uid, true)
		sendOnlineMembersToUser(context.Background(), appContainer, uid)
		sendOnlineFriendsToUser(context.Background(), appContainer, uid)
	})
	wss.UseUserDisconnect(func(uid int64) {
		broadcastMemberStatus(context.Background(), appContainer, uid, false)
		broadcastFriendStatus(context.Background(), appContainer, uid, false)
	})
}

func broadcastMemberStatus(ctx context.Context, appContainer *app.App, uid int64, online bool) {
	if appContainer.VaultMemberRepo == nil || appContainer.SharedVaultRepo == nil || appContainer.UserRepo == nil || appContainer.GetWSS() == nil {
		return
	}

	// Get user display name
	username := fmt.Sprintf("%d", uid)
	if user, err := appContainer.UserRepo.GetByUID(ctx, uid, true); err == nil && user != nil {
		if user.Username != "" {
			username = user.Username
		} else if user.Email != "" {
			username = user.Email
		}
	}

	// Collect all shared vaults this user is part of (as member or owner)
	vaultSet := make(map[string]struct{})

	// Vaults where user is a member
	memberships, err := appContainer.VaultMemberRepo.ListByMember(ctx, uid)
	if err == nil {
		for _, m := range memberships {
			vaultSet[m.VaultName] = struct{}{}
		}
	}

	// Vaults where user is the owner (accepted shares only)
	ownedShares, err := appContainer.SharedVaultRepo.ListByOwner(ctx, uid)
	if err == nil {
		for _, s := range ownedShares {
			if s.Status == domain.SharedVaultAccepted {
				vaultSet[s.VaultName] = struct{}{}
			}
		}
	}

	if len(vaultSet) == 0 {
		return
	}

	action := websocket_router.MemberOnline
	if !online {
		action = websocket_router.MemberOffline
	}

	for vaultName := range vaultSet {
		payload := dto.MemberStatusMessage{
			UID:      uid,
			Nickname: username,
			Vault:    vaultName,
			Online:   online,
		}
		c := code.Success.WithData(payload).WithVault(vaultName)

		vaultMembers, err := appContainer.VaultMemberRepo.ListByVault(ctx, vaultName)
		if err != nil {
			appContainer.Logger().Warn("broadcastMemberStatus: ListByVault failed",
				zap.String("vault", vaultName), zap.Error(err))
			continue
		}
		for _, vm := range vaultMembers {
			if vm.MemberUID == uid {
				continue
			}
			appContainer.GetWSS().BroadcastToUser(vm.MemberUID, c, action)
		}
	}
}

// broadcastFriendStatus broadcasts the user's online/offline status to all accepted friends.
func broadcastFriendStatus(ctx context.Context, appContainer *app.App, uid int64, online bool) {
	if appContainer.FriendRelationshipRepo == nil || appContainer.UserRepo == nil || appContainer.GetWSS() == nil {
		return
	}

	friends, err := appContainer.FriendRelationshipRepo.ListAcceptedByUID(ctx, uid)
	if err != nil || len(friends) == 0 {
		return
	}

	username := fmt.Sprintf("%d", uid)
	if user, err := appContainer.UserRepo.GetByUID(ctx, uid, true); err == nil && user != nil {
		if user.Username != "" {
			username = user.Username
		} else if user.Email != "" {
			username = user.Email
		}
	}

	action := websocket_router.MemberOnline
	if !online {
		action = websocket_router.MemberOffline
	}

	payload := dto.MemberStatusMessage{
		UID:    uid,
		Vault:  "",
		Online: online,
	}
	// Use Nickname for display name (the frontend shows this)
	payload.Nickname = username
	c := code.Success.WithData(payload)
	// No WithVault — frontend message.vault will be "" so vault filter passes it

	wss := appContainer.GetWSS()
	for _, f := range friends {
		friendUID := f.FriendUID
		if friendUID == uid {
			friendUID = f.UID
		}
		wss.BroadcastToUser(friendUID, c, action)
	}
}

// sendOnlineMembersToUser sends the current online vault members to the newly connected user.
func sendOnlineMembersToUser(ctx context.Context, appContainer *app.App, newUID int64) {
	if appContainer.VaultMemberRepo == nil || appContainer.SharedVaultRepo == nil || appContainer.GetWSS() == nil {
		return
	}

	vaultSet := make(map[string]struct{})

	memberships, err := appContainer.VaultMemberRepo.ListByMember(ctx, newUID)
	if err == nil {
		for _, m := range memberships {
			vaultSet[m.VaultName] = struct{}{}
		}
	}

	ownedShares, err := appContainer.SharedVaultRepo.ListByOwner(ctx, newUID)
	if err == nil {
		for _, s := range ownedShares {
			if s.Status == domain.SharedVaultAccepted {
				vaultSet[s.VaultName] = struct{}{}
			}
		}
	}

	if len(vaultSet) == 0 {
		return
	}

	wss := appContainer.GetWSS()
	for vaultName := range vaultSet {
		vaultMembers, err := appContainer.VaultMemberRepo.ListByVault(ctx, vaultName)
		if err != nil {
			continue
		}
		for _, vm := range vaultMembers {
			if vm.MemberUID == newUID || !wss.IsUserOnline(vm.MemberUID) {
				continue
			}

			username := fmt.Sprintf("%d", vm.MemberUID)
			if user, err := appContainer.UserRepo.GetByUID(ctx, vm.MemberUID, true); err == nil && user != nil {
				if user.Username != "" {
					username = user.Username
				} else if user.Email != "" {
					username = user.Email
				}
			}
			payload := dto.MemberStatusMessage{
				UID:      vm.MemberUID,
				Nickname: username,
				Vault:    vaultName,
				Online:   true,
			}
			c := code.Success.WithData(payload).WithVault(vaultName)
			wss.BroadcastToUser(newUID, c, websocket_router.MemberOnline)
		}
	}
}

// sendOnlineFriendsToUser sends the online status of all currently online friends to the newly connected user.
func sendOnlineFriendsToUser(ctx context.Context, appContainer *app.App, newUID int64) {
	if appContainer.FriendRelationshipRepo == nil || appContainer.GetWSS() == nil {
		return
	}

	friends, err := appContainer.FriendRelationshipRepo.ListAcceptedByUID(ctx, newUID)
	if err != nil || len(friends) == 0 {
		return
	}

	wss := appContainer.GetWSS()
	for _, f := range friends {
		friendUID := f.FriendUID
		if friendUID == newUID {
			friendUID = f.UID
		}
		if !wss.IsUserOnline(friendUID) {
			continue
		}

		username := fmt.Sprintf("%d", friendUID)
		if user, err := appContainer.UserRepo.GetByUID(ctx, friendUID, true); err == nil && user != nil {
			if user.Username != "" {
				username = user.Username
			} else if user.Email != "" {
				username = user.Email
			}
		}
		payload := dto.MemberStatusMessage{
			UID:      friendUID,
			Nickname: username,
			Vault:    "",
			Online:   true,
		}
		c := code.Success.WithData(payload)
		wss.BroadcastToUser(newUID, c, websocket_router.MemberOnline)
	}
}
