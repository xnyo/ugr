package commands

import (
	"github.com/xnyo/ugr/common"
	tb "gopkg.in/tucnak/telebot.v2"
)

// AdminNewArea handles /aggarea
func AdminNewArea(c *common.Ctx) {

}

func AdminMenu(c *common.Ctx) {
	inlineKeys := [][]tb.InlineButton{
		[]tb.InlineButton{
			tb.InlineButton{
				Unique: "Test",
				Text:   "🌕 Button #1",
			},
		},
	}
	c.Reply("Il menu di maria francesca", &tb.ReplyMarkup{
		InlineKeyboard: inlineKeys,
	})
}
