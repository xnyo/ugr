package volunteer

import (
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

// TakeOrder starts the take order procedure, asking for the zone.
func TakeOrder(c *common.Ctx) {
	c.SetState("volunteer/take_order/zone")
	c.ClearStateData()
	keyboard, err := common.AreasReplyKeyboard(c.Db)
	if err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	keyboard = append(keyboard, []tb.ReplyButton{{Text: text.MainMenu}})
	c.UpdateMenu(
		"👇 **Seleziona la tua zona** 👇",
		&tb.ReplyMarkup{ReplyKeyboard: keyboard},
		tb.ModeMarkdown,
	)
}

func TakeOrderZone(c *common.Ctx) {

}
