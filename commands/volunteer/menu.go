package volunteer

import (
	"fmt"
	"html"

	"github.com/xnyo/ugr/text"

	"github.com/xnyo/ugr/privileges"

	"github.com/xnyo/ugr/common"
	tb "gopkg.in/tucnak/telebot.v2"
)

// BackReplyButton is the "back" button goes back to the
// volunteer panel main menu
var BackReplyButton tb.InlineButton = tb.InlineButton{
	Unique: "volunteer",
	Text:   text.MainMenu,
}

// BackReplyKeyboard is a [][]tb.InlineButton with a
// single "Main menu" button that goes to
// the volunteer panel main menu
var BackReplyKeyboard [][]tb.InlineButton = [][]tb.InlineButton{{BackReplyButton}}

// BackReplyMarkup is a tb.ReplyMarkup with a
// single "Main menu" button that goes to
// the volunteer panel main menu
var BackReplyMarkup *tb.ReplyMarkup = &tb.ReplyMarkup{InlineKeyboard: BackReplyKeyboard}

// Menu handles the /start command
func Menu(c *common.Ctx) {
	s := "👋 Benvenuto"
	if c.TelegramUser().FirstName != "" {
		s += fmt.Sprintf(", <b>%s</b>!", html.EscapeString(c.TelegramUser().FirstName))
	}
	c.SetState("volunteer")
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
