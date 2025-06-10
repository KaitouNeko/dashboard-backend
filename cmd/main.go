package main

import (
	"ai-workshop/internal/config"
	"ai-workshop/internal/db"
	"ai-workshop/internal/router"
	"ai-workshop/pkg/util"
	"log"
	"os"
)

func main() {
	appConfig := config.NewConfig()
	defer appConfig.CleanUp()

	// setup database
	sqlDB, err := db.NewPostgresDB(
		appConfig.PostgresHost,
		appConfig.PostgresPort,
		appConfig.PostgresUser,
		appConfig.PostgresPassword,
		appConfig.PostgresDBName,
	)

	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	defer sqlDB.Close()

	// migrations
	db.RunMigrations(sqlDB.DB, "./internal/db/migrations/")

	// base gin and routes setup
	r := router.SetupRoutes(appConfig, sqlDB)

	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	port := util.GetEnvString("PORT", "8080")

	log.Printf("Server running on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("啟動服務器失敗: %v", err)
	}
}
