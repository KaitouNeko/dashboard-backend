package uploads

import (
	llmtype "ai-workshop/internal/constants"
	"ai-workshop/internal/llm"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EmbeddingProcessor 處理不同類型檔案的嵌入向量生成
type EmbeddingProcessor struct {
	openAIProvider *llm.OpenAIProvider
}

// NewEmbeddingProcessor 創建新的嵌入處理器
func NewEmbeddingProcessor(openAIKey string) *EmbeddingProcessor {
	return &EmbeddingProcessor{
		openAIProvider: llm.NewOpenAiProvider(openAIKey),
	}
}

// ProcessFile 根據檔案類型和指定的模型處理檔案
func (p *EmbeddingProcessor) ProcessFile(filePath string, fileType string, modelName string) (interface{}, error) {
	switch fileType {
	case FileTypeText:
		return p.processTextFile(filePath, modelName)
	case FileTypeAudio:
		return p.processAudioFile(filePath, modelName)
	case FileTypeImage:
		return p.processImageFile(filePath, modelName)
	case FileTypeVideo:
		return p.processVideoFile(filePath, modelName)
	case FileTypeDocument:
		return p.processDocumentFile(filePath, modelName)
	default:
		return nil, fmt.Errorf("不支援的檔案類型：%s", fileType)
	}
}

// processTextFile 處理文字類型檔案
func (p *EmbeddingProcessor) processTextFile(filePath string, modelName string) (interface{}, error) {
	// 讀取檔案內容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("讀取檔案失敗：%v", err)
	}

	textContent := string(content)

	// 選擇合適的嵌入模型
	var embeddingType llmtype.EmbeddingType
	switch modelName {
	case "text-embedding-3-small":
		embeddingType = llmtype.EmbeddingTypeOpenAI3Small
	case "text-embedding-3-large":
		embeddingType = llmtype.EmbeddingTypeOpenAI3Large
	case "text-embedding-ada-002":
		embeddingType = llmtype.EmbeddingTypeOpenAI
	default:
		// 預設使用 OpenAI 的 3-small 模型
		embeddingType = llmtype.EmbeddingTypeOpenAI3Small
	}

	// 產生嵌入向量
	embedding, err := p.openAIProvider.CreateEmbeddingWith(embeddingType, textContent)
	if err != nil {
		return nil, fmt.Errorf("生成嵌入向量失敗：%v", err)
	}

	// 獲取模型維度
	dimension, _ := p.openAIProvider.GetDimensionFor(embeddingType)

	// 返回結果
	result := map[string]interface{}{
		"model":       modelName,
		"fileType":    FileTypeText,
		"contentSize": len(textContent),
		"dimension":   dimension,
		"embedding":   embedding[:10],                 // 只返回前10個元素作為預覽，避免過大的響應
		"textPreview": truncateText(textContent, 200), // 截取前200個字符作為預覽
	}

	return result, nil
}

// truncateText 截取文本並確保不會中斷單詞
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	// 尋找最後一個空格，確保不會切斷詞彙
	truncated := text[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")

	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// processAudioFile 處理音訊類型檔案
func (p *EmbeddingProcessor) processAudioFile(filePath string, modelName string) (interface{}, error) {
	// 獲取檔案資訊
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("獲取檔案資訊失敗：%v", err)
	}

	// 對於音訊檔案，我們需要:
	// 1. 首先使用 Whisper 模型進行轉錄
	// 2. 然後對轉錄文本生成嵌入向量

	// 注意：這裡我們模擬 Whisper API 調用
	// 實際實現中，您需要調用 OpenAI Whisper API 或其他音訊轉錄服務

	// 模擬轉錄結果
	transcription := fmt.Sprintf("這是從音訊檔案 %s 生成的模擬轉錄文字。實際應用中，此處應為 Whisper API 的轉錄結果。", filepath.Base(filePath))

	// 使用轉錄文本生成嵌入向量
	embedding, err := p.openAIProvider.CreateEmbeddingWith(llmtype.EmbeddingTypeOpenAI3Small, transcription)
	if err != nil {
		return nil, fmt.Errorf("生成嵌入向量失敗：%v", err)
	}

	// 獲取模型維度
	dimension, _ := p.openAIProvider.GetDimensionFor(llmtype.EmbeddingTypeOpenAI3Small)

	// 返回結果
	result := map[string]interface{}{
		"model":         modelName,
		"fileType":      FileTypeAudio,
		"fileSize":      fileInfo.Size(),
		"transcription": transcription,
		"dimension":     dimension,
		"embedding":     embedding[:10], // 只返回前10個元素作為預覽
	}

	return result, nil
}

// processImageFile 處理圖片類型檔案
func (p *EmbeddingProcessor) processImageFile(filePath string, modelName string) (interface{}, error) {
	// 獲取檔案資訊
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("獲取檔案資訊失敗：%v", err)
	}

	// 根據選定的模型處理圖片檔案
	// 這裡是示例，實際應調用 CLIP 等 API
	result := map[string]interface{}{
		"model":       modelName,
		"fileType":    FileTypeImage,
		"fileSize":    fileInfo.Size(),
		"description": "模擬圖片描述文字",
		"embedding":   "模擬圖片嵌入向量",
	}

	return result, nil
}

// processVideoFile 處理影片類型檔案
func (p *EmbeddingProcessor) processVideoFile(filePath string, modelName string) (interface{}, error) {
	// 獲取檔案資訊
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("獲取檔案資訊失敗：%v", err)
	}

	// 根據選定的模型處理影片檔案
	// 這裡是示例，實際應提取音訊並調用 Whisper API 等
	result := map[string]interface{}{
		"model":         modelName,
		"fileType":      FileTypeVideo,
		"fileSize":      fileInfo.Size(),
		"transcription": "模擬影片音訊轉錄文字",
		"embedding":     "模擬影片嵌入向量",
	}

	return result, nil
}

// processDocumentFile 處理文件類型檔案
func (p *EmbeddingProcessor) processDocumentFile(filePath string, modelName string) (interface{}, error) {
	// 獲取檔案資訊
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("獲取檔案資訊失敗：%v", err)
	}

	// 對於文件檔案，實際實現中需要：
	// 1. 根據文件類型(PDF、DOCX、XLSX等)提取文本內容
	// 2. 對提取的文本生成嵌入向量

	// 根據檔案副檔名決定處理方式
	ext := strings.ToLower(filepath.Ext(filePath))

	// 模擬文本提取 (實際應該使用專門的庫進行處理)
	extractedText := fmt.Sprintf("這是從%s檔案提取的模擬文本內容。實際應用中，應根據檔案類型(%s)使用相應的庫提取文本。",
		filepath.Base(filePath), ext)

	// 使用文本生成嵌入向量
	embedding, err := p.openAIProvider.CreateEmbeddingWith(llmtype.EmbeddingTypeOpenAI3Small, extractedText)
	if err != nil {
		return nil, fmt.Errorf("生成嵌入向量失敗：%v", err)
	}

	// 獲取模型維度
	dimension, _ := p.openAIProvider.GetDimensionFor(llmtype.EmbeddingTypeOpenAI3Small)

	// 返回結果
	result := map[string]interface{}{
		"model":     modelName,
		"fileType":  FileTypeDocument,
		"fileSize":  fileInfo.Size(),
		"content":   truncateText(extractedText, 200), // 截取前200個字符作為預覽
		"dimension": dimension,
		"embedding": embedding[:10], // 只返回前10個元素作為預覽
		"docType":   ext,
	}

	return result, nil
}
