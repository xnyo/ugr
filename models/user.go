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
	StateData      string `gorm:"size:512" gorm:"default:'{}'"`
	LatestBotMsgID int
	AssignedOrders []Order `gorm:"foreignkey:AssignedUserID"`
}

// MessageSig makes User a struct that implements tb.Editable
func (x User) MessageSig() (string, int64) {
	return strconv.Itoa(x.LatestBotMsgID), int64(x.TelegramID)
}
