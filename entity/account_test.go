package entity_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mojlighetsministeriet/identity-provider/entity"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccountGetID(test *testing.T) {
	account := entity.Account{ID: uuid.Must(uuid.NewV4()).String()}
	assert.Equal(test, account.ID, account.GetID())
}

func TestAccountGetEmail(test *testing.T) {
	account := entity.Account{Email: "test@example.com"}
	assert.Equal(test, account.Email, account.GetEmail())
}

func TestAccountBeforeSave(test *testing.T) {
	account := entity.Account{Roles: []string{"administrator", "user"}}
	assert.Equal(test, 0, len(account.ID))
	account.BeforeSave()
	assert.Equal(test, 36, len(account.ID))
	assert.Equal(test, "administrator,user", account.RolesSerialized)
}

func TestAccountAfterFind(test *testing.T) {
	account := entity.Account{RolesSerialized: "administrator,user"}
	account.AfterFind()
	assert.Equal(test, []string{"administrator", "user"}, account.Roles)
}

func TestAccountCompareHashedPasswordAgainst(test *testing.T) {
	password := "mysecretpassword"
	account := entity.Account{}
	account.SetPassword(password)
	err := account.CompareHashedPasswordAgainst(password)
	assert.NoError(test, err)
}

func TestAccountCompareHashedPasswordResetTokenAgainst(test *testing.T) {
	token := "mysecrettoken"
	account := entity.Account{}
	account.SetPasswordResetToken(token)
	err := account.CompareHashedPasswordResetTokenAgainst(token)
	assert.NoError(test, err)
}

func TestAccountLoadAccountFromID(test *testing.T) {
	databaseConnection, err := gorm.Open("sqlite3", "/tmp/identity-provider-test-"+uuid.Must(uuid.NewV4()).String()+".db")
	assert.NoError(test, err)
	defer databaseConnection.Close()

	err = databaseConnection.AutoMigrate(&entity.Account{}).Error
	assert.NoError(test, err)

	account := entity.Account{ID: uuid.Must(uuid.NewV4()).String(), Email: "user@example.com"}
	err = databaseConnection.Create(&account).Error
	assert.NoError(test, err)

	loadedAccount, err := entity.LoadAccountFromID(databaseConnection, account.ID)
	assert.NoError(test, err)
	assert.Equal(test, "user@example.com", loadedAccount.Email)
	assert.Equal(test, account.ID, loadedAccount.ID)
}

func TestAccountLoadAccountFromEmailAndPassword(test *testing.T) {
	databaseConnection, err := gorm.Open("sqlite3", "/tmp/identity-provider-test-"+uuid.Must(uuid.NewV4()).String()+".db")
	assert.NoError(test, err)
	defer databaseConnection.Close()

	err = databaseConnection.AutoMigrate(&entity.Account{}).Error
	assert.NoError(test, err)

	account := entity.Account{ID: uuid.Must(uuid.NewV4()).String(), Email: "user@example.com"}
	account.SetPassword("mysecretpassword")
	err = databaseConnection.Create(&account).Error
	assert.NoError(test, err)

	loadedAccount, err := entity.LoadAccountFromEmailAndPassword(databaseConnection, account.Email, "mysecretpassword")
	assert.NoError(test, err)
	assert.Equal(test, "user@example.com", loadedAccount.Email)
	assert.Equal(test, account.ID, loadedAccount.ID)
}

func TestAccountLoadAccountFromEmailAndPasswordWithInvalidPassword(test *testing.T) {
	databaseConnection, err := gorm.Open("sqlite3", "/tmp/identity-provider-test-"+uuid.Must(uuid.NewV4()).String()+".db")
	assert.NoError(test, err)
	defer databaseConnection.Close()

	err = databaseConnection.AutoMigrate(&entity.Account{}).Error
	assert.NoError(test, err)

	account := entity.Account{ID: uuid.Must(uuid.NewV4()).String(), Email: "user@example.com"}
	account.SetPassword("mysecretpassword")
	err = databaseConnection.Create(&account).Error
	assert.NoError(test, err)

	loadedAccount, err := entity.LoadAccountFromEmailAndPassword(databaseConnection, account.Email, "invalidpassword")
	assert.Error(test, err)
	assert.Equal(test, "", loadedAccount.Email)
}

func TestAccountLoadAccountFromEmailAndPasswordWithInvalidEmail(test *testing.T) {
	databaseConnection, err := gorm.Open("sqlite3", "/tmp/identity-provider-test-"+uuid.Must(uuid.NewV4()).String()+".db")
	assert.NoError(test, err)
	defer databaseConnection.Close()

	err = databaseConnection.AutoMigrate(&entity.Account{}).Error
	assert.NoError(test, err)

	account := entity.Account{ID: uuid.Must(uuid.NewV4()).String(), Email: "user@example.com"}
	account.SetPassword("mysecretpassword")
	err = databaseConnection.Create(&account).Error
	assert.NoError(test, err)

	loadedAccount, err := entity.LoadAccountFromEmailAndPassword(databaseConnection, "invalidemail@example.com", "mysecretpassword")
	assert.Error(test, err)
	assert.Equal(test, "", loadedAccount.Email)
}
