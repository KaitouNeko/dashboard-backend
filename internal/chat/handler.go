package chat

import (
	"log"
	"net/http"
	"strings"

	"ai-workshop/internal/config"
	llmtype "ai-workshop/internal/constants"
	"ai-workshop/internal/llm"
	"ai-workshop/internal/models"

	"github.com/gin-gonic/gin"
)

// 保持向後兼容的舊請求格式
type LegacyRAGRequest struct {
	Message string `json:"message"`
}

type LegacyRAGResponse struct {
	Response string `json:"response"`
}

type Handler struct {
	client         llm.LLMProvider
	service        *Service
	sessionManager *SessionManager
}

func NewHandler(service *Service, llmType llmtype.LLMType, config *config.Config) *Handler {

	factory := llm.NewFactory(config)

	client, _ := factory.Create(llmType)

	return &Handler{
		client:         client,
		service:        service,
		sessionManager: NewSessionManager(),
	}
}

// buildConversationPrompt constructs a prompt that includes conversation history
func (h *Handler) buildConversationPrompt(conversationHistory models.ConversationHistory, currentMessage string) string {
	if len(conversationHistory) == 0 {
		// No history, just return the current message
		return currentMessage
	}

	var sb strings.Builder

	// Add system instruction for context
	sb.WriteString("以下是我們的對話歷史，請基於這些歷史內容來回答最新的問題：\n\n")

	// Add conversation history
	sb.WriteString("=== 對話歷史 ===\n")
	for _, msg := range conversationHistory {
		if msg.Role == "user" {
			sb.WriteString("用戶: ")
		} else {
			sb.WriteString("助手: ")
		}
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}

	// Add current message
	sb.WriteString("\n=== 當前問題 ===\n")
	sb.WriteString("用戶: ")
	sb.WriteString(currentMessage)
	sb.WriteString("\n\n請基於以上對話歷史來回答當前問題：")

	return sb.String()
}

// HandleChat handles chat requests with session support
func (h *Handler) ChatHandler(c *gin.Context) {
	var req models.ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// Log session info for monitoring (optional)
	if req.SessionID != "" {
		log.Printf("Chat request for session: %s, history length: %d", req.SessionID, len(req.ConversationHistory))
		// Update session info
		h.sessionManager.UpdateSession(req.SessionID, len(req.ConversationHistory)+1)
	}

	// 如果没有指定模型，默認使用 Gemini
	if req.Model == "" {
		req.Model = llmtype.LLMTypeGemini
	}

	// Build prompt with conversation history
	prompt := h.buildConversationPrompt(req.ConversationHistory, req.Message)

	// Generate response using the LLM with conversation context
	response, err := h.client.GenerateContent(c.Request.Context(), prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成回應失敗: " + err.Error()})
		return
	}

	// Build response with session support
	chatResponse := models.ChatResponse{
		Response:  response,
		Model:     req.Model,
		SessionID: req.SessionID, // Echo back the session ID
	}

	c.JSON(http.StatusOK, chatResponse)
}

func (h *Handler) RagChatHandler(c *gin.Context) {
	var req models.RAGRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// Log session info for monitoring (optional)
	if req.SessionID != "" {
		log.Printf("RAG request for session: %s, history length: %d", req.SessionID, len(req.ConversationHistory))
		// Update session info
		h.sessionManager.UpdateSession(req.SessionID, len(req.ConversationHistory)+1)
	}

	// 如果未指定嵌入模型，使用默認的 OpenAI
	embeddingModel := req.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = llmtype.EmbeddingTypeOpenAI
	}

	// 如果没有指定LLM模型，默認使用 Gemini
	if req.Model == "" {
		req.Model = llmtype.LLMTypeGemini
	}

	// For RAG, we need to modify the service to accept conversation history
	// For now, let's build the prompt with history and pass it to RAG
	queryWithHistory := h.buildConversationPrompt(req.ConversationHistory, req.Message)

	// 生成 RAG 回應 (using the enhanced query with conversation history)
	response, err := h.service.GenerateRAGResponseWithEmbedding(c.Request.Context(), queryWithHistory, req.Model, embeddingModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 RAG 回應失敗: " + err.Error()})
		return
	}

	// Build response with session support
	ragResponse := models.RAGResponse{
		Response:       response,
		Model:          req.Model,
		EmbeddingModel: embeddingModel,
		SessionID:      req.SessionID, // Echo back the session ID
	}

	c.JSON(http.StatusOK, ragResponse)
}

// Session management endpoints

// GetSessionInfo returns information about a specific session
func (h *Handler) GetSessionInfo(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	session, exists := h.sessionManager.GetSession(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetAllSessions returns all active sessions
func (h *Handler) GetAllSessions(c *gin.Context) {
	sessions := h.sessionManager.GetAllSessions()
	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// DeleteSession deletes a specific session
func (h *Handler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	deleted := h.sessionManager.DeleteSession(sessionID)
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session deleted successfully"})
}

// Legacy handlers for backward compatibility (optional)
func (h *Handler) LegacyChatHandler(c *gin.Context) {
	var req struct {
		Message string          `json:"message" binding:"required"`
		Model   llmtype.LLMType `json:"model,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// 如果没有指定模型，默認使用 Gemini
	if req.Model == "" {
		req.Model = llmtype.LLMTypeGemini
	}

	response, err := h.client.GenerateContent(c.Request.Context(), req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成回應失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
		"model":    req.Model,
	})
}

func (h *Handler) LegacyRagChatHandler(c *gin.Context) {
	var req struct {
		Message        string                `json:"message" binding:"required"`
		Model          llmtype.LLMType       `json:"model,omitempty"`
		EmbeddingModel llmtype.EmbeddingType `json:"embedding_model,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// 如果未指定嵌入模型，使用默認的 OpenAI
	embeddingModel := req.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = llmtype.EmbeddingTypeOpenAI
	}

	// 生成 RAG 回應
	response, err := h.service.GenerateRAGResponseWithEmbedding(c.Request.Context(), req.Message, req.Model, embeddingModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 RAG 回應失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response":        response,
		"model":           req.Model,
		"embedding_model": embeddingModel,
	})
}
