package models

import (
	"time"

	"github.com/google/uuid"
)

/**
* Types here are shared model entities that are imported by more than one package.
**/

/**
* User
**/
type User struct {
	BaseDBDateModel
	Email      string  `db:"email" json:"email"`
	Name       string  `db:"name" json:"name"`
	Password   string  `db:"password" json:"password,omitempty"`
	Status     int     `db:"status" json:"status"`
	Permission int     `db:"permission" json:"permission"`
	ClerkID    *string `db:"clerk_id" json:"clerkId,omitempty"`
}

/**
* Base models for default table columns.
**/

type BaseIDModel struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type BaseDBUserModel struct {
	ID          uuid.UUID `db:"id" json:"id"`
	UpdatedUser uuid.UUID `db:"updated_user" json:"updatedUser"`
	CreatedUser uuid.UUID `db:"created_user" json:"createdUser"`
}

type BaseDBDateModel struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type BaseDBUserDateModel struct {
	ID          uuid.UUID `db:"id" json:"id"`
	UpdatedUser uuid.UUID `db:"updated_user" json:"updatedUser"`
	CreatedUser uuid.UUID `db:"created_user" json:"createdUser"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}
