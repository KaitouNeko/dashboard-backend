package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type EmbeddingModelType string

const (
	EmbeddingModelTypeAda002 EmbeddingModelType = "ada-002"
	EmbeddingModelType3Small EmbeddingModelType = "3-small"
	EmbeddingModelType3Large EmbeddingModelType = "3-large"
	EmbeddingModelTypeGemini EmbeddingModelType = "gemini"
)

// OpenAIClient represents an OpenAI API client
type OpenAIClient struct {
	Client    *openai.Client
	ModelType EmbeddingModelType // 添加modelType欄位
}

const (
	OpenAIModelDimension = 1536 // OpenAI text-embedding-ada-002 的維度
	GeminiModelDimension = 768  // Google Gemini Embedding 的維度
)

// EmbeddingType 表示支援的嵌入模型類型
type EmbeddingType string

const (
	EmbeddingTypeOpenAI       EmbeddingType = "openai-ada-002"
	EmbeddingTypeGemini       EmbeddingType = "gemini-embedding"
	EmbeddingTypeOpenAI3Small EmbeddingType = "openai-3-small"
	EmbeddingTypeOpenAI3Large EmbeddingType = "openai-3-large"
)

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
	return s.CreateEmbeddingWith(EmbeddingTypeOpenAI, text)
}

// CreateEmbeddingWith 使用指定的嵌入模型創建嵌入向量
func (s *OpenAIProvider) CreateEmbeddingWith(embeddingType EmbeddingType, text string) ([]float32, error) {
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
	return s.CreateBatchEmbeddingsWith(EmbeddingTypeOpenAI, texts)
}

// CreateBatchEmbeddingsWith 使用指定的嵌入模型批量創建嵌入向量
func (s *OpenAIProvider) CreateBatchEmbeddingsWith(embeddingType EmbeddingType, texts []string) ([][]float32, error) {
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
func (s *OpenAIProvider) GetDimensionFor(embeddingType EmbeddingType) (int, error) {
	switch embeddingType {
	case EmbeddingTypeOpenAI, EmbeddingTypeOpenAI3Large:
		return OpenAIModelDimension, nil
	case EmbeddingTypeOpenAI3Small:
		return 1536, nil // OpenAI-3-Small 維度
	case EmbeddingTypeGemini:
		return GeminiModelDimension, nil
	default:
		return 0, fmt.Errorf("不支援的嵌入模型類型: %s", embeddingType)
	}
}

// getModelForType 根據嵌入類型獲取實際的模型名稱
func (s *OpenAIProvider) getModelForType(embeddingType EmbeddingType) openai.EmbeddingModel {
	switch embeddingType {
	case EmbeddingTypeOpenAI:
		return openai.AdaEmbeddingV2
	case EmbeddingTypeOpenAI3Small:
		return openai.EmbeddingModel("text-embedding-3-small")
	case EmbeddingTypeOpenAI3Large:
		return openai.EmbeddingModel("text-embedding-3-large")
	default:
		// 默認使用OpenAI的Ada-002
		return openai.AdaEmbeddingV2
	}
}
