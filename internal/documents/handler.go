// pkg/documents/handler.go
package documents

import (
	"ai-workshop/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(config *config.Config) *Handler {
	return &Handler{
		service: NewService(config),
	}
}

// ListCollections 列出所有集合
func (h *Handler) ListCollections(c *gin.Context) {
	collections, err := h.service.ListCollections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"collections": collections})
}

// CreateCollection 創建集合的 API
func (h *Handler) CreateCollection(c *gin.Context) {
	if err := h.service.CreateDocumentCollection(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "集合創建成功"})
}

// InsertDocument 插入文件的 API
func (h *Handler) InsertDocument(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 呼叫更新後的 service 方法，它會自動生成 ID
	id, err := h.service.InsertDocument(req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功訊息以及生成的 ID
	c.JSON(http.StatusOK, gin.H{
		"message": "文件插入成功",
		"id":      id,
	})
}

// ListVectors 列出文件的 API
func (h *Handler) ListVectors(c *gin.Context) {
	docs, err := h.service.ListVectors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, docs)
}

// DeleteDocument 刪除文件的 API
func (h *Handler) DeleteDocument(c *gin.Context) {
	var req struct {
		ID string `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式，ID 是必需的"})
		return
	}

	if err := h.service.DeleteDocument(req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件刪除成功"})
}

// DeleteDocuments 批量刪除文件的 API
func (h *Handler) DeleteDocuments(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式，IDs 是必需的"})
		return
	}

	if err := h.service.DeleteDocuments(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件批量刪除成功"})
}

// DeleteCollection 刪除整個集合的 API
func (h *Handler) DeleteCollection(c *gin.Context) {
	if err := h.service.DeleteCollection(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "集合刪除成功"})
}

// SearchDocuments 搜尋相似文檔的 API
func (h *Handler) SearchDocuments(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
		TopK  int    `json:"topK"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式，query 是必需的"})
		return
	}

	// 如果沒有指定 topK，則默認為 5
	if req.TopK <= 0 {
		req.TopK = 5
	}

	// 搜尋相似文檔
	docs, err := h.service.SearchSimilarDocuments(req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, docs)
}
