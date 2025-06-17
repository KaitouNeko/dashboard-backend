package documents

import (
	"fmt"
	"log"
	"time"

	"ai-workshop/internal/config"
	llmtype "ai-workshop/internal/constants"
	"ai-workshop/internal/llm"
	"ai-workshop/internal/milvus"

	"github.com/google/uuid"
)

const (
	CollectionName = "documents"
)

type Document struct {
	ID     string    `json:"id"`
	Text   string    `json:"text"`
	Vector []float32 `json:"vector,omitempty"`
	Score  float64   `json:"score,omitempty"`
}

type Service struct {
	milvusClient    *milvus.Client
	embeddingClient *llm.OpenAIProvider
}

// NewService 創建文件服務
func NewService(appConfig *config.Config) *Service {
	config := &milvus.ClientConfig{
		BaseURL: "http://localhost:19530",
		Timeout: 10 * time.Second,
	}
	return &Service{
		milvusClient:    milvus.NewClient(config),
		embeddingClient: llm.NewOpenAiProvider(appConfig.OpenAiAPIKey),
	}
}

// ListCollections 列出所有集合
func (s *Service) ListCollections() ([]string, error) {
	return s.milvusClient.ListCollections()
}

// CreateDocumentCollection 創建文件集合
func (s *Service) CreateDocumentCollection() error {
	// 獲取默認嵌入維度
	dimension, err := s.embeddingClient.GetDimensionFor(llmtype.EmbeddingTypeOpenAI)
	if err != nil {
		dimension = llmtype.OpenAIModelDimension // 使用默認值
	}

	err = s.milvusClient.CreateCollection(CollectionName, dimension)
	if err != nil {
		return fmt.Errorf("創建集合失敗: %v", err)
	}
	return nil
}

// InsertDocument 插入單個文件（使用默認嵌入提供者）
func (s *Service) InsertDocument(text string) (string, error) {
	// 使用 UUID 生成唯一 ID
	id := uuid.New().String()
	return id, s.InsertDocumentWithID(id, text, llmtype.EmbeddingTypeOpenAI)
}

// InsertDocumentWithID 使用指定 ID 插入單個文件（內部使用）
func (s *Service) InsertDocumentWithID(id string, text string, embeddingType llmtype.EmbeddingType) error {
	// 1. 生成文本的嵌入向量
	vector, err := s.embeddingClient.CreateEmbeddingWith(embeddingType, text)
	if err != nil {
		return fmt.Errorf("生成嵌入向量失敗: %v", err)
	}

	// 確保集合存在
	dimension, err := s.embeddingClient.GetDimensionFor(embeddingType)
	if err != nil {
		return fmt.Errorf("獲取嵌入維度失敗: %v", err)
	}

	err = s.ensureCollectionWithDimension(dimension, embeddingType)
	if err != nil {
		return fmt.Errorf("確保集合存在失敗: %v", err)
	}

	// 2. 插入到 Milvus
	vectors := []map[string]interface{}{
		{
			"vector": vector,
			"text":   text,
		},
	}

	collectionName := getCollectionName(embeddingType)
	if err := s.milvusClient.InsertVectors(collectionName, vectors); err != nil {
		return fmt.Errorf("插入向量失敗: %v", err)
	}

	fmt.Printf("成功插入文件到集合 %s，文本: %s\n", collectionName, text)
	return nil
}

// InsertBatchDocuments 批量插入文件（使用默認嵌入提供者）
func (s *Service) InsertBatchDocuments(documents []Document) error {
	return s.InsertBatchDocumentsWithEmbedding(documents, llmtype.EmbeddingTypeOpenAI)
}

// InsertBatchDocumentsWithEmbedding 使用指定嵌入提供者批量插入文件
func (s *Service) InsertBatchDocumentsWithEmbedding(documents []Document, embeddingType llmtype.EmbeddingType) error {
	texts := make([]string, len(documents))
	for i, doc := range documents {
		texts[i] = doc.Text
	}

	// 批量生成嵌入向量
	vectors, err := s.embeddingClient.CreateBatchEmbeddingsWith(embeddingType, texts)
	if err != nil {
		return fmt.Errorf("批量生成嵌入向量失敗: %v", err)
	}

	// 確保集合存在
	dimension, err := s.embeddingClient.GetDimensionFor(embeddingType)
	if err != nil {
		return fmt.Errorf("獲取嵌入維度失敗: %v", err)
	}

	err = s.ensureCollectionWithDimension(dimension, embeddingType)
	if err != nil {
		return fmt.Errorf("確保集合存在失敗: %v", err)
	}

	// 準備插入數據
	insertData := make([]map[string]interface{}, len(documents))
	for i, doc := range documents {
		insertData[i] = map[string]interface{}{
			"vector": vectors[i],
			"text":   doc.Text,
		}
	}

	collectionName := getCollectionName(embeddingType)
	err = s.milvusClient.InsertVectors(collectionName, insertData)
	if err != nil {
		return fmt.Errorf("批量插入向量失敗: %v", err)
	}
	return nil
}

// ListVectors 列出所有文件（使用默認嵌入提供者）
func (s *Service) ListVectors() ([]Document, error) {
	return s.ListVectorsWithEmbedding(llmtype.EmbeddingTypeOpenAI)
}

// ListVectorsWithEmbedding 使用指定嵌入提供者列出所有文件
func (s *Service) ListVectorsWithEmbedding(embeddingType llmtype.EmbeddingType) ([]Document, error) {
	collectionName := getCollectionName(embeddingType)
	vectors, err := s.milvusClient.ListVectors(collectionName)
	log.Printf("ListVectors vectors: %v", vectors)
	log.Printf("ListVectors CollectionName: %v", collectionName)
	if err != nil {
		return nil, fmt.Errorf("獲取向量列表失敗: %v", err)
	}

	documents := make([]Document, 0, len(vectors))
	for _, v := range vectors {
		// 將所有類型的 ID 轉換為字符串
		var idStr string
		switch idVal := v["id"].(type) {
		case int64:
			idStr = fmt.Sprintf("%d", idVal)
		case int:
			idStr = fmt.Sprintf("%d", idVal)
		case float64:
			idStr = fmt.Sprintf("%d", int64(idVal))
		case float32:
			idStr = fmt.Sprintf("%d", int64(idVal))
		case string:
			idStr = idVal
		default:
			log.Printf("警告: 無法處理 ID 類型 %T: %v", v["id"], v["id"])
			continue
		}

		// 處理 text 欄位
		var text string
		if textVal, ok := v["text"].(string); ok {
			text = textVal
		} else {
			log.Printf("警告: 無法處理 text 類型 %T: %v", v["text"], v["text"])
			continue
		}

		// 創建文檔對象
		doc := Document{
			ID:   idStr,
			Text: text,
		}
		log.Printf("處理文檔 - 原始ID: %v, 轉換後ID: %s", v["id"], idStr)
		documents = append(documents, doc)
	}

	log.Printf("ListVectors 處理完成，共 %d 個文檔", len(documents))
	return documents, nil
}

// DeleteDocument 刪除文件（使用默認嵌入提供者）
func (s *Service) DeleteDocument(id string) error {
	return s.DeleteDocumentWithEmbedding(id, llmtype.EmbeddingTypeOpenAI)
}

// DeleteDocumentWithEmbedding 使用指定嵌入提供者刪除文件
func (s *Service) DeleteDocumentWithEmbedding(id string, embeddingType llmtype.EmbeddingType) error {
	ids := []string{id}
	collectionName := getCollectionName(embeddingType)
	err := s.milvusClient.DeleteVectors(collectionName, ids)
	if err != nil {
		return fmt.Errorf("刪除文件失敗: %v", err)
	}
	return nil
}

// DeleteDocuments 批量刪除文件（使用默認嵌入提供者）
func (s *Service) DeleteDocuments(ids []string) error {
	return s.DeleteDocumentsWithEmbedding(ids, llmtype.EmbeddingTypeOpenAI)
}

// DeleteDocumentsWithEmbedding 使用指定嵌入提供者批量刪除文件
func (s *Service) DeleteDocumentsWithEmbedding(ids []string, embeddingType llmtype.EmbeddingType) error {
	collectionName := getCollectionName(embeddingType)
	err := s.milvusClient.DeleteVectors(collectionName, ids)
	if err != nil {
		return fmt.Errorf("批量刪除文件失敗: %v", err)
	}
	return nil
}

// DeleteCollection 刪除整個文件集合（使用默認嵌入提供者）
func (s *Service) DeleteCollection() error {
	return s.DeleteCollectionWithEmbedding(llmtype.EmbeddingTypeOpenAI)
}

// DeleteCollectionWithEmbedding 使用指定嵌入提供者刪除整個文件集合
func (s *Service) DeleteCollectionWithEmbedding(embeddingType llmtype.EmbeddingType) error {
	collectionName := getCollectionName(embeddingType)
	err := s.milvusClient.DeleteCollection(collectionName)
	if err != nil {
		return fmt.Errorf("刪除集合失敗: %v", err)
	}
	return nil
}

// SearchSimilarDocuments 搜尋相似文檔（使用默認嵌入提供者）
func (s *Service) SearchSimilarDocuments(query string, topK int) ([]Document, error) {
	// 使用默認嵌入提供者
	return s.SearchSimilarDocumentsWithEmbedding(query, topK, llmtype.EmbeddingTypeOpenAI)
}

// SearchSimilarDocumentsWithEmbedding 使用指定嵌入提供者搜尋相似文檔
func (s *Service) SearchSimilarDocumentsWithEmbedding(query string, topK int, embeddingType llmtype.EmbeddingType) ([]Document, error) {
	// 1. 生成查詢文本的嵌入向量
	queryVector, err := s.embeddingClient.CreateEmbeddingWith(embeddingType, query)
	if err != nil {
		return nil, fmt.Errorf("生成查詢嵌入向量失敗: %v", err)
	}

	// 獲取嵌入維度
	dimension, err := s.embeddingClient.GetDimensionFor(embeddingType)
	if err != nil {
		return nil, fmt.Errorf("獲取嵌入維度失敗: %v", err)
	}

	// 確保集合存在並具有正確的維度
	err = s.ensureCollectionWithDimension(dimension, embeddingType)
	if err != nil {
		return nil, fmt.Errorf("確保集合存在失敗: %v", err)
	}

	// 2. 使用向量在 Milvus 中搜尋相似文檔
	collectionName := getCollectionName(embeddingType)
	results, err := s.milvusClient.Search(collectionName, queryVector, topK)
	if err != nil {
		return nil, fmt.Errorf("搜尋相似文檔失敗: %v", err)
	}

	// 3. 將搜尋結果轉換為文檔對象
	documents := make([]Document, 0, len(results))
	for _, result := range results {
		// 小心處理不同類型的強制轉換
		var idStr string
		switch id := result["id"].(type) {
		case string:
			idStr = id
		case int64:
			idStr = fmt.Sprintf("%d", id)
		case float64:
			idStr = fmt.Sprintf("%d", int64(id))
		default:
			idStr = fmt.Sprintf("%v", id)
		}

		var text string
		if textValue, ok := result["text"].(string); ok {
			text = textValue
		} else {
			log.Printf("警告: 無法解析 text 欄位: %v", result["text"])
			continue
		}

		var score float64
		if scoreValue, ok := result["score"].(float64); ok {
			score = scoreValue
		}

		doc := Document{
			ID:    idStr,
			Text:  text,
			Score: score,
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// 確保集合存在並有正確的維度
func (s *Service) ensureCollectionWithDimension(dimension int, embeddingType llmtype.EmbeddingType) error {
	collectionName := getCollectionName(embeddingType)

	// 檢查集合是否存在
	exists, err := s.milvusClient.CollectionExists(collectionName)
	if err != nil {
		return fmt.Errorf("檢查集合存在失敗: %v", err)
	}

	if !exists {
		// 創建集合
		err = s.milvusClient.CreateCollection(collectionName, dimension)
		if err != nil {
			return fmt.Errorf("創建集合失敗: %v", err)
		}
		log.Printf("已創建集合 %s，維度: %d", collectionName, dimension)
	}

	return nil
}

// 根據嵌入類型獲取集合名稱
func getCollectionName(embeddingType llmtype.EmbeddingType) string {
	switch embeddingType {
	case llmtype.EmbeddingTypeGemini:
		return CollectionName + "_gemini"
	case llmtype.EmbeddingTypeOpenAI:
		return CollectionName
	case llmtype.EmbeddingTypeOpenAI3Small:
		return CollectionName + "_openai3small"
	case llmtype.EmbeddingTypeOpenAI3Large:
		return CollectionName + "_openai3large"
	default:
		return CollectionName
	}
}
