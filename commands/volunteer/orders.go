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

var (
	previousButton = tb.InlineButton{Text: "⬅️", Unique: "user__previous_order"}
	nextButton     = tb.InlineButton{Text: "➡️", Unique: "user__next_order"}
	takeButton     = tb.InlineButton{Text: "✔️", Unique: "user__take_order"}
)

// TakeOrderStart starts the take order procedure, asking for the zone.
func TakeOrderStart(c *common.Ctx) {
	c.SetState("volunteer/take_order/zone")
	c.ClearStateData()
	keyboard, err := common.AreasReplyKeyboard(c.Db)
	if err != nil {
		c.HandleErr(err)
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
	// TODO: use changeOrder somehow. Repeated code :(
	area, err := models.GetAreaByName(c.Db, c.Message.Text)
	if err != nil {
		c.HandleErr(err)
		return
	}
	if area == nil {
		c.HandleErr(common.ReportableError{T: "L'area specificata non esiste"})
		return
	}
	// Fetch next order
	var orders []models.Order
	if err := c.Db.Where(&models.Order{
		AreaID: &area.ID,
		Status: models.OrderStatusPending,
	}).Where(
		"assigned_user_id IS NULL",
	).Limit(2).Find(&orders).Error; err != nil {
		c.HandleErr(err)
		return
	}

	// Empty set check
	if len(orders) == 0 {
		c.HandleErr(common.ReportableError{T: text.NoMoreOrders})
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
		c.HandleErr(err)
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

func changeOrder(c *common.Ctx, next bool) error {
	var stateData statemodels.VolunteerOrder
	if err := json.Unmarshal([]byte(c.DbUser.StateData), &stateData); err != nil {
		return err
	}
	var orders []models.Order
	// nil in Where() does not work...
	where := "assigned_user_id IS NULL"
	args := []interface{}{stateData.CurrentOrderID}
	if next {
		where += " AND id > ?"
	} else {
		where += " AND id < ?"
	}
	if err := c.Db.Where(&models.Order{
		AreaID: &stateData.CurrentAreaID,
		Status: models.OrderStatusPending,
	}).Where(
		where, args...,
	).Limit(2).Find(&orders).Error; err != nil {
		return err
	}

	// Update menu
	var newOrderIdx int
	l := len(orders)
	if l == 0 {
		c.HandleErr(common.ReportableError{T: text.NoMoreOrders})
	}
	if next {
		newOrderIdx = 0
	} else {
		newOrderIdx = l - 1
	}
	s, err := orders[newOrderIdx].ToTelegram(c.Db)
	if err != nil {
		return err
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
	return nil
}

func NextOrder(c *common.Ctx) {
	err := changeOrder(c, true)
	if err != nil {
		switch v := err.(type) {
		case common.ReportableError:
			v.Report(c)
		default:
			c.HandleErr(err)
		}
		return
	}
}

func PreviousOrder(c *common.Ctx) {
	changeOrder(c, false)
}

func TakeOrder(c *common.Ctx) {
	var stateData statemodels.VolunteerOrder
	if err := json.Unmarshal([]byte(c.DbUser.StateData), &stateData); err != nil {
		c.HandleErr(err)
		return
	}
	err := c.Db.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := c.Db.Model(&order).Where(
			"id = ?", stateData.CurrentOrderID,
		).First(&order).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return common.ReportableError{T: "Ordine non trovato. Per favore, ricomincia la procedura."}
			}
			return err
		}
		if order.Status != models.OrderStatusPending || order.AssignedUserID != nil {
			return common.ReportableError{T: "Questo ordine è già stato preso da un altro volontario. Per favore, scegline un altro."}
		}
		id := uint(c.DbUser.TelegramID)
		if err := c.Db.Model(&order).Update("assigned_user_id", &id).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		switch v := err.(type) {
		case common.ReportableError:
			v.Report(c)
		default:
			c.HandleErr(err)
		}
		return
	}
	c.UpdateMenu("Utopico!")
}
