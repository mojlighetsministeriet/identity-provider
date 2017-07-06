package users

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// User represents a user
type User struct {
	gorm.Model
	ID       uuid.UUID `json:"id" gorm:"primary_key"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
	Roles    []string  `json:"roles"`
}
