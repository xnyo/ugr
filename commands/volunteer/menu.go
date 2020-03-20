package volunteer

import (
	"fmt"
	"html"

	"github.com/xnyo/ugr/models"
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

	// Count all orders
	var totalOrders int
	if err := c.Db.Model(&models.Order{}).Where(&models.Order{
		Status: models.OrderStatusPending,
	}).Count(&totalOrders).Error; err != nil {
		c.HandleErr(err)
		return
	}

	// Count user orders
	var myOrders int
	uid := uint(c.DbUser.TelegramID)
	if err := c.Db.Model(&models.Order{}).Where(&models.Order{
		Status:         models.OrderStatusPending,
		AssignedUserID: &uid,
	}).Count(&myOrders).Error; err != nil {
		c.HandleErr(err)
		return
	}

	s += fmt.Sprintf("\n\nAttualmente ci sono <b>%d</b> ordini disponibili, di cui <b>%d</b> assegnati a te.", totalOrders, myOrders)
	c.SetState("volunteer")
	c.ClearStateData()
	keyboard := [][]tb.InlineButton{
		{{Unique: "user__take_order_start", Text: "🛒 Scegli ordine"}},
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
