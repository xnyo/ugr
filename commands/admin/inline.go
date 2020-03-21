package admin

import (
	"fmt"

	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/privileges"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

func unauthorizedHandler(c *common.Ctx) {
	results := tb.Results{&tb.ArticleResult{Title: text.Unauthorized}}
	results[0].SetContent(&tb.InputTextMessageContent{Text: "⛔️🙅‍♂️🍆"})
	err := c.AnswerNoCache(&tb.QueryResponse{Results: results})
	// We panic because we already (tried to) reply to the query
	if err != nil {
		panic(err)
	}
}

// InlineInviteHandler handles incoming inline queries
func InlineInviteHandler(c *common.Ctx) {
	if c.DbUser == nil || c.DbUser.Privileges&privileges.Admin == 0 {
		unauthorizedHandler(c)
		return
	}

	// Determine invite type
	var inviteType models.InviteType
	if c.InlineQuery.Text == "volunteer" {
		inviteType = models.InviteTypeVolunteer
	} else if c.InlineQuery.Text == "admin" {
		inviteType = models.InviteTypeAdmin
	} else {
		// Unknown query
		err := c.AnswerNoCache(&tb.QueryResponse{Results: tb.Results{}})
		// We panic because we already (tried to) reply to the query
		if err != nil {
			panic(err)
		}
	}

	// Generate a new invite, with a valid token
	invite, err := models.NewInvite(c.Db, inviteType, c.DbUser.TelegramID)
	if err != nil {
		c.HandleErr(err)
		return
	}

	// Save invite in db
	if err := c.Db.Save(&invite).Error; err != nil {
		c.HandleErr(err)
		return
	}

	var results tb.Results
	if inviteType == models.InviteTypeVolunteer {
		results = tb.Results{
			&tb.ArticleResult{
				Title:       text.InvitePromptVolunteer,
				Description: text.InviteDescriptionVolunteer,
			},
		}
		q := fmt.Sprintf("accept_volunteer|%s", invite.Token)
		results[0].SetContent(&tb.InputTextMessageContent{Text: text.InviteVolunteer(), ParseMode: "markdown"})
		results[0].SetReplyMarkup([][]tb.InlineButton{{{Text: text.InviteAccept, Unique: q}}})
	} else if inviteType == models.InviteTypeAdmin {
		results = tb.Results{
			&tb.ArticleResult{
				Title:       text.InvitePromptAdmin,
				Description: text.InviteDescriptionAdmin,
			},
		}
		q := fmt.Sprintf("accept_admin|%s", invite.Token)
		results[0].SetContent(&tb.InputTextMessageContent{Text: text.InviteAdmin(), ParseMode: "markdown"})
		results[0].SetReplyMarkup([][]tb.InlineButton{{{Text: text.InviteAccept, Unique: q}}})
	}
	err = c.AnswerNoCache(&tb.QueryResponse{Results: results})
	// We panic because we already (tried to) reply to the query
	if err != nil {
		panic(err)
	}
}
