package models

import (
	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/utils"
)

// InviteType represents the type of an invite
type InviteType int

// Invite type values
const (
	InviteTypeVolunteer InviteType = iota
	InviteTypeAdmin
)

// Invite represents an invite request
type Invite struct {
	gorm.Model
	Type     InviteType
	Token    string `gorm:"type:varchar(16);unique_index"`
	IssuedBy int
}

// NewInvite generates a new invite, with a valid unique token
func NewInvite(db *gorm.DB, inviteType InviteType, issuedBy int) (*Invite, error) {
	var ok, invite Invite
	invite.Type = inviteType
	invite.IssuedBy = issuedBy
	for {
		invite.Token = utils.RandomString(16)
		if err := db.Where(&invite).First(&ok).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return &invite, nil
			}
			return nil, err
		}
	}
}
