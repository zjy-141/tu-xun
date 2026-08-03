package controller

import (
	"sync"
	"time"
)

type sessionEntry struct {
	UserSession UserSession
	CreatedAt   time.Time
}

var (
	sessionStoreMu sync.RWMutex
	sessionStore   = make(map[string]sessionEntry)
)

// StoreXSession 在服务端存储 session_id 与 UserSession 的映射
func StoreXSession(sid string, us UserSession) {
	sessionStoreMu.Lock()
	sessionStore[sid] = sessionEntry{UserSession: us, CreatedAt: time.Now()}
	sessionStoreMu.Unlock()
}

// GetXSession 根据 session_id 查找 UserSession
func GetXSession(sid string) (UserSession, bool) {
	sessionStoreMu.RLock()
	entry, ok := sessionStore[sid]
	sessionStoreMu.RUnlock()
	if !ok {
		return UserSession{}, false
	}
	return entry.UserSession, true
}

// RemoveXSession 删除 session_id 映射
func RemoveXSession(sid string) {
	sessionStoreMu.Lock()
	delete(sessionStore, sid)
	sessionStoreMu.Unlock()
}

// cleanExpiredXSessions 后台清理过期（超过14天）的 session 映射
func cleanExpiredXSessions() {
	for {
		time.Sleep(1 * time.Hour)
		now := time.Now()
		sessionStoreMu.Lock()
		for sid, entry := range sessionStore {
			if now.Sub(entry.CreatedAt) > 14*24*time.Hour {
				delete(sessionStore, sid)
			}
		}
		sessionStoreMu.Unlock()
	}
}
