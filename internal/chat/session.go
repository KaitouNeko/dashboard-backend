package chat

import (
	"ai-workshop/internal/models"
	"sync"
	"time"
)

// SessionManager manages chat sessions in memory
// This is a simple in-memory implementation for demonstration
// In production, you might want to use Redis or a database
type SessionManager struct {
	sessions map[string]*models.SessionInfo
	mutex    sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*models.SessionInfo),
		mutex:    sync.RWMutex{},
	}
}

// UpdateSession updates or creates session information
func (sm *SessionManager) UpdateSession(sessionID string, messageCount int) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()

	if session, exists := sm.sessions[sessionID]; exists {
		// Update existing session
		session.MessageCount = messageCount
		session.LastActivity = now
	} else {
		// Create new session
		sm.sessions[sessionID] = &models.SessionInfo{
			SessionID:    sessionID,
			MessageCount: messageCount,
			LastActivity: now,
			CreatedAt:    now,
		}
	}
}

// GetSession retrieves session information
func (sm *SessionManager) GetSession(sessionID string) (*models.SessionInfo, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if exists {
		// Return a copy to avoid race conditions
		sessionCopy := *session
		return &sessionCopy, true
	}
	return nil, false
}

// GetAllSessions returns all active sessions
func (sm *SessionManager) GetAllSessions() []models.SessionInfo {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sessions := make([]models.SessionInfo, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, *session)
	}
	return sessions
}

// CleanupOldSessions removes sessions that haven't been active for a certain duration
func (sm *SessionManager) CleanupOldSessions(maxAge time.Duration) int {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	cleaned := 0

	for sessionID, session := range sm.sessions {
		if session.LastActivity.Before(cutoff) {
			delete(sm.sessions, sessionID)
			cleaned++
		}
	}

	return cleaned
}

// DeleteSession removes a specific session
func (sm *SessionManager) DeleteSession(sessionID string) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		delete(sm.sessions, sessionID)
		return true
	}
	return false
}

// GetSessionCount returns the total number of active sessions
func (sm *SessionManager) GetSessionCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return len(sm.sessions)
}
