package account

import (
	"strings"

	"github.com/jinzhu/gorm"

	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID                 string   `json:"id" gorm:"not null;unique" validate:"uuid4,required"`
	Email              string   `json:"email" gorm:"not null;unique" validate:"email,required"`
	Roles              []string `json:"roles" gorm:"-"`
	RolesSerialized    string   `gorm:"roles"`
	PasswordResetToken string   `json:"-"`
	Password           string   `json:"-"`
}

type AccountWithPassword struct {
	Account
	Password string `json:"password"`
}

func (account *Account) BeforeSave() {
	account.RolesSerialized = strings.Join(account.Roles, ",")
}

func (account *Account) AfterFind() {
	account.Roles = strings.Split(account.RolesSerialized, ",")
}

func (account *Account) SetPassword(password string) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err == nil {
		account.Password = string(hash)
	}

	return
}

func (account *Account) CompareHashedPasswordAgainst(passwordToCompareAgainst string) error {
	return bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(passwordToCompareAgainst))
}

func (account *Account) SetPasswordResetToken(token string) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)

	if err == nil {
		account.PasswordResetToken = string(hash)
	}

	return
}

func (account *Account) CompareHashedPasswordResetTokenAgainst(PasswordResetTokenToCompareAgainst string) error {
	return bcrypt.CompareHashAndPassword([]byte(account.PasswordResetToken), []byte(PasswordResetTokenToCompareAgainst))
}

func LoadAccountFromEmailAndPassword(databaseConnection *gorm.DB, email string, password string) (account Account, err error) {
	err = databaseConnection.Where("email = ?", email).First(&account).Error
	if err != nil {
		return
	}

	err = account.CompareHashedPasswordAgainst(password)
	if err != nil {
		account = Account{}
	}

	return
}
