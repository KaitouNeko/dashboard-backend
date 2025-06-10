package chat

import (
	"net/http"

	"ai-workshop/internal/config"
	"ai-workshop/internal/llm"

	"github.com/gin-gonic/gin"
)

type RAGRequest struct {
	Message string `json:"message"`
}

type RAGResponse struct {
	Response string `json:"response"`
}

type Handler struct {
	client  llm.LLMProvider
	service *Service
}

func NewHandler(service *Service, llmType llm.LLMType, config *config.Config) *Handler {

	factory := llm.NewFactory(config)

	client, _ := factory.Create(llmType)

	return &Handler{
		client:  client,
		service: service,
	}
}

// HandleChat handles chat requests
func (h *Handler) ChatHandler(c *gin.Context) {
	var req struct {
		Message string      `json:"message" binding:"required"`
		Model   llm.LLMType `json:"model,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// 如果没有指定模型，默認使用 Gemini
	if req.Model == "" {
		req.Model = llm.LLMTypeGemini
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

func (h *Handler) RagChatHandler(c *gin.Context) {
	var req struct {
		Message        string            `json:"message" binding:"required"`
		Model          llm.LLMType       `json:"model,omitempty"`
		EmbeddingModel llm.EmbeddingType `json:"embedding_model,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// TODO need to use new flow

	// 如果未指定嵌入模型，使用默認的 OpenAI
	embeddingModel := req.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = llm.EmbeddingTypeOpenAI
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
