package volunteer

import (
	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/statemodels"
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
	// Fetch area
	area, err := models.GetAreaByName(c.Db, c.Message.Text)
	if err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	if area == nil {
		c.Reply(text.W("L'area specificata non esiste"), tb.ModeMarkdown)
		return
	}

	// Fetch next order
	var orders []models.Order
	if err := c.Db.Where(&models.Order{
		AreaID:         &area.ID,
		AssignedUserID: nil,
	}).Find(&orders).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.UpdateMenu("Non ci sono altri ordini.")
		} else {
			c.SessionError(err, BackReplyMarkup)
		}
		return
	}

	// Determine if we have multiple orders
	hasNext := len(orders) > 1
	c.SetStateData(
		statemodels.VolunteerOrder{
			CurrentOrderID: orders[0].ID,
			HasNext:        hasNext,
			HasPrevious:    false,
		},
	)

	// Update menu
	s, err := orders[0].ToTelegram(area)
	if err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	keyboard := [][]tb.InlineButton{{{Text: "✔️", Unique: "dummy"}}, {BackReplyButton}}
	if hasNext {
		keyboard[0] = append(keyboard[0], tb.InlineButton{Text: "➡️", Unique: "dummy"})
	}
	c.UpdateMenu(
		s,
		tb.ModeHTML,
		&tb.ReplyMarkup{InlineKeyboard: keyboard},
	)
}
