package admin

import (
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/privileges"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

func unauthorizedHandler(c *common.Ctx) {
	results := tb.Results{
		&tb.ArticleResult{
			Title: text.Unauthorized,
		},
	}
	results[0].SetResultID("unauthorized_" + string(c.TelegramUser().ID))
	results[0].SetContent(&tb.InputTextMessageContent{Text: "⛔️🙅‍♂️🍆"})
	err := c.B.Answer(c.InlineQuery, &tb.QueryResponse{Results: results})
	if err != nil {
		panic(err)
	}
}

// InlineInviteHandler handles incoming inline queries
func InlineInviteHandler(c *common.Ctx) {
	if c.DbUser.Privileges&privileges.Admin == 0 {
		unauthorizedHandler(c)
		return
	}
	var results tb.Results
	if c.InlineQuery.Text == "volunteer" {
		results = tb.Results{
			&tb.ArticleResult{
				Title:       text.InvitePromptVolunteer,
				Description: text.InviteDescriptionVolunteer,
			},
		}
		results[0].SetResultID("volunteer")
		results[0].SetContent(&tb.InputTextMessageContent{Text: text.InviteVolunteer(), ParseMode: "markdown"})
		results[0].SetReplyMarkup([][]tb.InlineButton{{{Text: text.InviteAccept, Unique: "accept_volunteer"}}})
	} else if c.InlineQuery.Text == "admin" {
		results = tb.Results{
			&tb.ArticleResult{
				Title:       text.InvitePromptAdmin,
				Description: text.InviteDescriptionAdmin,
			},
		}
		results[0].SetResultID("admin")
		results[0].SetContent(&tb.InputTextMessageContent{Text: text.InviteAdmin(), ParseMode: "markdown"})
		results[0].SetReplyMarkup([][]tb.InlineButton{{{Text: text.InviteAccept, Unique: "accept_admin"}}})
	} else {
		// Unknown query
		c.B.Answer(c.InlineQuery, &tb.QueryResponse{Results: tb.Results{}})
	}

	err := c.B.Answer(c.InlineQuery, &tb.QueryResponse{
		Results: results,
	})
	if err != nil {
		panic(err)
	}
}
