package uploads

import (
	"ai-workshop/internal/config"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	service FileUploadService
	config  *config.Config
}

func NewFileHandler(service FileUploadService, config *config.Config) *FileHandler {
	return &FileHandler{
		service: service,
		config:  config,
	}
}

// HandleUpload 處理單一檔案上傳
func (h *FileHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "儲存檔案失敗: " + err.Error()})
		return
	}

	uploadedFile, err := h.service.SaveFile(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "儲存檔案失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "檔案上傳成功",
		"file":    uploadedFile,
	})
}

// HandleUploads 處理多個檔案上傳
func (h *FileHandler) UploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "讀取表單資料失敗: " + err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請至少上傳一個檔案"})
		return
	}
	// 儲存所有檔案
	uploadedFiles, err := h.service.SaveFiles(files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "儲存檔案失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "檔案上傳成功",
		"files":   uploadedFiles,
	})
}

// HandleListFiles 列出所有已上傳的檔案
func (h *FileHandler) HandleListFiles(c *gin.Context) {
	files, err := h.service.ListFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取得檔案列表失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

// HandleDeleteFile 刪除指定檔案
func (h *FileHandler) HandleDeleteFile(c *gin.Context) {
	fileName := c.Param("fileName")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未指定檔案名稱"})
		return
	}

	// 記錄嘗試刪除的檔案
	println("嘗試刪除檔案:", fileName)

	if err := h.service.DeleteFile(fileName); err != nil {
		println("刪除檔案失敗:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "刪除檔案失敗: " + err.Error()})
		return
	}

	println("檔案成功刪除:", fileName)

	c.JSON(http.StatusOK, gin.H{
		"message": "檔案已成功刪除",
	})
}

// HandleDownloadFile 下載指定檔案
func (h *FileHandler) HandleDownloadFile(c *gin.Context) {
	fileName := c.Param("fileName")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未指定檔案名稱"})
		return
	}

	filePath, err := h.service.GetFilePath(fileName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到檔案: " + err.Error()})
		return
	}

	// 設定下載檔案時使用的檔名
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Description", "File Transfer")
	c.File(filePath)
}

// HandleServeFile 提供檔案訪問（瀏覽而非下載）
func (h *FileHandler) HandleServeFile(c *gin.Context) {
	fileName := c.Param("fileName")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未指定檔案名稱"})
		return
	}

	// 記錄嘗試訪問的檔案
	println("嘗試檢視檔案:", fileName)

	filePath, err := h.service.GetFilePath(fileName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到檔案: " + err.Error()})
		return
	}

	// 記錄找到的檔案路徑
	println("檔案路徑:", filePath)

	// 根據檔案類型設定 Content-Type
	ext := filepath.Ext(fileName)

	contentType := getContentType(ext)
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}

	// 直接服務檔案
	c.File(filePath)
}

// getContentType 根據檔案擴展名取得對應的 MIME 類型
func getContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	case ".csv":
		return "text/csv"
	default:
		return "application/octet-stream" // 預設二進制流
	}
}

// HandleGetEmbeddingModels 獲取所有支援的嵌入模型
func (h *FileHandler) HandleGetEmbeddingModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"models": defaultEmbeddingModels,
	})
}

// HandleProcessFile 處理檔案的嵌入向量生成
func (h *FileHandler) HandleProcessFile(c *gin.Context) {
	fileName := c.Param("fileName")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未指定檔案名稱"})
		return
	}

	// 自定義嵌入模型
	customModel := c.Query("model")

	// 獲取檔案路徑
	filePath, err := h.service.GetFilePath(fileName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到檔案: " + err.Error()})
		return
	}

	// 獲取檔案類型
	ext := filepath.Ext(fileName)
	fileType := h.service.GetFileType(ext)

	// 選擇嵌入模型 (優先使用自定義模型)
	var embeddingModel EmbeddingModel
	if customModel != "" {
		embeddingModel = EmbeddingModel{
			Name:        customModel,
			Description: "自定義嵌入模型",
		}
	} else {
		embeddingModel = h.service.GetEmbeddingModel(fileType)
	}

	// 確認 OpenAI API Key
	openaiAPIKey := ""
	if h.config != nil && h.config.OpenAiAPIKey != "" {
		openaiAPIKey = h.config.OpenAiAPIKey
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "未設定 OpenAI API Key",
		})
		return
	}

	// 創建嵌入處理器
	processor := NewEmbeddingProcessor(openaiAPIKey)

	// 處理檔案並生成嵌入向量
	result, err := processor.ProcessFile(filePath, fileType, embeddingModel.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "處理檔案失敗: " + err.Error(),
		})
		return
	}

	// 返回處理結果
	c.JSON(http.StatusOK, gin.H{
		"fileName":       fileName,
		"filePath":       filePath,
		"fileType":       fileType,
		"embeddingModel": embeddingModel,
		"result":         result,
		"message":        "檔案處理完成，使用 " + embeddingModel.Name + " 模型",
	})
}
