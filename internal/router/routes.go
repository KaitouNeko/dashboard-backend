package router

import (
	"ai-workshop/internal/auth"
	"ai-workshop/internal/chat"
	clerkauth "ai-workshop/internal/clerkAuth"
	"ai-workshop/internal/config"
	llmtype "ai-workshop/internal/constants"
	"ai-workshop/internal/documents"
	"ai-workshop/internal/energy"
	esgchat "ai-workshop/internal/esgChat"
	"ai-workshop/internal/uploads"
	"ai-workshop/internal/user"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// choose ai model
const (
	aiModel = llmtype.LLMTypeOpenAI // llm.LLMTypeGemini or llm.LLMTypeOpenAI
)

func SetupRoutes(config *config.Config, db *sqlx.DB) *gin.Engine {

	// 設置 Gin 模式
	gin.SetMode(gin.DebugMode)
	routes := gin.Default()

	// 配置 CORS

	// TODO: remove in production
	routes.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3333",
			"https://dashboard-frontend-git-develop-kaitounekos-projects.vercel.app",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // 加入 preflight 快取時間
	}))

	// 設置靜態文件服務
	routes.Static("/static", "./static")
	routes.StaticFile("/", "./static/index.html")
	routes.StaticFile("/uploads", "./static/uploads.html")

	// base route
	api := routes.Group("/api")
	api.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "v0.0.1",
		})
	})
	// --- Chat ---

	// -- setup --
	chatService, err := chat.NewService(config)
	if err != nil {
		fmt.Printf("error when initiating chat handler: %v\n", err)
	}
	chatHandler := chat.NewHandler(chatService, aiModel, config)

	// -- routes --
	api.POST("/chat", chatHandler.ChatHandler)
	api.POST("/rag", chatHandler.RagChatHandler)

	// Session management routes
	sessionRoutes := api.Group("/sessions")
	sessionRoutes.GET("", chatHandler.GetAllSessions)
	sessionRoutes.GET("/:sessionId", chatHandler.GetSessionInfo)
	sessionRoutes.DELETE("/:sessionId", chatHandler.DeleteSession)

	// --- ESG Chat ---
	esgRoutes := api.Group("/esg")
	esgChatService, err := esgchat.NewService(config)
	if err != nil {
		fmt.Printf("error when initiating chat handler: %v\n", err)
	}
	esgChatHandler := esgchat.NewHandler(esgChatService, config)

	esgRoutes.POST("/chat", esgChatHandler.ChatHandler)

	// --- Documents ---

	// -- setup --
	documentHandler := documents.NewHandler(config)

	// -- routes --
	// api.GET("/collections", documentHandler.ListCollections)
	api.POST("/collections/create", documentHandler.CreateCollection)
	api.DELETE("/collections", documentHandler.DeleteCollection)
	api.GET("/documents", documentHandler.ListVectors)
	api.POST("/documents/insert", documentHandler.InsertDocument)
	api.POST("/documents/delete", documentHandler.DeleteDocument)
	api.POST("/documents/delete/batch", documentHandler.DeleteDocuments)
	api.POST("/documents/search", documentHandler.SearchDocuments)

	// --- Energy (demo only) ---

	// -- setup --
	energyRepo := energy.NewRepository(db)
	energyService := energy.NewService(energyRepo)
	energyHandler := energy.NewHandler(energyService)

	// -- routes --
	energyRoutes := api.Group("/energy")
	energyRoutes.POST("", energyHandler.CreateEnergyUsage)
	energyRoutes.GET("", energyHandler.GetByTemperatureRange)
	energyRoutes.POST("/forecast", energyHandler.StoreForecast)
	energyRoutes.GET("/forecast", energyHandler.GetForecasts)

	// --- File Uploads ---

	// -- setup --
	uploadService := uploads.NewFileService(
		uploads.ServiceConfig{
			UploadDir:   "./uploads",
			MaxFileSize: 50 << 20, // 50MB
			AllowedFileTypes: []string{
				".pdf", ".doc", ".docx", ".txt", ".csv", ".xls", ".xlsx", ".json",
				".jpg", ".jpeg", ".png", ".gif", ".mp3", ".mp4", ".html",
			},
		},
	)
	uploadHandler := uploads.NewFileHandler(*uploadService, config)

	// -- routes --
	api.POST("/upload", uploadHandler.UploadFile)
	api.POST("/upload/multiple", uploadHandler.UploadFiles)
	api.GET("/list", uploadHandler.HandleListFiles)
	api.DELETE("/:fileName", uploadHandler.HandleDeleteFile)
	api.GET("/download/:fileName", uploadHandler.HandleDownloadFile)
	api.GET("/view/:fileName", uploadHandler.HandleServeFile)
	api.GET("/embedding-models", uploadHandler.HandleGetEmbeddingModels)
	api.POST("/process/:fileName", uploadHandler.HandleProcessFile)

	// --- USER ---

	// -- User Setup --
	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

	// -- User Routes --
	userRoutes := api.Group("/user")

	// Public Routes
	// userRoutes.GET("/:id", userHandler.GetUserByIdHandler)
	userRoutes.POST("/signin", userHandler.LoginUserHandler)

	// Protected Routes
	protectedUserRoutes := userRoutes.Group("")
	protectedUserRoutes.Use(auth.AuthMiddleware())
	protectedUserRoutes.GET("/users", userHandler.GetUsersHandler)
	protectedUserRoutes.POST("/signup", userHandler.CreateUserHandler)
	protectedUserRoutes.POST("/update-password", userHandler.UpdatePasswordUserHandler)
	protectedUserRoutes.POST("/update-info", userHandler.UpdateInfoUserHandler)

	// Clerk Authentication Routes
	clerkService := clerkauth.NewService()
	clerkHandler := clerkauth.NewHandler(clerkService)

	// 公開
	publicAuthRoutes := api.Group("/auth")
	{
		publicAuthRoutes.GET("/verify-token", clerkHandler.VerifyToken)
		publicAuthRoutes.POST("/logout", func(c *gin.Context) {
			// 登出邏輯 (前端處理 token 清除)
			c.JSON(http.StatusOK, gin.H{
				"message": "Successfully logged out",
				"success": true,
			})
		})
	}

	// 受保護
	protectedAuthRoutes := api.Group("/auth")
	protectedAuthRoutes.Use(clerkHandler.VerifyTokenMiddleware())
	{
		protectedAuthRoutes.GET("/user-profile", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			email, _ := c.Get("email")
			valid, _ := c.Get("valid")
			c.JSON(http.StatusOK, gin.H{
				"userID": userID,
				"email":  email,
				"valid":  valid,
			})
		})
	}

	return routes
}
