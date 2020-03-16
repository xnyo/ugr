package models

import (
	"github.com/xnyo/ugr/privileges"
)

// User represents a bot user
type User struct {
	TelegramID int `gorm:"primary_key"`
	Privileges privileges.Privileges
	State      string
}

// TableName returns the sql table name
func (User) TableName() string {
	return "users"
}
