package esgllm

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient represents a Gemini API client
type GeminiClient struct {
	client         *genai.Client
	model          *genai.GenerativeModel
	embeddingModel *genai.EmbeddingModel
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiProvider(apiKey string) (*GeminiClient, error) {
	// 使用配置中的 API Key
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("創建 Gemini 客戶端失敗: %v", err)
	}

	// 創建生成模型
	model := client.GenerativeModel("gemini-2.0-flash")

	// 創建嵌入模型
	embeddingModel := client.EmbeddingModel("embedding-001")
	return &GeminiClient{
		client:         client,
		model:          model,
		embeddingModel: embeddingModel,
	}, nil
}

// GenerateContent generates content using the Gemini model
func (c *GeminiClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	systemPrompt := "你是一個擅長處理ESG指標的助手，請根據用戶的問題提供準確和有用的回答。"
	fullPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, prompt)
	resp, err := c.model.GenerateContent(ctx, genai.Text(fullPrompt))

	if err != nil {
		return "", fmt.Errorf("生成內容失敗: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", ErrNoValidResponse
	}

	return fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0]), nil
}

// CreateEmbedding creates an embedding for the given text
func (c *GeminiClient) CreateEmbedding(text string) ([]float32, error) {
	ctx := context.Background()
	if c.embeddingModel == nil {
		return nil, fmt.Errorf("嵌入模型未初始化")
	}

	// 使用Gemini的EmbedContent方法生成嵌入
	res, err := c.embeddingModel.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("生成嵌入向量失敗: %v", err)
	}

	// 檢查嵌入是否成功生成
	if res.Embedding == nil || len(res.Embedding.Values) == 0 {
		return nil, fmt.Errorf("未獲得嵌入向量")
	}

	return res.Embedding.Values, nil
}

// CreateBatchEmbeddings creates batch embeddings for the given texts
func (c *GeminiClient) CreateBatchEmbeddings(texts []string) ([][]float32, error) {
	if c.embeddingModel == nil {
		return nil, fmt.Errorf("嵌入模型未初始化")
	}

	// 由於Gemini API目前不支持批量嵌入，我們需要逐個處理
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := c.CreateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("批量創建嵌入向量失敗 (第 %d 項): %v", i+1, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// Close closes the Gemini client
func (c *GeminiClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
