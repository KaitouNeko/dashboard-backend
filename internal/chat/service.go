package chat

import (
	"context"
	"fmt"
	"strings"

	"ai-workshop/internal/config"
	"ai-workshop/internal/documents"
	"ai-workshop/internal/llm"
)

// Service 是 RAG 服務的實現
type Service struct {
	docService       *documents.Service
	embeddingService *llm.OpenAIProvider
	llmFactory       *llm.Factory
}

// NewService 創建一個新的 RAG 服務
func NewService(config *config.Config) (*Service, error) {
	// 創建文檔服務
	docService := documents.NewService(config)

	// 創建嵌入服務
	embeddingService := llm.NewOpenAiProvider(config.OpenAiAPIKey)

	// 創建 LLM 工廠
	llmFactory := llm.NewFactory(config)

	return &Service{
		docService:       docService,
		embeddingService: embeddingService,
		llmFactory:       llmFactory,
	}, nil
}

// GenerateRAGResponse 生成 RAG 回應
func (s *Service) GenerateRAGResponse(ctx context.Context, query string, modelType llm.LLMType) (string, error) {
	// 使用默認的嵌入提供者
	return s.GenerateRAGResponseWithEmbedding(ctx, query, modelType, llm.EmbeddingTypeOpenAI)
}

// GenerateRAGResponseWithEmbedding 使用指定的嵌入提供者生成 RAG 回應
func (s *Service) GenerateRAGResponseWithEmbedding(ctx context.Context, query string, modelType llm.LLMType, embeddingType llm.EmbeddingType) (string, error) {
	// 1. 搜尋相關文檔（默認獲取前3個最相關的文檔）
	docs, err := s.docService.SearchSimilarDocumentsWithEmbedding(query, 3, embeddingType)
	if err != nil {
		return "", fmt.Errorf("搜尋相關文檔失敗: %v", err)
	}

	if len(docs) == 0 {
		return "沒有找到相關文檔，請嘗試其他問題。", nil
	}

	// 2. 構建提示詞
	prompt := buildRAGPrompt(query, docs)

	// 如果没有指定模型，默認使用 OpenAI
	if modelType == "" {
		modelType = llm.LLMTypeOpenAI
	}

	// 3. 獲取指定的 LLM 提供者
	llmProvider, err := s.llmFactory.Create(modelType)
	if err != nil {
		return "", fmt.Errorf("創建 LLM 客戶端失敗: %v", err)
	}
	defer llmProvider.Close()

	// 4. 使用 LLM 生成回應
	response, err := llmProvider.GenerateContent(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("生成回應失敗: %v", err)
	}

	return response, nil
}

// buildRAGPrompt 構建 RAG 提示詞
func buildRAGPrompt(query string, docs []documents.Document) string {
	var sb strings.Builder

	// 添加系統指令
	sb.WriteString("你是一個智能助手。請基於以下提供的上下文資料來回答用戶的問題。如果無法從上下文中找到答案，請誠實地說明你不知道，不要編造答案。\n\n")

	// 添加上下文（相關文檔）
	sb.WriteString("### 上下文資料：\n")
	for i, doc := range docs {
		sb.WriteString(fmt.Sprintf("文檔 %d：%s\n\n", i+1, doc.Text))
	}

	// 添加用戶問題
	sb.WriteString("### 用戶問題：\n")
	sb.WriteString(query)
	sb.WriteString("\n\n")
	sb.WriteString("### 回答：\n")

	return sb.String()
}
