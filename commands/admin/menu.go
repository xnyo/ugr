package admin

import (
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

// BackReplyButton is the "back" button goes back to the
// admin panel main menu
var BackReplyButton tb.InlineButton = tb.InlineButton{
	Unique: "admin",
	Text:   text.MainMenu,
}

// BackReplyMarkup is a tb.ReplyMarkup with a
// single "Main menu" button that goes to
// the admin panel main menu
var BackReplyMarkup *tb.ReplyMarkup = &tb.ReplyMarkup{
	InlineKeyboard: [][]tb.InlineButton{
		{BackReplyButton},
	},
}

// Menu send the admin menu as a reply
func Menu(c *common.Ctx) {
	c.SetState("admin")
	c.ClearStateData()
	c.UpdateMenu(
		"🔧 <b>Pannello amministratore</b>",
		&tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					{
						Unique: "admin__add_order",
						Text:   "🛒 Aggiungi ordine",
					},
					{
						Unique: "admin__orders",
						Text:   "🛍 Lista ordini",
					},
				},
				{
					{
						InlineQuery: "volunteer",
						Text:        "🛵 Aggiungi volontario",
					},
					{
						InlineQuery: "admin",
						Text:        "👑 Aggiungi admin",
					},
				},
				{
					{
						Unique: "admin__ban",
						Text:   "💣 Banna utente",
					},
					{
						Unique: "admin__users",
						Text:   "👨‍👩‍👧‍👦 Lista utenti",
					},
				},
				{
					{
						Unique: "admin__add_area",
						Text:   "➕ Aggiungi zona",
					},
					{
						Unique: "admin__areas",
						Text:   "🌆 Lista zone",
					},
				},
				{
					{
						Unique: "volunteer",
						Text:   "🛵 Pannello volontario",
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
