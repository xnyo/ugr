package volunteer

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/statemodels"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

var previousButton = tb.InlineButton{Text: "⬅️", Unique: "user__previous_order"}
var nextButton = tb.InlineButton{Text: "➡️", Unique: "user__next_order"}
var takeButton = tb.InlineButton{Text: "✔️", Unique: "dummy"}

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
	}).Where(
		"status == ?",
		models.OrderStatusPending,
	).Limit(2).Find(&orders).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.UpdateMenu("Non ci sono altri ordini.")
		} else {
			c.SessionError(err, BackReplyMarkup)
		}
		return
	}

	// Determine if we have multiple orders
	c.SetState("volunteer/take_order/order")
	c.SetStateData(
		statemodels.VolunteerOrder{
			CurrentOrderID: orders[0].ID,
			CurrentAreaID:  *orders[0].AreaID,
		},
	)
	s, err := orders[0].ToTelegram(area)
	if err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	keyboard := [][]tb.InlineButton{{takeButton}, {BackReplyButton}}
	if len(orders) > 1 {
		keyboard[0] = append(keyboard[0], nextButton)
	}
	c.UpdateMenu(
		s,
		tb.ModeHTML,
		&tb.ReplyMarkup{InlineKeyboard: keyboard},
	)
}

func changeOrder(c *common.Ctx, next bool) {
	var stateData statemodels.VolunteerOrder
	if err := json.Unmarshal([]byte(c.DbUser.StateData), &stateData); err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	var orders []models.Order
	where := "status == ?"
	args := []interface{}{models.OrderStatusPending}
	if next {
		where += " AND id > ?"
	} else {
		where += " AND id < ?"
	}
	args = append(args, stateData.CurrentOrderID)
	if err := c.Db.Where(&models.Order{
		AreaID:         &stateData.CurrentAreaID,
		AssignedUserID: nil,
	}).Where(
		where, args...,
	).Limit(2).Find(&orders).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.UpdateMenu("Non ci sono altri ordini.")
		} else {
			c.SessionError(err, BackReplyMarkup)
		}
	}

	// Update menu
	var newOrderIdx int
	l := len(orders)
	if next {
		newOrderIdx = 0
	} else {
		newOrderIdx = l - 1
	}
	s, err := orders[newOrderIdx].ToTelegram(c.Db)
	if err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	var keyboard [][]tb.InlineButton
	if next {
		keyboard = [][]tb.InlineButton{
			{previousButton, takeButton},
			{BackReplyButton},
		}
		if l >= 2 {
			// has next
			keyboard[0] = append(keyboard[0], tb.InlineButton{Text: "➡️", Unique: "user__next_order"})
		}
	} else {
		keyboard = [][]tb.InlineButton{
			{takeButton, nextButton},
			{BackReplyButton},
		}
		if l >= 2 {
			// has prev
			keyboard[0] = append([]tb.InlineButton{previousButton}, keyboard[0]...)
		}
	}

	c.SetStateData(
		statemodels.VolunteerOrder{
			CurrentOrderID: orders[newOrderIdx].ID,
			CurrentAreaID:  stateData.CurrentAreaID,
		},
	)
	c.UpdateMenu(
		s,
		tb.ModeHTML,
		&tb.ReplyMarkup{InlineKeyboard: keyboard},
	)
}

func NextOrder(c *common.Ctx) {
	changeOrder(c, true)
}

func PreviousOrder(c *common.Ctx) {
	changeOrder(c, false)
}