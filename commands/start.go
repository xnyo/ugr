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
	if c.TelegramUser().FirstName != "" {
		s += fmt.Sprintf(", <b>%s</b>!", html.EscapeString(c.TelegramUser().FirstName))
	}
	c.SetState("start")
	c.ClearStateData()
	keyboard := [][]tb.InlineButton{
		{{Unique: "user__take_order", Text: "🛒 Scegli ordine"}},
		{{Unique: "user__my_orders", Text: "📑 I miei ordini"}},
	}
	if c.DbUser.Privileges&privileges.Admin > 0 {
		keyboard = append(keyboard, []tb.InlineButton{{Unique: "admin", Text: "🔧 Pannello admin"}})
	}
	keyboard = append(keyboard, []tb.InlineButton{{Unique: "delete_self", Text: "❌ Chiudi"}})
	c.UpdateMenu(
		s,
		&tb.ReplyMarkup{
			InlineKeyboard: keyboard,
		},
		tb.ModeHTML,
	)
}
