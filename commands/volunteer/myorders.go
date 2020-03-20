package volunteer

import (
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

type findArgs struct {
	userID  *uint
	orderID int
	first   bool
	next    bool
}

func myFind(db *gorm.DB, args findArgs) *gorm.DB {
	r := db.Where(&models.Order{
		AssignedUserID: args.userID,
		Status:         models.OrderStatusPending,
	})
	if !args.first {
		if args.next {
			r = r.Where("id > ?", args.orderID).Order("id asc")
		} else {
			r = r.Where("id < ?", args.orderID).Order("id desc")
		}
	}
	return r.Limit(2)
}

func MyOrders(c *common.Ctx) {
	var orders []models.Order
	uid := uint(c.DbUser.TelegramID)
	if err := myFind(
		c.Db,
		findArgs{
			userID: &uid,
			first:  true,
		},
	).Find(&orders).Error; err != nil {
		c.HandleErr(err)
		return
	}
	if len(orders) == 0 {
		c.HandleErr(common.ReportableError{T: "Non hai ordini."})
		return
	}

	// Determine if we have multiple orders
	c.SetState("volunteer/my")
	s, err := orders[0].ToTelegram(c.Db)
	if err != nil {
		c.HandleErr(err)
		return
	}
	c.UpdateMenu(s, myOrdersKeyboard(int(orders[0].ID)), tb.ModeHTML)
}

func myChangeOrder(c *common.Ctx, next bool) error {
	payloadOID, err := strconv.Atoi(c.Callback.Data)
	if err != nil {
		return err
	}

	var orders []models.Order
	uid := uint(c.DbUser.TelegramID)
	if err := myFind(
		c.Db,
		findArgs{
			userID:  &uid,
			orderID: payloadOID,
			first:   false,
			next:    next,
		},
	).Find(&orders).Error; err != nil {
		return err
	}

	l := len(orders)
	if l == 0 {
		return common.ReportableError{T: text.NoMoreOrders}
	}
	// New order is always in index 0 because we order the results
	newOID := int(orders[0].ID)
	s, err := orders[0].ToTelegram(c.Db)
	if err != nil {
		return err
	}
	c.UpdateMenu(
		s,
		tb.ModeHTML,
		myOrdersKeyboard(newOID),
	)
	return nil
}

func MyNext(c *common.Ctx) {
	err := myChangeOrder(c, true)
	if err != nil {
		c.HandleErr(err)
	}
}

func MyPrevious(c *common.Ctx) {
	err := myChangeOrder(c, false)
	if err != nil {
		c.HandleErr(err)
	}
}
