package admin

import (
	"github.com/xnyo/ugr/common"
	tb "gopkg.in/tucnak/telebot.v2"
)

// BackReplyButton is the "back" button goes back to the
// admin panel main menu
var BackReplyButton tb.InlineButton = tb.InlineButton{
	Unique: "admin",
	Text:   "👈 Menu principale",
}

// BackReplyMarkup is a tb.ReplyMarkup with a
// single "Main menu" button that goes to
// the admin panel main menu
var BackReplyMarkup *tb.ReplyMarkup = &tb.ReplyMarkup{
	InlineKeyboard: [][]tb.InlineButton{
		[]tb.InlineButton{BackReplyButton},
	},
}

// Menu send the admin menu as a reply
func Menu(c *common.Ctx) {
	c.SetState("admin")
	c.ClearStateData()
	c.UpdateMenu(
		"🔧 **Menu amministratore**",
		&tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				[]tb.InlineButton{
					tb.InlineButton{
						Unique: "admin__add_order",
						Text:   "🛒 Aggiungi ordine",
					},
					tb.InlineButton{
						Unique: "admin__orders",
						Text:   "🛍 Lista ordini",
					},
				},
				[]tb.InlineButton{
					tb.InlineButton{
						Unique: "admin__ban",
						Text:   "💣 Banna utente",
					},
					tb.InlineButton{
						Unique: "admin__users",
						Text:   "👨‍👩‍👧‍👦 Lista utenti",
					},
				},
				[]tb.InlineButton{
					tb.InlineButton{
						Unique: "admin__add_admin",
						Text:   "👑 Aggiungi admin",
					},
					tb.InlineButton{
						Unique: "admin__remove_admin",
						Text:   "👋 Rimuovi admin",
					},
				},
				[]tb.InlineButton{
					tb.InlineButton{
						Unique: "admin__add_area",
						Text:   "➕ Aggiungi zona",
					},
					tb.InlineButton{
						Unique: "admin__areas",
						Text:   "🌆 Lista zone",
					},
				},
			},
		},
		tb.ModeMarkdown,
	)
}
