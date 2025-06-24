package uploads

import (
	"ai-workshop/internal/config"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"encoding/csv"

	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
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

	// 真正提取文件內容
	var textContent string
	if fileType == FileTypeDocument && strings.ToLower(ext) == ".pdf" {
		// 真正解析 PDF 內容
		textContent, err = h.extractPDFContent(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "PDF 解析失敗: " + err.Error(),
			})
			return
		}
	} else if fileType == FileTypeText && strings.ToLower(ext) == ".csv" {
		// 真正解析 CSV 內容
		textContent, err = h.extractCSVContent(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "CSV 解析失敗: " + err.Error(),
			})
			return
		}
	} else if fileType == FileTypeDocument && (strings.ToLower(ext) == ".xlsx" || strings.ToLower(ext) == ".xls") {
		// 真正解析 Excel 內容
		textContent, err = h.extractExcelContent(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Excel 解析失敗: " + err.Error(),
			})
			return
		}
	} else {
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

		// 提取文本內容
		if resultMap, ok := result.(map[string]interface{}); ok {
			if content, exists := resultMap["content"]; exists {
				textContent = content.(string)
			} else if textPreview, exists := resultMap["textPreview"]; exists {
				textContent = textPreview.(string)
			}
		}
	}

	// 返回處理結果，包含真實的文本內容
	c.JSON(http.StatusOK, gin.H{
		"fileName":       fileName,
		"filePath":       filePath,
		"fileType":       fileType,
		"embeddingModel": embeddingModel,
		"result": map[string]interface{}{
			"content":     textContent,
			"contentSize": len(textContent),
			"docType":     ext,
		},
		"message": "檔案處理完成，使用 " + embeddingModel.Name + " 模型",
	})
}

// extractPDFContent 真正解析 PDF 內容
func (h *FileHandler) extractPDFContent(filePath string) (string, error) {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("無法打開 PDF 文件: %v", err)
	}
	defer file.Close()

	var textContent strings.Builder
	totalPages := reader.NumPage()

	// 限制最多處理前 10 頁，避免內容過長
	maxPages := totalPages
	if maxPages > 10 {
		maxPages = 10
	}

	for pageNum := 1; pageNum <= maxPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		// 提取頁面文本
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			// 如果某一頁解析失敗，繼續處理下一頁
			continue
		}

		textContent.WriteString(fmt.Sprintf("=== 第 %d 頁 ===\n", pageNum))
		textContent.WriteString(pageText)
		textContent.WriteString("\n\n")
	}

	extractedText := textContent.String()
	if len(strings.TrimSpace(extractedText)) == 0 {
		return "", fmt.Errorf("PDF 文件中沒有可提取的文本內容")
	}

	return extractedText, nil
}

// extractCSVContent 真正解析 CSV 內容
func (h *FileHandler) extractCSVContent(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("無法打開 CSV 文件: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return "", fmt.Errorf("讀取 CSV 內容失敗: %v", err)
	}

	if len(records) == 0 {
		return "", fmt.Errorf("CSV 文件中沒有數據")
	}

	var textContent strings.Builder

	// 處理標題行
	if len(records) > 0 {
		textContent.WriteString("=== CSV 數據內容 ===\n")
		textContent.WriteString("標題欄位: " + strings.Join(records[0], " | ") + "\n\n")
	}

	// 處理數據行，限制最多處理前 100 行避免內容過長
	maxRows := len(records)
	if maxRows > 100 {
		maxRows = 100
	}

	for i := 1; i < maxRows; i++ {
		textContent.WriteString(fmt.Sprintf("第 %d 行: %s\n", i, strings.Join(records[i], " | ")))
	}

	if len(records) > 100 {
		textContent.WriteString(fmt.Sprintf("\n... 共 %d 行數據，僅顯示前 100 行\n", len(records)-1))
	}

	extractedText := textContent.String()
	if len(strings.TrimSpace(extractedText)) == 0 {
		return "", fmt.Errorf("CSV 文件中沒有可提取的內容")
	}

	return extractedText, nil
}

// extractExcelContent 真正解析 Excel 內容
func (h *FileHandler) extractExcelContent(filePath string) (string, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return "", fmt.Errorf("無法打開 Excel 文件: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("關閉 Excel 文件時出錯: %v\n", err)
		}
	}()

	var textContent strings.Builder
	textContent.WriteString("=== Excel 數據內容 ===\n")

	// 獲取所有工作表名稱
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return "", fmt.Errorf("Excel 文件中沒有工作表")
	}

	// 限制最多處理前 3 個工作表
	maxSheets := len(sheets)
	if maxSheets > 3 {
		maxSheets = 3
	}

	for i := 0; i < maxSheets; i++ {
		sheetName := sheets[i]
		textContent.WriteString(fmt.Sprintf("\n=== 工作表: %s ===\n", sheetName))

		// 獲取工作表中的所有行
		rows, err := file.GetRows(sheetName)
		if err != nil {
			textContent.WriteString(fmt.Sprintf("讀取工作表 %s 失敗: %v\n", sheetName, err))
			continue
		}

		if len(rows) == 0 {
			textContent.WriteString("此工作表沒有數據\n")
			continue
		}

		// 處理標題行
		if len(rows) > 0 {
			textContent.WriteString("標題欄位: " + strings.Join(rows[0], " | ") + "\n")
		}

		// 處理數據行，限制最多處理前 50 行避免內容過長
		maxRows := len(rows)
		if maxRows > 50 {
			maxRows = 50
		}

		for rowIndex := 1; rowIndex < maxRows; rowIndex++ {
			if len(rows[rowIndex]) > 0 {
				textContent.WriteString(fmt.Sprintf("第 %d 行: %s\n", rowIndex, strings.Join(rows[rowIndex], " | ")))
			}
		}

		if len(rows) > 50 {
			textContent.WriteString(fmt.Sprintf("... 共 %d 行數據，僅顯示前 50 行\n", len(rows)-1))
		}
	}

	if len(sheets) > 3 {
		textContent.WriteString(fmt.Sprintf("\n... 共 %d 個工作表，僅顯示前 3 個\n", len(sheets)))
	}

	extractedText := textContent.String()
	if len(strings.TrimSpace(extractedText)) == 0 {
		return "", fmt.Errorf("Excel 文件中沒有可提取的內容")
	}

	return extractedText, nil
}
