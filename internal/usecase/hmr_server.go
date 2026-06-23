package usecase

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

type HMRMessage struct {
	Type string `json:"type"`
}

type HMRServer struct {
	clients   map[*websocket.Conn]bool
	mu        sync.RWMutex
	upgrader  websocket.Upgrader
	assetsDir string
}

func NewHMRServer(assetsDir string) *HMRServer {
	return &HMRServer{
		clients:   make(map[*websocket.Conn]bool),
		assetsDir: assetsDir,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *HMRServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", "error", err)
		return
	}

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	logger.Debug("HMR client connected", "total", len(s.clients))

	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
		logger.Debug("HMR client disconnected", "total", len(s.clients))
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var m HMRMessage
		if err := json.Unmarshal(msg, &m); err == nil {
			logger.Debug("HMR client message", "type", m.Type)
		}
	}
}

func (s *HMRServer) SendReload() {
	s.broadcast(HMRMessage{Type: "reload"})
}

func (s *HMRServer) SendCSSUpdate(path string) {
	rel, err := filepath.Rel(s.assetsDir, path)
	if err != nil {
		rel = filepath.Base(path)
	}
	s.broadcast(map[string]any{
		"type":     "css-update",
		"filepath": rel,
	})
}

func (s *HMRServer) SendJSUpdate(path string) {
	rel, err := filepath.Rel(s.assetsDir, path)
	if err != nil {
		rel = filepath.Base(path)
	}
	s.broadcast(map[string]any{
		"type":     "js-update",
		"filepath": rel,
	})
}

func (s *HMRServer) SendHTMLUpdate(path string) {
	s.broadcast(HMRMessage{Type: "reload"})
}

func (s *HMRServer) broadcast(msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			logger.Debug("HMR send failed", "error", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

func (s *HMRServer) ClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

func (s *HMRServer) OnFileChanged(path string) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".css":
		s.SendCSSUpdate(path)
	case ".js":
		s.SendJSUpdate(path)
	case ".html", ".htm":
		s.SendHTMLUpdate(path)
	default:
		s.SendReload()
	}
}
