package db

import (
	"ai-workshop/internal/user"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresDB(host, port, user, password, dbname string) (*sqlx.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	SeedDefaults(db)

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}

func SeedDefaults(db *sqlx.DB) {
	// --- default users ---
	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo)

	user := []user.CreateDefaultUser{{
		ID:         uuid.MustParse("6f60f94a-6c90-45a1-96f6-32174cc0f908"),
		Email:      "god_admin@gmail.com",
		Name:       "God Admin",
		Password:   "QWE@asd123",
		Status:     2,
		Permission: 1},
	}
	err := userService.CreateDefaultUsersService(user)

	if err != nil {
		log.Fatal("Error when attempting to create default users:", err)
	}

	fmt.Printf("Successfully created all default users.\n\n")

}
