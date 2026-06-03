package app

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/lxzan/gws"
	"go.uber.org/zap"
)

// mockAppContainer implements AppContainer minimally for testing.
type mockAppContainer struct{}

func (m *mockAppContainer) Logger() *zap.Logger                                                  { return zap.NewNop() }
func (m *mockAppContainer) SubmitTask(ctx context.Context, task func(context.Context) error) error  { return nil }
func (m *mockAppContainer) SubmitTaskAsync(ctx context.Context, task func(context.Context) error) error { return nil }
func (m *mockAppContainer) Version() VersionInfo                                                   { return VersionInfo{} }
func (m *mockAppContainer) CheckVersion(pluginVersion string) CheckVersionInfo                      { return CheckVersionInfo{} }
func (m *mockAppContainer) Validator() ValidatorInterface                                          { return nil }
func (m *mockAppContainer) IsReturnSuccess() bool                                                   { return false }
func (m *mockAppContainer) GetAuthTokenKey() string                                                { return "" }
func (m *mockAppContainer) IsProductionMode() bool                                                  { return false }
func (m *mockAppContainer) GetTokenService() any                                                   { return nil }

func newTestUser(uid int64) *UserEntity {
	return &UserEntity{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: fmt.Sprintf("%d", uid),
		},
	}
}

var testConnCounter int

func newTestConn() *gws.Conn {
	testConnCounter++
	return &gws.Conn{}
}

func newTestClient(uid int64) *WebsocketClient {
	return &WebsocketClient{
		conn:      newTestConn(),
		User:      newTestUser(uid),
		StartTime: timex.Now(),
	}
}

// TestWebsocketServer_UserConnectDisconnectHooks verifies the onUserConnect/onUserDisconnect
// hook mechanism: first connection triggers connect, last disconnect triggers disconnect.
func TestWebsocketServer_UserConnectDisconnectHooks(t *testing.T) {
	ws := NewWebsocketServer(WSConfig{}, &mockAppContainer{})

	var connectCalls int32
	var disconnectCalls int32
	var lastConnectUID int64
	var lastDisconnectUID int64

	ws.UseUserConnect(func(uid int64) {
		atomic.AddInt32(&connectCalls, 1)
		lastConnectUID = uid
	})
	ws.UseUserDisconnect(func(uid int64) {
		atomic.AddInt32(&disconnectCalls, 1)
		lastDisconnectUID = uid
	})

	uid := int64(42)

	// Simulate first connection for uid=42
	c1 := newTestClient(uid)
	prev := ws.hasUserClients(uid)
	if prev {
		t.Fatal("expected no existing clients before first add")
	}
	_ = ws.AddUserClient(c1)
	ws.triggerUserConnect(uid, prev)

	if connectCalls != 1 {
		t.Fatalf("expected 1 connect call, got %d", connectCalls)
	}
	if lastConnectUID != uid {
		t.Fatalf("expected connect uid=%d, got %d", uid, lastConnectUID)
	}
	if disconnectCalls != 0 {
		t.Fatalf("expected 0 disconnect calls before any disconnect, got %d", disconnectCalls)
	}

	// Second connection for same uid: should NOT trigger connect again
	c2 := newTestClient(uid)
	prev = ws.hasUserClients(uid)
	if !prev {
		t.Fatal("expected existing clients before second add")
	}
	_ = ws.AddUserClient(c2)
	ws.triggerUserConnect(uid, prev)

	if connectCalls != 1 {
		t.Fatalf("expected still 1 connect call (second conn), got %d", connectCalls)
	}

	// Remove first connection: should NOT trigger disconnect (still have c2)
	ws.RemoveUserClient(c1)
	ws.triggerUserDisconnect(uid, ws.hasUserClients(uid))

	if disconnectCalls != 0 {
		t.Fatalf("expected 0 disconnect calls (still have c2), got %d", disconnectCalls)
	}

	// Remove second (last) connection: SHOULD trigger disconnect
	ws.RemoveUserClient(c2)
	ws.triggerUserDisconnect(uid, ws.hasUserClients(uid))

	if disconnectCalls != 1 {
		t.Fatalf("expected 1 disconnect call, got %d", disconnectCalls)
	}
	if lastDisconnectUID != uid {
		t.Fatalf("expected disconnect uid=%d, got %d", uid, lastDisconnectUID)
	}
	if connectCalls != 1 {
		t.Fatalf("expected 1 connect call total, got %d", connectCalls)
	}
}

// TestWebsocketServer_HookNilSafety verifies hooks are safe when nil.
func TestWebsocketServer_HookNilSafety(t *testing.T) {
	ws := NewWebsocketServer(WSConfig{}, &mockAppContainer{})

	uid := int64(99)

	// Without registering hooks, trigger should be no-ops
	c1 := newTestClient(uid)
	prev := ws.hasUserClients(uid)
	_ = ws.AddUserClient(c1)
	ws.triggerUserConnect(uid, prev)

	ws.RemoveUserClient(c1)
	ws.triggerUserDisconnect(uid, ws.hasUserClients(uid))

	// If we get here without panicking, nil safety works
}

// TestWebsocketServer_MultipleUIDs verifies independent tracking per UID.
func TestWebsocketServer_MultipleUIDs(t *testing.T) {
	ws := NewWebsocketServer(WSConfig{}, &mockAppContainer{})

	var connectUIDs []int64
	var disconnectUIDs []int64

	ws.UseUserConnect(func(uid int64) {
		connectUIDs = append(connectUIDs, uid)
	})
	ws.UseUserDisconnect(func(uid int64) {
		disconnectUIDs = append(disconnectUIDs, uid)
	})

	u1 := int64(1)
	u2 := int64(2)

	c1_1 := newTestClient(u1)
	c2_1 := newTestClient(u2)

	// u1 connects
	ws.triggerUserConnect(u1, ws.hasUserClients(u1))
	_ = ws.AddUserClient(c1_1)

	if len(connectUIDs) != 1 || connectUIDs[0] != u1 {
		t.Fatalf("expected connect for u1, got %v", connectUIDs)
	}

	// u2 connects
	ws.triggerUserConnect(u2, ws.hasUserClients(u2))
	_ = ws.AddUserClient(c2_1)

	if len(connectUIDs) != 2 || connectUIDs[1] != u2 {
		t.Fatalf("expected connect for u2, got %v", connectUIDs)
	}

	// u1 disconnects
	ws.RemoveUserClient(c1_1)
	ws.triggerUserDisconnect(u1, ws.hasUserClients(u1))

	if len(disconnectUIDs) != 1 || disconnectUIDs[0] != u1 {
		t.Fatalf("expected disconnect for u1, got %v", disconnectUIDs)
	}

	// u2 disconnects (last)
	ws.RemoveUserClient(c2_1)
	ws.triggerUserDisconnect(u2, ws.hasUserClients(u2))

	if len(disconnectUIDs) != 2 || disconnectUIDs[1] != u2 {
		t.Fatalf("expected disconnect for u2, got %v", disconnectUIDs)
	}
}

// TestWebsocketServer_UserVerifyNilHook verifies UseUserConnect/UseUserDisconnect don't panic with nil.
func TestWebsocketServer_UserVerifyNilHook(t *testing.T) {
	ws := NewWebsocketServer(WSConfig{}, &mockAppContainer{})

	// These should not panic
	ws.UseUserConnect(nil)
	ws.UseUserDisconnect(nil)

	if ws.onUserConnect != nil || ws.onUserDisconnect != nil {
		t.Fatal("expected nil hooks after registering nil")
	}
}

// TestWebsocketServer_AuthorizationUserConnect simulates the Authorization flow
// to verify that the connect hook fires only for first connection.
func TestWebsocketServer_AuthorizationUserConnect(t *testing.T) {
	ws := NewWebsocketServer(WSConfig{}, &mockAppContainer{})

	var connectCount int32
	var disconnectCount int32

	ws.UseUserConnect(func(uid int64) {
		atomic.AddInt32(&connectCount, 1)
	})
	ws.UseUserDisconnect(func(uid int64) {
		atomic.AddInt32(&disconnectCount, 1)
	})

	uid := int64(100)

	// Simulate Authorization: first conn
	prev := ws.hasUserClients(uid)
	c1 := newTestClient(uid)
	_ = ws.AddUserClient(c1)
	ws.triggerUserConnect(uid, prev)

	if connectCount != 1 {
		t.Fatalf("expected 1 connect after first auth, got %d", connectCount)
	}

	// Simulate Authorization: second conn (same user, different device)
	prev = ws.hasUserClients(uid)
	c2 := newTestClient(uid)
	_ = ws.AddUserClient(c2)
	ws.triggerUserConnect(uid, prev)

	if connectCount != 1 {
		t.Fatalf("expected 1 connect total after second auth, got %d", connectCount)
	}

	// Simulate OnClose: first disconnect
	ws.RemoveUserClient(c1)
	ws.triggerUserDisconnect(uid, ws.hasUserClients(uid))

	if disconnectCount != 0 {
		t.Fatalf("expected 0 disconnect after first close (still connected), got %d", disconnectCount)
	}

	// Simulate OnClose: last disconnect
	ws.RemoveUserClient(c2)
	ws.triggerUserDisconnect(uid, ws.hasUserClients(uid))

	if disconnectCount != 1 {
		t.Fatalf("expected 1 disconnect after last close, got %d", disconnectCount)
	}

	// New connection after full disconnect
	prev = ws.hasUserClients(uid)
	c3 := newTestClient(uid)
	_ = ws.AddUserClient(c3)
	ws.triggerUserConnect(uid, prev)

	if connectCount != 2 {
		t.Fatalf("expected 2 connect total after reconnect, got %d", connectCount)
	}
}
