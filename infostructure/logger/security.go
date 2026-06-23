package logger

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientInfo struct {
	count     int
	lastReset time.Time
}

type Guard struct {
	mu        sync.Mutex
	clients   map[string]*clientInfo
	blacklist map[string]bool
	maxReqs   int
}

func NewGuard(maxReqsPerSecond int) *Guard {
	return &Guard{
		clients:   make(map[string]*clientInfo),
		blacklist: make(map[string]bool),
		maxReqs:   maxReqsPerSecond,
	}
}

func (g *Guard) LimitAndCheck(s *Server, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r.RemoteAddr)

		g.mu.Lock()
		
		if g.blacklist[ip] {
			g.mu.Unlock()
			http.Error(w, "ip is blacklisted", http.StatusForbidden)
			return
		}

		now := time.Now()
		client, exists := g.clients[ip]

		if !exists {
			client = &clientInfo{lastReset: now}
			g.clients[ip] = client
		}

		if now.Sub(client.lastReset) > time.Second {
			client.count = 0
			client.lastReset = now
		}

		client.count++
		g.mu.Unlock()

		if client.count > g.maxReqs {
			_ = s.append("" + ip + "limit")
			http.Error(w, "too many requests.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (g *Guard) BlockIP(ip string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.blacklist[ip] = true
}

func (g *Guard) UnblockIP(ip string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.blacklist, ip)
}

func extractIP(remoteAddr string) string {
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}
	return remoteAddr
}
