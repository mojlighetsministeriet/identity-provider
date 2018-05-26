package entity

import (
	"strings"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// Account represents an account that can be used to access the system
type Account struct {
	ID                 string   `json:"id" gorm:"not null;unique;size:36" validate:"uuid4,required"`
	Email              string   `json:"email" gorm:"not null;unique;size:100" validate:"email,required"`
	Roles              []string `json:"roles" gorm:"-"`
	RolesSerialized    string   `json:"-" gorm:"roles"`
	PasswordResetToken string   `json:"-"`
	Password           string   `json:"-"`
}

// AccountWithPassword represents an account but includes a seriaziable password property
type AccountWithPassword struct {
	Account
	Password string `json:"password"`
}

// GetID returns the ID for the account
func (account *Account) GetID() string {
	return account.ID
}

// GetEmail returns the email for the account
func (account *Account) GetEmail() string {
	return account.Email
}

// GetRolesSerialized returns the roles for the account serialized into a comma separated string
func (account *Account) GetRolesSerialized() string {
	return strings.Join(account.Roles, ",")
}

// BeforeSave will run before the struct is persisted with gorm
func (account *Account) BeforeSave() {
	if account.ID == "" {
		account.ID = uuid.Must(uuid.NewV4()).String()
	}

	account.RolesSerialized = account.GetRolesSerialized()
}

// AfterFind will run after the struct has been read from persistence
func (account *Account) AfterFind() {
	account.Roles = strings.Split(account.RolesSerialized, ",")
}

// SetPassword will update the accounts password
func (account *Account) SetPassword(password string) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err == nil {
		account.Password = string(hash)
	}

	return
}

// CompareHashedPasswordAgainst will compare a string with the accounts hashed password
func (account *Account) CompareHashedPasswordAgainst(passwordToCompareAgainst string) error {
	return bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(passwordToCompareAgainst))
}

// SetPasswordResetToken will update the password reset token
func (account *Account) SetPasswordResetToken(token string) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)

	if err == nil {
		account.PasswordResetToken = string(hash)
	}

	return
}

// CompareHashedPasswordResetTokenAgainst will compare a reset token string agains the accounts hashed reset token
func (account *Account) CompareHashedPasswordResetTokenAgainst(passwordResetTokenToCompareAgainst string) error {
	return bcrypt.CompareHashAndPassword([]byte(account.PasswordResetToken), []byte(passwordResetTokenToCompareAgainst))
}

// LoadAccountFromEmailAndPassword is used when authenticating to verify that email and password combination is valid
func LoadAccountFromEmailAndPassword(databaseConnection *gorm.DB, email string, password string) (account Account, err error) {
	account, err = LoadAccountFromEmail(databaseConnection, email)
	if err != nil {
		return
	}

	err = account.CompareHashedPasswordAgainst(password)
	if err != nil {
		account = Account{}
	}

	return
}

// LoadAccountFromEmail is used for example when resetting the password for an account
func LoadAccountFromEmail(databaseConnection *gorm.DB, email string) (account Account, err error) {
	err = databaseConnection.Where("email = ?", email).First(&account).Error
	return
}

// LoadAccountFromID will fetch the account from the persistence
func LoadAccountFromID(databaseConnection *gorm.DB, id string) (account Account, err error) {
	err = databaseConnection.Where("id = ?", id).First(&account).Error
	return
}
