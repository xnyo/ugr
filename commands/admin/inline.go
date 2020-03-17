package admin

import (
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/privileges"
	tb "gopkg.in/tucnak/telebot.v2"
)

func unauthorizedHandler(c *common.Ctx) {
	results := tb.Results{
		&tb.ArticleResult{
			Title: "⛔️ Non puoi usare questa funzionalità",
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
	if c.DbUser.Privileges&privileges.Admin == 0 {
		unauthorizedHandler(c)
		return
	}

	// Caller is admin
	results := tb.Results{
		&tb.ArticleResult{
			Title:       "🛵 Invita come volontario",
			Description: "L'utente verrà invitato come volontario",
		},
		&tb.ArticleResult{
			Title:       "🔧 Invita come amministratore",
			Description: "L'utente verrà invitato come amministratore",
		},
	}

	results[0].SetResultID("volunteer")
	results[0].SetContent(&tb.InputTextMessageContent{
		Text:      "👋 **Ciao**, sei stato invitato come __volontario__.\n\n👇 Fai click qui per accettare! 👇",
		ParseMode: "markdown",
	})
	results[0].SetReplyMarkup([][]tb.InlineButton{{{Text: "✅ Accetta", Unique: "dummy"}}})

	results[1].SetResultID("admin")
	results[1].SetContent(&tb.InputTextMessageContent{
		Text:      "👋 **Ciao**, sei stato invitato come __amministratore__.\n\n👇 Fai click qui per accettare! 👇",
		ParseMode: "markdown",
	})
	results[1].SetReplyMarkup([][]tb.InlineButton{{{Text: "✅ Accetta", Unique: "dummy"}}})
	err := c.B.Answer(c.InlineQuery, &tb.QueryResponse{
		Results: results,
	})
	if err != nil {
		panic(err)
	}
}
