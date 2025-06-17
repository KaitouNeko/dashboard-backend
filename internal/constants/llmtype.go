package llmtype

type LLMType string

const (
	LLMTypeGemini  LLMType = "gemini"
	LLMTypeOpenAI  LLMType = "openai"
	LLMTypeWatsonx LLMType = "watsonx"
)

type EmbeddingModelType string

const (
	EmbeddingModelTypeAda002 EmbeddingModelType = "ada-002"
	EmbeddingModelType3Small EmbeddingModelType = "3-small"
	EmbeddingModelType3Large EmbeddingModelType = "3-large"
	EmbeddingModelTypeGemini EmbeddingModelType = "gemini"
)

// EmbeddingType 表示支援的嵌入模型類型
type EmbeddingType string

const (
	EmbeddingTypeOpenAI       EmbeddingType = "openai-ada-002"
	EmbeddingTypeGemini       EmbeddingType = "gemini-embedding"
	EmbeddingTypeOpenAI3Small EmbeddingType = "openai-3-small"
	EmbeddingTypeOpenAI3Large EmbeddingType = "openai-3-large"
)

const (
	OpenAIModelDimension = 1536 // OpenAI text-embedding-ada-002 的維度
	GeminiModelDimension = 768  // Google Gemini Embedding 的維度
)
