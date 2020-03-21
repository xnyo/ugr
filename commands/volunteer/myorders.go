package volunteer

import (
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

type myFindArgs struct {
	userID  *uint
	orderID int
	next    bool
	first   bool
}

// myFind builds a gorm query that looks for either:
// - The first pending order assigned to the current user
// - The next order
// - The previous order
// The query always has a "LIMIT 2"
// The result can be:
// - An empty slice. In this case, there's no (next/previous) order
// - A slice of length 1. In this case, there's a (next/previous) order,
// but there's not another one in the same "direction"
// - A slice of length 2. In this case, the element at index 0 is the
// next (previous) order that comes after (before) args.orderID, and the
// element at index 1 is the element that comes after (before) the element
// at index 0.
func myFind(db *gorm.DB, args myFindArgs) *gorm.DB {
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

// MyOrders fetches the order assigned to the current user
func MyOrders(c *common.Ctx) {
	var orders []models.Order
	uid := uint(c.DbUser.TelegramID)
	if err := myFind(
		c.Db,
		myFindArgs{
			userID: &uid,
			first:  true,
		},
	).Find(&orders).Error; err != nil {
		c.HandleErr(err)
		return
	}
	l := len(orders)
	if l == 0 {
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
	c.UpdateMenu(s, myOrdersKeyboard(int(orders[0].ID), false, l > 1), tb.ModeHTML)
}

// myChangeOrder changes the order displayed in the "my orders"
// panel held by the current ctx with the next (previous) order
func myChangeOrder(c *common.Ctx, next bool) error {
	payloadOID, err := strconv.Atoi(c.Callback.Data)
	if err != nil {
		return err
	}

	var orders []models.Order
	uid := uint(c.DbUser.TelegramID)
	if err := myFind(
		c.Db,
		myFindArgs{
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

	// Determine hasPrevious/hasNext
	var hasPrevious, hasNext bool
	if next {
		hasPrevious = true
		hasNext = l > 1
	} else {
		hasNext = true
		hasPrevious = l > 1
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
		myOrdersKeyboard(newOID, hasPrevious, hasNext),
	)
	return nil
}

// MyNext handles the "->" inline button of the "my orders" menu
func MyNext(c *common.Ctx) {
	err := myChangeOrder(c, true)
	if err != nil {
		c.HandleErr(err)
	}
}

// MyNext handles the "<-" inline button of the "my orders" menu
func MyPrevious(c *common.Ctx) {
	err := myChangeOrder(c, false)
	if err != nil {
		c.HandleErr(err)
	}
}
