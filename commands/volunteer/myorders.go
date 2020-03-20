package volunteer

import (
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/statemodels"
	tb "gopkg.in/tucnak/telebot.v2"
)

func MyOrders(c *common.Ctx) {
	var orders []models.Order
	uid := uint(c.DbUser.TelegramID)
	if err := c.Db.Where(&models.Order{
		AssignedUserID: &uid,
		Status:         models.OrderStatusPending,
	}).Limit(2).Find(&orders).Error; err != nil {
		c.HandleErr(err)
		return
	}
	if len(orders) == 0 {
		c.HandleErr(common.ReportableError{T: "Non hai ordini."})
		return
	}

	// Determine if we have multiple orders
	c.SetState("volunteer/my_orders")
	c.SetStateData(statemodels.VolunteerOrder{CurrentOrderID: orders[0].ID})
	s, err := orders[0].ToTelegram(c.Db)
	if err != nil {
		c.HandleErr(err)
		return
	}
	c.UpdateMenu(s, myOrderReplyMarkup, tb.ModeHTML)
}
