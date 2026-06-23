package logger

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Server struct {
	addr    string
	logPath string
	mu      sync.Mutex
	guard   *Guard
}

func New(addr, logPath string) *Server {
	return &Server{
		addr:    addr,
		logPath: logPath,
		guard:   NewGuard(5),
	}
}

func (s *Server) Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/log", s.handleLog)
	mux.HandleFunc("/logs", s.handleLogs)

	mux.HandleFunc("/block", s.handleBlock)

	protectedHandler := s.guard.LimitAndCheck(s, mux)

	return http.ListenAndServe(s.addr, protectedHandler)
}

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, "empty log", http.StatusBadRequest)
		return
	}

	if err := s.append(string(body)); err != nil {
		http.Error(w, "write error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(filepath.Clean(s.logPath))
	if err != nil {
		http.Error(w, "cannot read log file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(data)
}

func (s *Server) handleBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return
	}

	ipToBlock := r.URL.Query().Get("ip")
	if ipToBlock == "" {
		http.Error(w, "missing ip parameter", http.StatusBadRequest)
		return
	}

	s.guard.BlockIP(ipToBlock)
	_ = s.append("[SECURITY] Manually blacklisted IP: " + ipToBlock)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("IP " + ipToBlock + " has been blacklisted"))
}

func (s *Server) append(line string) error {
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(filepath.Clean(s.logPath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(line)
	return err
}
