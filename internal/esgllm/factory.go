package esgllm

import (
	"ai-workshop/internal/config"
	llmtype "ai-workshop/internal/constants"
	"context"
	"fmt"
)

type LLMProvider interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
	CreateEmbedding(text string) ([]float32, error)
	CreateBatchEmbeddings(texts []string) ([][]float32, error)
	Close()
}

// factory that generates more llm constructors, e.g. openAI llm constructor
type Factory struct {
	config *config.Config
}

// NewFactory 創建一個新的 LLM 工廠
func NewFactory(config *config.Config) *Factory {
	return &Factory{
		config: config,
	}
}

func (f *Factory) CreateOpenAi() LLMProvider {
	provider := NewOpenAiProvider(f.config.OpenAiAPIKey)
	return provider
}

func (f *Factory) CreateGemini() (LLMProvider, error) {
	provider, err := NewGeminiProvider(f.config.GeminiAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini provider: %v", err)
	}
	return provider, nil
}

func (f *Factory) CreateWatsonx() (LLMProvider, error) {
	provider := NewWatsonxProvider(f.config.WatsonxAPIKey)
	return provider, nil
}

// a factory that generates LLM creators
func (f *Factory) Create(llmType llmtype.LLMType) (LLMProvider, error) {
	switch llmType {
	case llmtype.LLMTypeOpenAI:
		provider := NewOpenAiProvider(f.config.OpenAiAPIKey)
		return provider, nil

	case llmtype.LLMTypeGemini:
		provider, err := NewGeminiProvider(f.config.GeminiAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini provider: %v", err)
		}
		return provider, nil

	default:
		msg := "No llm constructor match llmtype provided."
		fmt.Println(msg)
		return nil, fmt.Errorf(msg)
	}
}
