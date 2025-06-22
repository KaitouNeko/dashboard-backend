package models

import (
	llmtype "ai-workshop/internal/constants"
	"time"
)

// Message represents a single message in a conversation
type Message struct {
	Role      string    `json:"role"`      // "user" or "assistant"
	Content   string    `json:"content"`   // message content
	Timestamp time.Time `json:"timestamp"` // when the message was created
}

// ConversationHistory represents the complete conversation history
type ConversationHistory []Message

// ChatRequest represents the incoming chat request with session support
type ChatRequest struct {
	Message             string              `json:"message" binding:"required"`
	Model               llmtype.LLMType     `json:"model,omitempty"`
	SessionID           string              `json:"sessionId,omitempty"`
	ConversationHistory ConversationHistory `json:"conversationHistory,omitempty"`
}

// RAGRequest represents the incoming RAG request with session support
type RAGRequest struct {
	Message             string                `json:"message" binding:"required"`
	Model               llmtype.LLMType       `json:"model,omitempty"`
	EmbeddingModel      llmtype.EmbeddingType `json:"embedding_model,omitempty"`
	SessionID           string                `json:"sessionId,omitempty"`
	ConversationHistory ConversationHistory   `json:"conversationHistory,omitempty"`
}

// ChatResponse represents the response sent back to the client
type ChatResponse struct {
	Response  string          `json:"response"`
	Model     llmtype.LLMType `json:"model"`
	SessionID string          `json:"sessionId,omitempty"`
}

// RAGResponse represents the RAG response sent back to the client
type RAGResponse struct {
	Response       string                `json:"response"`
	Model          llmtype.LLMType       `json:"model"`
	EmbeddingModel llmtype.EmbeddingType `json:"embedding_model"`
	SessionID      string                `json:"sessionId,omitempty"`
}

// SessionInfo contains basic session information for logging/monitoring
type SessionInfo struct {
	SessionID    string    `json:"sessionId"`
	MessageCount int       `json:"messageCount"`
	LastActivity time.Time `json:"lastActivity"`
	CreatedAt    time.Time `json:"createdAt"`
}
