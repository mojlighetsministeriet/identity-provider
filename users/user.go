package users

import (
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

// User represents a user
type User struct {
	ID          uuid.UUID  `sql:"id;not null;type:varchar(36);primary_key" validate:"uuid4"`
	CreatedAt   time.Time  `sql:"not null"`
	UpdatedAt   time.Time  `validate:"gtefield=CreatedAt"`
	DeletedAt   *time.Time `sql:"index" validate:"gtefield=CreatedAt"`
	Email       string     `sql:"not null" json:"email" validate:"email"`
	Password    string     `json:"-"`
	Roles       []string   `json:"roles" sql:"-" validate:"dive,eq=user|eq=admin|eq=super-admin"`
	RolesString string     `json:"-" sql:"roles;not null"`
}

// AfterFind splits the Roles field after loaded
func (user *User) AfterFind() {
	user.Roles = strings.Split(user.RolesString, ";")
}

// BeforeSave populates the rolesString field before saving to database
func (user *User) BeforeSave() {
	user.RolesString = strings.Join(user.Roles, ";")
}
