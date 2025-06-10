package uploads

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 檔案類型分類常量
const (
	FileTypeText     = "text"     // 文字檔案
	FileTypeAudio    = "audio"    // 音訊檔案
	FileTypeImage    = "image"    // 圖片檔案
	FileTypeVideo    = "video"    // 影片檔案
	FileTypeDocument = "document" // 文件檔案
	FileTypeOther    = "other"    // 其他類型
)

// EmbeddingModel 代表嵌入模型配置
type EmbeddingModel struct {
	Name        string `json:"name"`        // 模型名稱
	Description string `json:"description"` // 模型描述
}

// 預設嵌入模型映射表
var defaultEmbeddingModels = map[string]EmbeddingModel{
	FileTypeText:     {Name: "text-embedding-3-small", Description: "適用於文字檔案的嵌入模型"},
	FileTypeAudio:    {Name: "openai/whisper-base", Description: "適用於音訊檔案的嵌入模型"},
	FileTypeImage:    {Name: "clip", Description: "適用於圖片檔案的嵌入模型"},
	FileTypeVideo:    {Name: "openai/whisper-base", Description: "適用於影片檔案的嵌入模型，處理音訊部分"},
	FileTypeDocument: {Name: "text-embedding-3-small", Description: "適用於文件檔案的嵌入模型"},
	FileTypeOther:    {Name: "text-embedding-3-small", Description: "適用於其他類型檔案的嵌入模型"},
}

// ServiceConfig 代表上傳服務的配置選項
type ServiceConfig struct {
	UploadDir        string   // 上傳檔案的儲存目錄
	MaxFileSize      int64    // 允許的最大檔案大小
	AllowedFileTypes []string // 允許的檔案類型
}

// UploadedFile 代表已上傳的檔案信息
type UploadedFile struct {
	FileName       string    `json:"fileName"`       // 儲存的檔案名稱
	OriginalName   string    `json:"originalName"`   // 原始檔案名稱
	FilePath       string    `json:"filePath"`       // 檔案完整路徑
	FileSize       int64     `json:"fileSize"`       // 檔案大小
	MimeType       string    `json:"mimeType"`       // MIME 類型
	FileType       string    `json:"fileType"`       // 檔案分類類型
	EmbeddingModel string    `json:"embeddingModel"` // 建議的嵌入模型
	UploadTime     time.Time `json:"uploadTime"`     // 上傳時間
}

// FileUploadService 是 FileService 介面的實作
type FileUploadService struct {
	config ServiceConfig
}

// NewFileService 創建一個新的檔案服務
func NewFileService(config ServiceConfig) *FileUploadService {
	// 設置默認值
	if config.UploadDir == "" {
		config.UploadDir = "./uploads"
	}

	if config.MaxFileSize <= 0 {
		config.MaxFileSize = 10 << 20 // 10 MB
	}

	if len(config.AllowedFileTypes) == 0 {
		config.AllowedFileTypes = []string{ // default allowed file types
			".pdf", ".doc", ".docx", ".txt", ".csv", ".xls", ".xlsx", ".json",
			".jpg", ".jpeg", ".png", ".gif", ".mp3", ".mp4", ".wav", ".ogg", ".m4a", ".html",
		}
	}

	// 確保上傳目錄存在
	os.MkdirAll(config.UploadDir, 0755)

	return &FileUploadService{
		config: config,
	}
}

func (s *FileUploadService) SaveFile(file *multipart.FileHeader) (*UploadedFile, error) {
	// 檢查檔案大小
	if file.Size > s.config.MaxFileSize {
		return nil, fmt.Errorf("檔案大小超過限制: %d bytes", file.Size)
	}

	// 檢查檔案類型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !s.AllowedFileTypes(ext) {
		return nil, fmt.Errorf("不允許的檔案類型: %s", ext)
	}

	// 生成唯一的檔案名稱
	timestamp := time.Now().UnixNano()
	fileName := fmt.Sprintf("%d%s", timestamp, ext)

	// 生成完整的檔案路徑
	filePath := filepath.Join(s.config.UploadDir, fileName)

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("無法開啟檔案: %v", err)
	}
	defer src.Close()

	// 創建目標檔案
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("無法創建檔案: %v", err)
	}
	defer dst.Close()

	// 複製檔案內容
	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("無法複製檔案內容: %v", err)
	}

	// 判斷檔案類型
	fileType := s.GetFileType(ext)

	// 獲取對應的嵌入模型
	embeddingModel := s.GetEmbeddingModel(fileType)

	// 返回上傳的檔案資訊
	return &UploadedFile{
		FileName:       fileName,
		OriginalName:   file.Filename,
		FilePath:       filePath,
		FileSize:       file.Size,
		MimeType:       file.Header.Get("Content-Type"),
		FileType:       fileType,
		EmbeddingModel: embeddingModel.Name,
		UploadTime:     time.Now(),
	}, nil
}

func (s *FileUploadService) SaveFiles(files []*multipart.FileHeader) ([]*UploadedFile, error) {
	result := make([]*UploadedFile, 0, len(files))

	for _, file := range files {
		uploadedFile, err := s.SaveFile(file)
		if err != nil {
			return nil, fmt.Errorf("儲存檔案失敗: %v", err)
		}
		result = append(result, uploadedFile)
	}
	return result, nil
}

// ListFiles 實作檔案列表邏輯
func (s *FileUploadService) ListFiles() ([]*UploadedFile, error) {
	files := make([]*UploadedFile, 0)

	// 讀取上傳目錄中的所有檔案
	entries, err := os.ReadDir(s.config.UploadDir)
	if err != nil {
		return nil, fmt.Errorf("讀取上傳目錄失敗：%v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // 跳過子目錄
		}

		info, err := entry.Info()
		if err != nil {
			continue // 跳過讀取失敗的檔案
		}

		filePath := filepath.Join(s.config.UploadDir, entry.Name())

		file := &UploadedFile{
			FileName:     entry.Name(),
			OriginalName: entry.Name(), // 原始檔名無法恢復，使用當前檔名
			FilePath:     filePath,
			FileSize:     info.Size(),
			UploadTime:   info.ModTime(),
		}

		files = append(files, file)
	}

	return files, nil
}

// DeleteFile 實作檔案刪除邏輯
func (s *FileUploadService) DeleteFile(fileName string) error {
	// 確保檔案名稱合法
	if fileName == "" {
		return errors.New("檔案名稱不能為空")
	}

	// 處理檔案路徑
	filePath := filepath.Join(s.config.UploadDir, fileName)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("無法獲取絕對路徑：%v", err)
	}

	// 獲取上傳目錄的絕對路徑
	absUploadDir, err := filepath.Abs(s.config.UploadDir)
	if err != nil {
		return fmt.Errorf("無法獲取上傳目錄絕對路徑：%v", err)
	}

	// 確保檔案在上傳目錄中，防止路徑穿越攻擊
	if !strings.HasPrefix(absPath, absUploadDir) {
		return errors.New("無效的檔案路徑")
	}

	// 檢查檔案是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("檔案不存在：%s", fileName)
	}

	// 刪除檔案
	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("刪除檔案失敗：%v", err)
	}

	return nil
}

// GetFilePath 實作獲取檔案路徑邏輯
func (s *FileUploadService) GetFilePath(fileName string) (string, error) {
	// 確保檔案名稱合法
	if fileName == "" {
		return "", errors.New("檔案名稱不能為空")
	}

	// 處理檔案路徑
	filePath := filepath.Join(s.config.UploadDir, fileName)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("無法獲取絕對路徑：%v", err)
	}

	// 獲取上傳目錄的絕對路徑
	absUploadDir, err := filepath.Abs(s.config.UploadDir)
	if err != nil {
		return "", fmt.Errorf("無法獲取上傳目錄絕對路徑：%v", err)
	}

	// 確保檔案在上傳目錄中，防止路徑穿越攻擊
	if !strings.HasPrefix(absPath, absUploadDir) {
		return "", errors.New("無效的檔案路徑")
	}

	// 檢查檔案是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("檔案不存在：%s", fileName)
	}

	return absPath, nil
}

// isAllowedFileType 檢查檔案類型是否允許
func (s *FileUploadService) AllowedFileTypes(ext string) bool {
	for _, allowedExt := range s.config.AllowedFileTypes {
		if strings.EqualFold(allowedExt, ext) {
			return true
		}
	}
	return false
}

// GetFileType 根據檔案副檔名判斷檔案類型
func (s *FileUploadService) GetFileType(ext string) string {
	switch strings.ToLower(ext) {
	case ".txt", ".csv", ".json", ".html", ".xml", ".md":
		return FileTypeText
	case ".mp3", ".wav", ".ogg", ".m4a":
		return FileTypeAudio
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return FileTypeImage
	case ".mp4", ".avi", ".mov", ".mkv", ".webm":
		return FileTypeVideo
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return FileTypeDocument
	default:
		return FileTypeOther
	}
}

// GetEmbeddingModel 根據檔案類型獲取對應的嵌入模型
func (s *FileUploadService) GetEmbeddingModel(fileType string) EmbeddingModel {
	if model, exists := defaultEmbeddingModels[fileType]; exists {
		return model
	}
	return defaultEmbeddingModels[FileTypeOther]
}
