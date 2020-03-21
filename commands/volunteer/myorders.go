package volunteer

import (
	"fmt"
	"strconv"
	"strings"

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

func getPendingOrderFor(db *gorm.DB, orderID, userID uint) (*models.Order, error) {
	var order models.Order
	uID := uint(userID)
	if err := db.Where(&models.Order{
		AssignedUserID: &uID,
		Status:         models.OrderStatusPending,
	}).Where(
		"id = ?", orderID,
	).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
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

func MyDone(c *common.Ctx) {
	parts := strings.Split(c.Callback.Data, "|")
	if len(parts) == 0 {
		c.HandleErr(common.IllegalPayloadReportableError)
		return
	}
	orderID, err := strconv.Atoi(parts[0])
	if err != nil {
		c.HandleErr(common.IllegalPayloadReportableError)
		return
	}
	definitive, err := strconv.ParseBool(parts[1])
	if err != nil {
		c.HandleErr(common.IllegalPayloadReportableError)
		return
	}

	// Make sure the order exists
	order, err := getPendingOrderFor(c.Db, uint(orderID), uint(c.DbUser.TelegramID))
	if err != nil {
		c.HandleErr(err)
		return
	}
	if order == nil {
		c.HandleErr(common.ReportableError{T: "L'ordine non esiste."})
		return
	}

	// Ok
	if !definitive {
		// Ask for confirmation
		c.UpdateMenu(
			"❓ **Sei sicuro di voler segnare questo ordine come completato?**",
			&tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{
					{
						{Unique: fmt.Sprintf("user__my_done|%d|1", orderID), Text: "👍 Sì"},
						{Unique: "user__my_orders", Text: "👎 No"},
					},
				},
			},
			tb.ModeMarkdown,
		)
	} else {
		// Delete for real
	}
}
