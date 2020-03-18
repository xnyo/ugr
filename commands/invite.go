package commands

import (
	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/privileges"
	tb "gopkg.in/tucnak/telebot.v2"
)

func handleInvite(c *common.Ctx, p privileges.Privileges) {
	// Register the user if necessary
	var new bool
	if c.DbUser == nil {
		// We use this instead of always adding
		// so we avoid duplicate primary key errors
		// in case of concurrency
		var user models.User
		if err := c.Db.Model(&models.User{
			TelegramID: c.TelegramUser().ID,
		}).Attrs(&models.User{
			TelegramID: c.TelegramUser().ID,
			Privileges: privileges.Normal,
		}).FirstOrCreate(&user).Error; err != nil {
			c.RespondSessionError(err)
			return
		}
		c.DbUser = &user
		new = true
	}

	if !new || p != privileges.Normal {
		// Extra privileges needed, update privileges
		if err := c.Db.Model(&c.DbUser).Update("privileges", gorm.Expr("privileges | ?", p)).Error; err != nil {
			c.RespondSessionError(err)
			return
		}
	}

	// Normal privileges, no need to update privileges on new records
	c.B.Edit(c.Callback.Message, "👍 **Invito accettato!**", tb.ModeMarkdown, &tb.ReplyMarkup{})
}

// AcceptAdminInvite handles handle admin invites
func AcceptAdminInvite(c *common.Ctx) { handleInvite(c, privileges.Admin) }

// AcceptVolunteerInvite handles handle volunteer invites
func AcceptVolunteerInvite(c *common.Ctx) { handleInvite(c, privileges.Normal) }
