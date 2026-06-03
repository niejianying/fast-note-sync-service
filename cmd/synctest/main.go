package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lxzan/gws"
)

var (
	serverURL = flag.String("url", "https://fns.niejianying.cn", "Backend URL")
	aPhone    = flag.String("a-phone", "18206398190", "User A phone")
	aPwd      = flag.String("a-pwd", "", "User A password (auto-derived if empty)")
	bPhone    = flag.String("b-phone", "15601691325", "User B phone")
	bPwd      = flag.String("b-pwd", "123456", "User B password")
	vault     = flag.String("vault", "synctest-vault", "Vault name for testing")
)

type httpClient struct{ base string }

func (c *httpClient) do(method, path, body string, token string) (int, json.RawMessage) {
	req, _ := http.NewRequest(method, c.base+path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client", "WebGui")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("   [http] %s %s error: %v\n", method, path, err)
		return 0, nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	json.Unmarshal(b, &r)
	if r.Code != 1 && r.Code != 2 && r.Code != 3 && r.Code != 4 {
		fmt.Printf("   [http] %s %s -> code=%d msg=%s\n", method, path, r.Code, r.Message)
	}
	return r.Code, r.Data
}

type wsClient struct {
	conn *gws.Conn
	ch   chan string
}

func newWSClient(addr, token, name string) (*wsClient, error) {
	ch := make(chan string, 100)
	c := &wsClient{ch: ch}
	h := &wsHandler{name: name, ch: ch}
	conn, httpResp, err := gws.NewClient(h, &gws.ClientOption{Addr: addr, HandshakeTimeout: 10 * time.Second})
	if err != nil {
		if httpResp != nil {
			b, _ := io.ReadAll(httpResp.Body)
			httpResp.Body.Close()
			return nil, fmt.Errorf("dial: %v (http %d: %s)", err, httpResp.StatusCode, string(b))
		}
		return nil, fmt.Errorf("dial: %w", err)
	}
	c.conn = conn
	go conn.ReadLoop()
	// Authorize
	conn.WriteMessage(gws.OpcodeText, []byte("Authorization|"+token))
	// Wait for auth response
	select {
	case msg := <-ch:
		if strings.Contains(msg, `"code":1`) {
			fmt.Printf("   %s auth OK\n", name)
		} else {
			fmt.Printf("   %s auth failed: %.80s\n", name, msg)
		}
	case <-time.After(5 * time.Second):
		fmt.Printf("   %s auth timeout\n", name)
	}
	return c, nil
}

func (c *wsClient) Send(action, payload string) {
	c.conn.WriteMessage(gws.OpcodeText, []byte(action+"|"+payload))
}

func (c *wsClient) Close() {
	c.conn.WriteClose(1000, []byte("bye"))
}

type wsHandler struct {
	name string
	ch   chan string
}

func (h *wsHandler) OnOpen(*gws.Conn)                {}
func (h *wsHandler) OnClose(*gws.Conn, error)        {}
func (h *wsHandler) OnPing(*gws.Conn, []byte)        {}
func (h *wsHandler) OnPong(*gws.Conn, []byte)        {}
func (h *wsHandler) OnMessage(_ *gws.Conn, msg *gws.Message) {
	s := msg.Data.String()
	if strings.HasPrefix(s, "ClientInfo") {
		return
	}
	h.ch <- s
}

func main() {
	flag.Parse()
	if *aPwd == "" {
		h := sha256.Sum256([]byte("sync:" + *aPhone + ":niejianying"))
		*aPwd = hex.EncodeToString(h[:])[:32]
	}

	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║    Cross-User Sync Test Tool         ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Printf("Server: %s\n", *serverURL)
	fmt.Printf("Vault:  %s\n\n", *vault)

	api := &httpClient{base: *serverURL}
	wsURL := strings.Replace(*serverURL, "https://", "wss://", 1) +
		"/api/user/sync?client=FlutterApp&clientName=SyncTest&clientVersion=1.0"

	// 1. Login
	fmt.Println("1. Login")
	_, d := api.do("POST", "/api/user/login",
		fmt.Sprintf(`{"credentials":"%s","password":"%s"}`, *aPhone, *aPwd), "")
	var lr struct {
		UID   int64  `json:"uid"`
		Token string `json:"token"`
	}
	json.Unmarshal(d, &lr)
	tokenA, uidA := lr.Token, lr.UID
	_, d = api.do("POST", "/api/user/login",
		fmt.Sprintf(`{"credentials":"%s","password":"%s"}`, *bPhone, *bPwd), "")
	json.Unmarshal(d, &lr)
	tokenB, uidB := lr.Token, lr.UID
	fmt.Printf("   A: uid=%d\n   B: uid=%d\n", uidA, uidB)

	// 2. Friendship
	fmt.Println("\n2. Friendship")
	code, d := api.do("GET", "/api/friends", "", tokenA)
	var friends []struct {
		FriendUID int64  `json:"friendUid"`
		Status    string `json:"status"`
	}
	if code == 1 {
		json.Unmarshal(d, &friends)
	}
	var friendUID int64
	for _, f := range friends {
		if f.Status == "accepted" {
			friendUID = f.FriendUID
			break
		}
	}
	if friendUID == 0 {
		fmt.Print("   Creating friendship...")
		api.do("POST", "/api/friend/add", fmt.Sprintf(`{"friendUid":%d}`, uidA), tokenB)
		api.do("POST", "/api/friend/respond", fmt.Sprintf(`{"friendUid":%d,"accept":true}`, uidB), tokenA)
		_, d = api.do("GET", "/api/friends", "", tokenA)
		json.Unmarshal(d, &friends)
		if len(friends) > 0 {
			friendUID = friends[0].FriendUID
		}
		fmt.Println(" done")
	}
	fmt.Printf("   Friend: uid=%d\n", friendUID)

	// 3. Create vault
	fmt.Println("\n3. Create vault")
	api.do("POST", "/api/vault", fmt.Sprintf(`{"vault":"%s","id":0}`, *vault), tokenA)
	fmt.Printf("   Vault: %s\n", *vault)

	// 4. Share vault
	fmt.Println("4. Share vault")
	// Check existing shares first
	code, d = api.do("GET", "/api/vault/shares/outgoing", "", tokenA)
	var existingShares []struct {
		ID        int64  `json:"id"`
		VaultName string `json:"vaultName"`
		Status    string `json:"status"`
		TargetUID int64  `json:"targetUid"`
	}
	var shareID int64
	if code == 1 {
		json.Unmarshal(d, &existingShares)
		for _, s := range existingShares {
			if s.VaultName == *vault && s.TargetUID == friendUID && (s.Status == "accepted" || s.Status == "pending") {
				shareID = s.ID
				fmt.Printf("   Already shared: id=%d status=%s\n", shareID, s.Status)
				break
			}
		}
	}
	if shareID == 0 {
		code, d = api.do("POST", "/api/vault/share",
			fmt.Sprintf(`{"friendUid":%d,"vaultName":"%s","vaultKey":"test-key"}`, friendUID, *vault), tokenA)
		var shareResp struct{ ID int64 }
		if code == 1 {
			json.Unmarshal(d, &shareResp)
			shareID = shareResp.ID
		}
		fmt.Printf("   Share ID: %d\n", shareID)
	}

	// 5. Accept share
	fmt.Println("5. Accept share")
	api.do("POST", fmt.Sprintf("/api/vault/share/%d/respond", shareID), `{"accept":true}`, tokenB)
	fmt.Println("   Accepted")

	// 6. Create WS tokens
	fmt.Println("\n6. Create WS tokens")
	code, d = api.do("POST", "/api/token",
		`{"clientType":"FlutterApp","protocol":"ws","client":"FlutterApp","function":"*","expiredDays":365}`, tokenA)
	var tr struct{ Token string }
	if code == 1 && len(d) > 0 {
		json.Unmarshal(d, &tr)
	}
	wsTokenA := tr.Token
	code, d = api.do("POST", "/api/token",
		`{"clientType":"FlutterApp","protocol":"ws","client":"FlutterApp","function":"*","expiredDays":365}`, tokenB)
	if code == 1 && len(d) > 0 {
		json.Unmarshal(d, &tr)
	}
	wsTokenB := tr.Token

	// 7. Connect WebSocket
	fmt.Println("\n7. Connect WebSocket")
	clientA, err := newWSClient(wsURL, wsTokenA, "A")
	if err != nil {
		fmt.Printf("   ❌ A: %v\n", err)
		return
	}
	fmt.Printf("   A connected\n")
	time.Sleep(500 * time.Millisecond)
	clientB, err := newWSClient(wsURL, wsTokenB, "B")
	if err != nil {
		fmt.Printf("   ❌ B: %v\n", err)
		return
	}
	fmt.Printf("   B connected\n")
	time.Sleep(1 * time.Second)

	time.Sleep(500 * time.Millisecond)

	// 8. Note broadcast test
	fmt.Println("\n8. A sends NoteModify")
	mtime := time.Now().UnixMilli()
	ph := md5.Sum([]byte("/synctest.md"))
	noteReq := map[string]interface{}{
		"vault": *vault, "path": "/synctest.md",
		"pathHash": hex.EncodeToString(ph[:]), "baseHash": "init",
		"content": "# SyncTest\n\nCross-user broadcast works!",
		"contentHash": fmt.Sprintf("st_%d", mtime),
		"size": 55, "mtime": mtime, "ctime": mtime,
	}
	p, _ := json.Marshal(noteReq)
	clientA.Send("NoteModify", string(p))
	fmt.Println("   Sent, waiting for broadcast...")

	select {
	case msg := <-clientB.ch:
		fmt.Printf("\n   ✅ B received (%d bytes): %.150s\n", len(msg), msg)
		if strings.Contains(msg, "NoteSyncModify") {
			fmt.Println("\n   ✅ NOTE BROADCAST: WORKING")
		}
	case <-time.After(15 * time.Second):
		fmt.Println("\n   ❌ B did not receive (timeout)")
	}

	// 9. Folder broadcast test
	fmt.Println("\n9. A sends FolderModify")
	fph := md5.Sum([]byte("/synctest-folder"))
	folderReq := map[string]interface{}{
		"vault": *vault, "path": "/synctest-folder",
		"pathHash": hex.EncodeToString(fph[:]),
	}
	p, _ = json.Marshal(folderReq)
	clientA.Send("FolderModify", string(p))
	fmt.Println("   Sent, waiting for broadcast...")

	select {
	case msg := <-clientB.ch:
		fmt.Printf("\n   ✅ B received (%d bytes): %.150s\n", len(msg), msg)
		if strings.Contains(msg, "FolderSyncModify") {
			fmt.Println("\n   ✅ FOLDER BROADCAST: WORKING")
		}
	case <-time.After(15 * time.Second):
		fmt.Println("\n   ❌ B did not receive folder broadcast (timeout)")
	}

	// 10. Online status test
	fmt.Println("\n10. Online status test")
	// Connect B after A to trigger a MemberOnline broadcast from B's connection
	// When B reconnects, the onUserConnect hook fires and broadcasts MemberOnline
	// A should receive MemberOnline for B's vault membership

	// Disconnect B first to clear state, then reconnect
	clientB.Close()
	time.Sleep(1 * time.Second)

	// Wait for A to have received the disconnect message
	select {
	case msg := <-clientA.ch:
		if strings.Contains(msg, "MemberOffline") {
			fmt.Printf("   ✅ A received MemberOffline: %.120s\n", msg)
		}
	case <-time.After(3 * time.Second):
		fmt.Println("   ⚠️  A did not receive MemberOffline (may use different vault)")
	}

	// Reconnect B
	clientB2, err := newWSClient(wsURL, wsTokenB, "B")
	if err != nil {
		fmt.Printf("   ❌ B reconnect: %v\n", err)
	} else {
		clientB = clientB2
		time.Sleep(1 * time.Second)
		select {
		case msg := <-clientA.ch:
			if strings.Contains(msg, "MemberOnline") {
				fmt.Printf("   ✅ A received MemberOnline: %.120s\n", msg)
			} else if strings.Contains(msg, "NoteSyncModify") || strings.Contains(msg, "FolderSyncModify") {
				fmt.Printf("   ⚠️  A received other (MemberOnline may have been earlier): %.80s\n", msg)
			}
		case <-time.After(3 * time.Second):
			fmt.Println("   ⚠️  A did not receive MemberOnline")
		}
	}

	// 11. File broadcast (skipped - requires multi-step binary upload)
	fmt.Println("\n11. File broadcast test (skipped)")
	fmt.Println("   Note: File broadcast requires FileUploadChunkBinary + UploadComplete")
	fmt.Println("   Manual test: use Flutter app to upload a file in shared vault")

	clientA.Close()
	if clientB != nil {
		clientB.Close()
	}

	fmt.Println("\n╔══════════════════════════════════════╗")
	fmt.Println("║          Test Complete               ║")
	fmt.Println("╚══════════════════════════════════════╝")
}
