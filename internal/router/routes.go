package router

import (
	"ai-workshop/internal/auth"
	"ai-workshop/internal/chat"
	"ai-workshop/internal/config"
	"ai-workshop/internal/documents"
	"ai-workshop/internal/energy"
	"ai-workshop/internal/llm"
	"ai-workshop/internal/uploads"
	"ai-workshop/internal/user"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// choose ai model
const (
	aiModel = llm.LLMTypeOpenAI
)

func SetupRoutes(config *config.Config, db *sqlx.DB) *gin.Engine {

	// 設置 Gin 模式
	gin.SetMode(gin.DebugMode)
	routes := gin.Default()

	// 配置 CORS
	// TODO: remove in production
	routes.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3333"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 設置靜態文件服務
	routes.Static("/static", "./static")
	routes.StaticFile("/", "./static/index.html")
	routes.StaticFile("/uploads", "./static/uploads.html")

	// base route
	api := routes.Group("/api")

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

	return routes
}
