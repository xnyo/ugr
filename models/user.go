package models

import (
	"strconv"

	"github.com/xnyo/ugr/privileges"
)

// User represents a bot user
type User struct {
	TelegramID     int `gorm:"primary_key"`
	Privileges     privileges.Privileges
	State          string
	LatestBotMsgID int
}

// TableName returns the sql table name
func (User) TableName() string {
	return "users"
}

// MessageSig makes User a struct that implements tb.Editable
func (x User) MessageSig() (string, int64) {
	return strconv.Itoa(x.LatestBotMsgID), int64(x.TelegramID)
}
