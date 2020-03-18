package commands

import (
	"fmt"
	"html"

	"github.com/xnyo/ugr/privileges"

	"github.com/xnyo/ugr/common"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Start handles the /start command
func Start(c *common.Ctx) {
	s := "👋 Benvenuto"
	if c.Message.Sender.FirstName != "" {
		s += fmt.Sprintf(", <b>%s</b>!", html.EscapeString(c.Message.Sender.FirstName))
	}
	if c.DbUser.Privileges&privileges.Admin > 0 {
		s += "\n\n<i>Puoi accedere al pannello amministratore con il comando /admin</i>"
	}
	c.SetState("start")
	c.ClearStateData()
	c.UpdateMenu(
		s,
		&tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					{
						Unique: "user__take_order",
						Text:   "🛒 Scegli ordine",
					},
				},
				{
					{
						Unique: "user__my_orders",
						Text:   "📑 I miei ordini",
					},
				},
				{
					{
						Unique: "delete_self",
						Text:   "❌ Chiudi",
					},
				},
			},
		},
		tb.ModeHTML,
	)
}
