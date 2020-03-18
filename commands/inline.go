package commands

import (
	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/privileges"
	tb "gopkg.in/tucnak/telebot.v2"
)

func handleInvite(c *common.Ctx, p privileges.Privileges) {
	if err := c.Db.Model(&c.DbUser).Update("privileges", gorm.Expr("privileges | ?", p)).Error; err != nil {
		c.RespondSessionError(err)
		return
	}
	c.B.Edit(c.Callback.Message, "👍 **Invito accettato!**", tb.ModeMarkdown, &tb.ReplyMarkup{})
}

// AcceptAdminInvite handles handle admin invites
func AcceptAdminInvite(c *common.Ctx) { handleInvite(c, privileges.Admin) }

// AcceptVolunteerInvite handles handle volunteer invites
func AcceptVolunteerInvite(c *common.Ctx) { handleInvite(c, privileges.Normal) }
