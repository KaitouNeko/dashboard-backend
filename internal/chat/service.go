package chat

import (
	"context"
	"fmt"
	"strings"

	"ai-workshop/internal/config"
	llmtype "ai-workshop/internal/constants"
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
func (s *Service) GenerateRAGResponse(ctx context.Context, query string, modelType llmtype.LLMType) (string, error) {
	// 使用默認的嵌入提供者
	return s.GenerateRAGResponseWithEmbedding(ctx, query, modelType, llmtype.EmbeddingTypeOpenAI)
}

// GenerateRAGResponseWithEmbedding 使用指定的嵌入提供者生成 RAG 回應
func (s *Service) GenerateRAGResponseWithEmbedding(ctx context.Context, query string, modelType llmtype.LLMType, embeddingType llmtype.EmbeddingType) (string, error) {
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
		modelType = llmtype.LLMTypeOpenAI
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
	sb.WriteString("你是一個優秀的智能助手，專精於提供精準的推薦清單。請基於以下提供的上下文資料來回答用戶的問題。\n\n")
	sb.WriteString("**回答格式要求：**\n")
	sb.WriteString("1. 優先列出「推薦清單」（如果資料中有相關產品/項目）\n")
	sb.WriteString("2. 在清單後明確說明「推薦依據」（例如：評分最高、銷量最佳、客戶喜好、價格優勢等）\n")
	sb.WriteString("3. 如果上下文資料不足以提供完整清單，才詢問用戶更多偏好來協助推薦\n")
	sb.WriteString("4. 每個推薦項目應包含名稱和簡短說明\n")
	sb.WriteString("5. 如果無法從上下文中找到相關答案，請誠實地說明你不知道，不要編造答案\n\n")
	sb.WriteString("**特殊處理規則：**\n")
	sb.WriteString("- 如果問題涉及「進銷存」、「庫存」、「產品管理」、「商品」等相關內容，請務必使用 Product ID 來關聯和整理資料\n")
	sb.WriteString("- 進銷存相關回答應包含：Product ID、產品名稱、庫存狀況、價格等關鍵資訊\n")
	sb.WriteString("- 按 Product ID 順序或相關性排列清單\n\n")

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
