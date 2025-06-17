package esgllm

import (
	llmtype "ai-workshop/internal/constants"
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// OpenAIClient represents an OpenAI API client
type OpenAIClient struct {
	Client    *openai.Client
	ModelType llmtype.EmbeddingModelType // 添加modelType欄位
}

// OpenAIProvider 是 OpenAI 客戶端適配器
type OpenAIProvider struct {
	client *OpenAIClient
}

func NewOpenAiProvider(apiKey string) *OpenAIProvider {
	openAIClient := openai.NewClient(apiKey)

	client := &OpenAIClient{
		Client: openAIClient,
	}

	return &OpenAIProvider{
		client: client,
	}
}

func (p *OpenAIProvider) GenerateContent(ctx context.Context, prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是一個擅長處理ESG指標的助手，請根據用戶的問題提供準確和有用的回答。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}
	resp, err := p.client.Client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("生成內容失敗: %v", err)
	}

	if len(resp.Choices) == 0 {
		// TODO: update to custom error
		return "", fmt.Errorf("No valid responses")
	}

	return resp.Choices[0].Message.Content, nil
}

// Close 關閉客戶端
func (p *OpenAIProvider) Close() {
	// OpenAI 客戶端不需要關閉操作
}

// CreateEmbedding 創建單個文本的嵌入向量
func (s *OpenAIProvider) CreateEmbedding(text string) ([]float32, error) {
	return s.CreateEmbeddingWith(llmtype.EmbeddingTypeOpenAI, text)
}

// CreateEmbeddingWith 使用指定的嵌入模型創建嵌入向量
func (s *OpenAIProvider) CreateEmbeddingWith(embeddingType llmtype.EmbeddingType, text string) ([]float32, error) {
	model := s.getModelForType(embeddingType)

	resp, err := s.client.Client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: model,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("創建嵌入向量失敗: %v", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("未獲得嵌入向量")
	}

	return resp.Data[0].Embedding, nil
}

// CreateBatchEmbeddings 批量創建文本的嵌入向量
func (s *OpenAIProvider) CreateBatchEmbeddings(texts []string) ([][]float32, error) {
	return s.CreateBatchEmbeddingsWith(llmtype.EmbeddingTypeOpenAI, texts)
}

// CreateBatchEmbeddingsWith 使用指定的嵌入模型批量創建嵌入向量
func (s *OpenAIProvider) CreateBatchEmbeddingsWith(embeddingType llmtype.EmbeddingType, texts []string) ([][]float32, error) {
	model := s.getModelForType(embeddingType)

	resp, err := s.client.Client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: texts,
			Model: model,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("批量創建嵌入向量失敗: %v", err)
	}
	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

// GetDimensionFor 獲取指定嵌入模型的維度
func (s *OpenAIProvider) GetDimensionFor(embeddingType llmtype.EmbeddingType) (int, error) {
	switch embeddingType {
	case llmtype.EmbeddingTypeOpenAI, llmtype.EmbeddingTypeOpenAI3Large:
		return llmtype.OpenAIModelDimension, nil
	case llmtype.EmbeddingTypeOpenAI3Small:
		return 1536, nil // OpenAI-3-Small 維度
	case llmtype.EmbeddingTypeGemini:
		return llmtype.GeminiModelDimension, nil
	default:
		return 0, fmt.Errorf("不支援的嵌入模型類型: %s", embeddingType)
	}
}

// getModelForType 根據嵌入類型獲取實際的模型名稱
func (s *OpenAIProvider) getModelForType(embeddingType llmtype.EmbeddingType) openai.EmbeddingModel {
	switch embeddingType {
	case llmtype.EmbeddingTypeOpenAI:
		return openai.AdaEmbeddingV2
	case llmtype.EmbeddingTypeOpenAI3Small:
		return openai.EmbeddingModel("text-embedding-3-small")
	case llmtype.EmbeddingTypeOpenAI3Large:
		return openai.EmbeddingModel("text-embedding-3-large")
	default:
		// 默認使用OpenAI的Ada-002
		return openai.AdaEmbeddingV2
	}
}
