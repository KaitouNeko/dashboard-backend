package clerkauth

import (
	"ai-workshop/internal/models"
	"ai-workshop/internal/user"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ClerkAuthHandler struct {
	service     ClerkService
	userService *user.UserService
}

func NewHandler(service ClerkService, userService *user.UserService) *ClerkAuthHandler {
	return &ClerkAuthHandler{
		service:     service,
		userService: userService,
	}
}

// 提取共用的 token 驗證邏輯
func (t *ClerkAuthHandler) extractAndVerifyToken(authHeader string) (*jwt.Token, error) {
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header missing")
	}

	parts := strings.Split(authHeader, "Bearer ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenStr := parts[1]

	return t.service.VerifyClerkToken(tokenStr)
}

// 從 JWT claims 中提取並補全 email 資訊
func (t *ClerkAuthHandler) extractUserInfo(claims jwt.MapClaims) (string, string, error) {
	userID := claims["sub"]
	email := claims["email"]

	userIDStr, ok := userID.(string)
	if !ok {
		return "", "", fmt.Errorf("invalid user ID in token")
	}

	// 如果 JWT 中沒有 email，從 Clerk API 取得
	if email == nil || email == "" {
		user, err := t.service.GetUser(userIDStr)
		if err == nil && len(user.EmailAddresses) > 0 {
			email = user.EmailAddresses[0].EmailAddress
		}
	}

	emailStr, ok := email.(string)
	if !ok || emailStr == "" {
		return userIDStr, "", fmt.Errorf("cannot get email for user")
	}

	return userIDStr, emailStr, nil
}

// 從 Clerk API 取得用戶的顯示名稱
func (t *ClerkAuthHandler) getClerkUserDisplayName(clerkID, fallbackEmail string) string {
	clerkUser, err := t.service.GetUser(clerkID)
	if err != nil {
		return fallbackEmail // 如果 API 呼叫失敗，使用 email 作為名稱
	}

	if clerkUser.FirstName != "" || clerkUser.LastName != "" {
		name := strings.TrimSpace(clerkUser.FirstName + " " + clerkUser.LastName)
		if name != "" {
			return name
		}
	}

	if clerkUser.Username != "" {
		return clerkUser.Username
	}

	return fallbackEmail
}

// 同步 Clerk 用戶到本地資料庫
func (t *ClerkAuthHandler) syncClerkUserToLocalDB(clerkID, email string) (*models.User, error) {
	name := t.getClerkUserDisplayName(clerkID, email)

	localUser, err := t.userService.CreateOrUpdateClerkUserService(clerkID, email, name)
	if err != nil {
		fmt.Printf("Warning: Failed to sync Clerk user to DB: %v\n", err)
		return nil, err
	}

	return localUser, nil
}

func (t *ClerkAuthHandler) VerifyToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	token, err := t.extractAndVerifyToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userIDStr, emailStr, err := t.extractUserInfo(claims)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token claims: " + err.Error()})
		return
	}

	// 同步用戶到本地 DB
	_, err = t.syncClerkUserToLocalDB(userIDStr, emailStr)
	if err != nil {
		fmt.Printf("Warning: User sync failed but continuing authentication: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":  true,
		"userID": userIDStr,
		"email":  emailStr,
	})
}

// VerifyTokenMiddleware 中間件，用於驗證請求中的 Clerk token
func (t *ClerkAuthHandler) VerifyTokenMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token, err := t.extractAndVerifyToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userIDStr, emailStr, err := t.extractUserInfo(claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid token claims: " + err.Error()})
			return
		}

		// 同步用戶到本地 DB
		localUser, err := t.syncClerkUserToLocalDB(userIDStr, emailStr)
		if err != nil {
			fmt.Printf("Warning: User sync failed but continuing authentication: %v\n", err)
		}

		// 存入 context，方便後續 handler 使用
		c.Set("userID", userIDStr)
		c.Set("email", emailStr)
		c.Set("valid", true)
		c.Set("claims", claims)
		if localUser != nil {
			c.Set("localUser", localUser)
		}
		c.Next()
	})
}

// SyncClerkUser 手動同步 Clerk 用戶到本地 DB
func (t *ClerkAuthHandler) SyncClerkUser(c *gin.Context) {
	var req struct {
		ClerkID string `json:"clerkId" binding:"required"`
		Email   string `json:"email" binding:"required"`
		Name    string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// 如果沒有提供 name，從 Clerk API 取得
	name := req.Name
	if name == "" {
		name = t.getClerkUserDisplayName(req.ClerkID, req.Email)
	}

	// 同步用戶到本地 DB
	localUser, err := t.userService.CreateOrUpdateClerkUserService(req.ClerkID, req.Email, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to sync user to local database",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User synchronized successfully",
		"user":    localUser,
	})
}
