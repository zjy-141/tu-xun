package controller

import (
	"sync"
	"time"
)

type sessionEntry struct {
	UserID    int64
	CreatedAt time.Time
}

var (
	sessionStoreMu sync.RWMutex
	sessionStore   = make(map[string]sessionEntry)
)

// StoreXSession 在服务端存储 session_id 与用户 id 的映射
func StoreXSession(sid string, userID int64) {
	sessionStoreMu.Lock()
	sessionStore[sid] = sessionEntry{UserID: userID, CreatedAt: time.Now()}
	sessionStoreMu.Unlock()
}

// GetXSession 根据 session_id 查找用户 id
func GetXSession(sid string) (int64, bool) {
	sessionStoreMu.RLock()
	entry, ok := sessionStore[sid]
	sessionStoreMu.RUnlock()
	if !ok {
		return 0, false
	}
	return entry.UserID, true
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
