package user

import (
	"ai-workshop/internal/models"
	"ai-workshop/internal/utils/errorutils"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	DB *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) Create(user models.User) error {
	query := `INSERT INTO users (name, email, password, permission) VALUES (:name, :email, :password, :permission)`

	_, err := r.DB.NamedExec(query, user)

	if err != nil {
		fmt.Println("Error when creating user:", err)
		return errorutils.AnalyzeDBErr(err)
	}

	return nil
}

func (r *UserRepository) UpdatePassword(params UserUpdatePasswordParams) error {
	query := `UPDATE users SET password = :password WHERE id = :id`

	result, err := r.DB.NamedExec(query, params)
	if err != nil {
		return errorutils.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no member found with id: %v", params.ID)
	}

	return nil
}

func (r *UserRepository) UpdateInfo(params UserUpdateInfoParams, userId uuid.UUID) error {
	query := `UPDATE users SET name = :name, permission = :permission WHERE id = :id`

	result, err := r.DB.NamedExec(query, params)

	fmt.Println("result", result)
	if err != nil {
		return errorutils.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id: %v", params.ID)
	}

	return nil
}

func (r *UserRepository) GetByIdWithPassword(id uuid.UUID) (*models.User, error) {
	query := `SELECT * FROM users WHERE users.id = $1`

	var user models.User

	err := r.DB.Get(&user, query, id)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUsers() (*[]models.User, error) {
	query := `SELECT id, email, name, updated_at, permission, status FROM users`

	var users []models.User

	err := r.DB.Select(&users, query)

	if err != nil {
		return nil, err
	}

	return &users, nil
}

func (r *UserRepository) GetById(id uuid.UUID) (*models.User, error) {
	query := `SELECT * FROM users WHERE users.id = $1`

	var user models.User

	err := r.DB.Get(&user, query, id)

	if err != nil {
		return nil, err
	}

	// Remove password from the struct
	user.Password = ""

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE users.email = $1`

	err := r.DB.Get(&user, query, email)
	fmt.Println("Error:", err)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByClerkID(clerkID string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE clerk_id = $1`

	err := r.DB.Get(&user, query, clerkID)
	if err != nil {
		return nil, err
	}

	// Remove password from the struct
	user.Password = ""
	return &user, nil
}

func (r *UserRepository) CreateOrUpdateClerkUser(clerkID, email, name string) (*models.User, error) {
	var user models.User

	// 嘗試更新已存在的用戶
	updateQuery := `
		UPDATE users 
		SET clerk_id = $1, name = $2, updated_at = CURRENT_TIMESTAMP
		WHERE email = $3 AND clerk_id IS NULL
		RETURNING *
	`

	err := r.DB.Get(&user, updateQuery, clerkID, name, email)
	if err == nil {
		// 更新成功
		user.Password = ""
		return &user, nil
	}

	// 如果更新失敗，檢查是否已經有 clerk_id 的用戶
	existingQuery := `SELECT * FROM users WHERE clerk_id = $1`
	err = r.DB.Get(&user, existingQuery, clerkID)
	if err == nil {
		// 已經存在，更新名稱
		updateNameQuery := `
			UPDATE users 
			SET name = $1, updated_at = CURRENT_TIMESTAMP
			WHERE clerk_id = $2
			RETURNING *
		`
		err = r.DB.Get(&user, updateNameQuery, name, clerkID)
		if err == nil {
			user.Password = ""
			return &user, nil
		}
	}

	// 創建新用戶
	insertQuery := `
		INSERT INTO users (email, name, clerk_id, password, status, permission)
		VALUES ($1, $2, $3, '', 1, 1)
		RETURNING *
	`

	err = r.DB.Get(&user, insertQuery, email, name, clerkID)
	if err != nil {
		return nil, fmt.Errorf("failed to create clerk user: %w", err)
	}

	user.Password = ""
	return &user, nil
}

func (r *UserRepository) CreateDefaultUsers(users []CreateDefaultUser) error {
	query := `
	INSERT INTO users(id, email, name, password, status)
	VALUES(:id, :email, :name, :password, :status)
	ON CONFLICT (id) DO NOTHING
	`
	_, err := r.DB.NamedExec(query, users)

	fmt.Printf("DEBUG: Error when creating default user: %s\n", err)

	if err != nil {
		return errorutils.AnalyzeDBErr(err)
	}

	return nil
}
