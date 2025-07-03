package user

import (
	"ai-workshop/internal/models"

	"github.com/google/uuid"
)

type UserResponse struct {
	models.BaseDBDateModel
	Email string `db:"email" json:"email"`
	Name  string `db:"name" json:"name"`
}

type UserLoginRequest struct {
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"password"`
}

type UserLoginResponse struct {
	RefreshToken     string `json:"refreshToken"`
	AccessToken      string `json:"accessToken"`
	AccessExpiresIn  int    `json:"accessExpiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`

	UserInfo *models.User `json:"userInfo"`
}

type UserUpdatePasswordRequest struct {
	Password          string `db:"password" json:"password"`
	NewPassword       string `json:"newPassword"`
	RepeatNewPassword string `json:"repeatNewPassword"`
}

type UserUpdatePasswordParams struct {
	ID       uuid.UUID `db:"id" json:"id"`
	Password string    `db:"password" json:"password"`
}

type UserUpdateInfoRequest struct {
	Name string `db:"name" json:"name"`
	// Status     string `db:"status" json:"status"`
	Permission int `db:"permission" json:"permission"`
}

type UserUpdateInfoParams struct {
	ID   uuid.UUID `db:"id" json:"id"`
	Name string    `db:"name" json:"name"`
	// Status     string    `db:"status" json:"status"`
	Permission int `db:"permission" json:"permission"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type CreateDefaultUser struct {
	ID         uuid.UUID `db:"id"`
	Email      string    `db:"email" `
	Name       string    `db:"name" `
	Password   string    `db:"password" `
	Status     int       `db:"status"`
	Permission int       `db:"permission"`
}

type ClerkUserSyncRequest struct {
	ClerkID string `json:"clerkId"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}

type ClerkUserSyncResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	User    *models.User `json:"user,omitempty"`
}
