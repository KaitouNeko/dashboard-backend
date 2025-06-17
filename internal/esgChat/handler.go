package esgchat

import (
	"fmt"
	"net/http"

	"ai-workshop/internal/config"
	llmtype "ai-workshop/internal/constants"
	"ai-workshop/internal/esgllm"

	"github.com/gin-gonic/gin"
)

type RAGRequest struct {
	Message string `json:"message"`
}

type RAGResponse struct {
	Response string `json:"response"`
}

type Handler struct {
	openAiClient  esgllm.LLMProvider
	geminiClient  esgllm.LLMProvider
	watsonxClient esgllm.LLMProvider
	service       *Service
}

func NewHandler(service *Service, config *config.Config) *Handler {

	factory := esgllm.NewFactory(config)

	// client, _ := factory.Create(llmType)
	openAiClient := factory.CreateOpenAi()
	geminiClient, _ := factory.CreateGemini()
	watsonxClient, _ := factory.CreateWatsonx()

	return &Handler{
		openAiClient:  openAiClient,
		geminiClient:  geminiClient,
		watsonxClient: watsonxClient,
		service:       service,
	}
}

// HandleChat handles chat requests
func (h *Handler) ChatHandler(c *gin.Context) {
	fmt.Println("ChatHandler")
	var req struct {
		Message string          `json:"message" binding:"required"`
		Model   llmtype.LLMType `json:"model,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}
	fmt.Println("req.Model:", req.Model)
	var response string
	var err error

	if req.Model == "" {
		req.Model = llmtype.LLMTypeGemini // 默認使用 Gemini 模型
	}
	// 如果没有指定模型，默認使用 Gemini
	if req.Model == llmtype.LLMTypeGemini {
		response, err = h.geminiClient.GenerateContent(c.Request.Context(), req.Message)
	} else if req.Model == llmtype.LLMTypeOpenAI {
		response, err = h.openAiClient.GenerateContent(c.Request.Context(), req.Message)
	} else if req.Model == llmtype.LLMTypeWatsonx {
		response, err = h.watsonxClient.GenerateContent(c.Request.Context(), req.Message)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支援的模型類型"})
		return

	}

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
		Message        string                `json:"message" binding:"required"`
		Model          llmtype.LLMType       `json:"model,omitempty"`
		EmbeddingModel llmtype.EmbeddingType `json:"embedding_model,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// TODO need to use new flow

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
