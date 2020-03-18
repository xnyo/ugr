package admin

import (
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

func unauthorizedHandler(c *common.Ctx) {
	results := tb.Results{
		&tb.ArticleResult{
			Title: text.Unauthorized,
		},
	}
	results[0].SetResultID("unauthorized")
	results[0].SetContent(&tb.InputTextMessageContent{Text: "⛔️🙅‍♂️🍆"})
	err := c.B.Answer(c.InlineQuery, &tb.QueryResponse{Results: results})
	if err != nil {
		panic(err)
	}
}

// InlineInviteHandler handles incoming inline queries
func InlineInviteHandler(c *common.Ctx) {
	/* TODO
	if c.DbUser.Privileges&privileges.Admin == 0 {
		unauthorizedHandler(c)
		return
	}*/

	// Caller is admin
	results := tb.Results{
		&tb.ArticleResult{
			Title:       text.InvitePromptVolunteer,
			Description: text.InviteDescriptionVolunteer,
		},
		&tb.ArticleResult{
			Title:       text.InvitePromptAdmin,
			Description: text.InviteDescriptionAdmin,
		},
	}

	results[0].SetResultID("volunteer")
	results[0].SetContent(&tb.InputTextMessageContent{Text: text.InviteVolunteer(), ParseMode: "markdown"})
	results[0].SetReplyMarkup([][]tb.InlineButton{{{Text: text.InviteAccept, Unique: "accept_volunteer"}}})

	results[1].SetResultID("admin")
	results[1].SetContent(&tb.InputTextMessageContent{Text: text.InviteAdmin(), ParseMode: "markdown"})
	results[1].SetReplyMarkup([][]tb.InlineButton{{{Text: text.InviteAccept, Unique: "accept_admin"}}})
	err := c.B.Answer(c.InlineQuery, &tb.QueryResponse{
		Results: results,
	})
	if err != nil {
		panic(err)
	}
}
