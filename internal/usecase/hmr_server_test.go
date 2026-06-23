package usecase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func dialTestWS(t *testing.T, serverURL string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial test WS: %v", err)
	}
	return conn
}

func TestHMRServer_NewServer(t *testing.T) {
	srv := NewHMRServer("/tmp/assets")
	if srv == nil {
		t.Fatal("NewHMRServer returned nil")
	}
	if srv.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", srv.ClientCount())
	}
	if srv.assetsDir != "/tmp/assets" {
		t.Errorf("assetsDir = %q, want /tmp/assets", srv.assetsDir)
	}
}

func TestHMRServer_ClientCount(t *testing.T) {
	srv := NewHMRServer("/tmp/assets")
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.HandleWebSocket)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	if srv.ClientCount() != 0 {
		t.Fatalf("initial client count = %d, want 0", srv.ClientCount())
	}

	conn1 := dialTestWS(t, ts.URL)
	defer conn1.Close()
	time.Sleep(10 * time.Millisecond)
	if srv.ClientCount() != 1 {
		t.Errorf("after 1st connect: count = %d, want 1", srv.ClientCount())
	}

	conn2 := dialTestWS(t, ts.URL)
	defer conn2.Close()
	time.Sleep(10 * time.Millisecond)
	if srv.ClientCount() != 2 {
		t.Errorf("after 2nd connect: count = %d, want 2", srv.ClientCount())
	}

	conn1.Close()
	time.Sleep(10 * time.Millisecond)
	if srv.ClientCount() != 1 {
		t.Errorf("after 1st disconnect: count = %d, want 1", srv.ClientCount())
	}

	conn2.Close()
	time.Sleep(10 * time.Millisecond)
	if srv.ClientCount() != 0 {
		t.Errorf("after all disconnects: count = %d, want 0", srv.ClientCount())
	}
}

func readMessage(t *testing.T, conn *websocket.Conn) map[string]any {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read WS message: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(msg, &m); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}
	return m
}

func TestHMRServer_OnFileChanged_CSS(t *testing.T) {
	srv := NewHMRServer(t.TempDir())
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.HandleWebSocket)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	conn := dialTestWS(t, ts.URL)
	defer conn.Close()
	time.Sleep(10 * time.Millisecond)

	srv.OnFileChanged("/tmp/assets/css/styles.css")

	msg := readMessage(t, conn)
	if msg["type"] != "css-update" {
		t.Errorf("type = %v, want css-update", msg["type"])
	}
	fp, ok := msg["filepath"].(string)
	if !ok || fp == "" {
		t.Errorf("filepath missing or empty in css-update message")
	}
}

func TestHMRServer_OnFileChanged_JS(t *testing.T) {
	srv := NewHMRServer(t.TempDir())
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.HandleWebSocket)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	conn := dialTestWS(t, ts.URL)
	defer conn.Close()
	time.Sleep(10 * time.Millisecond)

	srv.OnFileChanged("/tmp/assets/js/app.js")

	msg := readMessage(t, conn)
	if msg["type"] != "js-update" {
		t.Errorf("type = %v, want js-update", msg["type"])
	}
	fp, ok := msg["filepath"].(string)
	if !ok || fp == "" {
		t.Errorf("filepath missing or empty in js-update message")
	}
}

func TestHMRServer_OnFileChanged_HTML(t *testing.T) {
	srv := NewHMRServer(t.TempDir())
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.HandleWebSocket)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	conn := dialTestWS(t, ts.URL)
	defer conn.Close()
	time.Sleep(10 * time.Millisecond)

	srv.OnFileChanged("/tmp/assets/index.html")

	msg := readMessage(t, conn)
	if msg["type"] != "reload" {
		t.Errorf("type = %v, want reload", msg["type"])
	}
}

func TestHMRServer_OnFileChanged_Other(t *testing.T) {
	srv := NewHMRServer(t.TempDir())
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.HandleWebSocket)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	conn := dialTestWS(t, ts.URL)
	defer conn.Close()
	time.Sleep(10 * time.Millisecond)

	srv.OnFileChanged("/tmp/assets/image.png")

	msg := readMessage(t, conn)
	if msg["type"] != "reload" {
		t.Errorf("type = %v, want reload", msg["type"])
	}
}

func TestHMRServer_CSSUpdate_Filepath(t *testing.T) {
	tmpDir := t.TempDir()
	srv := NewHMRServer(tmpDir)
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.HandleWebSocket)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	conn := dialTestWS(t, ts.URL)
	defer conn.Close()
	time.Sleep(10 * time.Millisecond)

	srv.SendCSSUpdate(filepath.Join(tmpDir, "css", "styles.css"))

	msg := readMessage(t, conn)
	if msg["type"] != "css-update" {
		t.Errorf("type = %v, want css-update", msg["type"])
	}
	fp := msg["filepath"].(string)
	if fp != filepath.Join("css", "styles.css") {
		t.Errorf("filepath = %q, want css/styles.css", fp)
	}
}

func TestHMRServer_NoClients(t *testing.T) {
	srv := NewHMRServer(t.TempDir())
	srv.SendReload()
	srv.SendCSSUpdate("/tmp/test.css")
	srv.SendJSUpdate("/tmp/test.js")
	srv.SendHTMLUpdate("/tmp/index.html")
}
