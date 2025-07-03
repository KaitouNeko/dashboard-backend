package clerkauth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ClerkAuthHandler struct {
	service ClerkService
}

func NewHandler(service ClerkService) *ClerkAuthHandler {
	return &ClerkAuthHandler{service: service}
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

func (t *ClerkAuthHandler) VerifyToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	token, err := t.extractAndVerifyToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"]
	email := claims["email"]

	// 如果 JWT 中沒有 email，從 Clerk API 取得
	if email == nil || email == "" {
		if userIDStr, ok := userID.(string); ok {
			user, err := t.service.GetUser(userIDStr)
			if err == nil && len(user.EmailAddresses) > 0 {
				email = user.EmailAddresses[0].EmailAddress
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":  true,
		"userID": userID,
		"email":  email,
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

		// 取得 user_id 和 email
		userID := claims["sub"]
		email := claims["email"]

		// 如果 JWT 中沒有 email，從 Clerk API 取得
		if email == nil || email == "" {
			if userIDStr, ok := userID.(string); ok {
				user, err := t.service.GetUser(userIDStr)
				if err == nil && len(user.EmailAddresses) > 0 {
					email = user.EmailAddresses[0].EmailAddress
				}
			}
		}

		// 存入 context，方便後續 handler 使用
		c.Set("userID", userID)
		c.Set("email", email)
		c.Set("valid", true)
		c.Set("claims", claims)
		c.Next()
	})
}
