package commands

import (
	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/privileges"
	tb "gopkg.in/tucnak/telebot.v2"
)

var unknownErrorCallbackResponse *tb.CallbackResponse = &tb.CallbackResponse{
	Text:      "Errore sconosciuto",
	ShowAlert: true,
}
var invalidInviteCallbackResponse *tb.CallbackResponse = &tb.CallbackResponse{
	Text:      "Invito non valido.",
	ShowAlert: true,
}

func handleInvite(c *common.Ctx, p privileges.Privileges) {
	// Operate in NoMenu mode, so we can use c.HandleErr()
	// as it will not update the menu (there's no menu!)
	c.NoMenu = true

	// Make sure the invite is valid
	err := c.Db.Transaction(func(tx *gorm.DB) error {
		var invite models.Invite
		var requiredType models.InviteType
		if p&privileges.Admin > 0 {
			requiredType = models.InviteTypeAdmin
		} else {
			requiredType = models.InviteTypeVolunteer
		}
		if err := tx.Where(&models.Invite{
			Token: c.Callback.Data,
			Type:  requiredType,
		}).First(&invite).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return common.ReportableError{T: "Invito non valido."}
			}
			return err
		}

		// Make sure the issuer is not dumb and clicking the button
		if invite.IssuedBy == c.Callback.Sender.ID {
			return common.ReportableError{T: "Non puoi accettare il tuo invito."}
		}

		// Register the user if necessary
		var new bool
		if c.DbUser == nil {
			// We use this instead of always adding
			// so we avoid duplicate primary key errors
			// in case of concurrency
			var user models.User
			if err := tx.Where(&models.User{
				TelegramID: c.TelegramUser().ID,
			}).Attrs(&models.User{
				TelegramID: c.TelegramUser().ID,
				Privileges: privileges.Normal,
			}).FirstOrCreate(&user).Error; err != nil {
				return err
			}
			c.DbUser = &user
			new = true
		}
		if !new || p != privileges.Normal {
			// Extra privileges needed, update privileges
			if err := tx.Model(&c.DbUser).Update("privileges", gorm.Expr("privileges | ?", p)).Error; err != nil {
				return err
			}
		}

		// Invalidate invite
		return tx.Unscoped().Delete(&invite).Error
	})
	if err != nil {
		c.HandleErr(err)
	} else {
		c.B.Edit(
			c.Callback.Message,
			"👍 **Invito accettato!**",
			tb.ModeMarkdown,
			&tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{{{Text: "🤖 Vai al bot", URL: "https://t.me/" + c.BotUsername}}},
			},
		)
	}

}

// AcceptAdminInvite handles handle admin invites
func AcceptAdminInvite(c *common.Ctx) { handleInvite(c, privileges.Admin) }

// AcceptVolunteerInvite handles handle volunteer invites
func AcceptVolunteerInvite(c *common.Ctx) { handleInvite(c, privileges.Normal) }
