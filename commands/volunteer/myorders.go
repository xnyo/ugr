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

func getPhotos(db *gorm.DB, order *models.Order) (*[]models.Photo, error) {
	var photos []models.Photo
	if err := db.Model(
		&order,
	).Association(
		"Photos",
	).Find(&photos).Error; err != nil {
		return nil, err
	}
	return &photos, nil
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

	// Get photos
	photos, err := getPhotos(c.Db, &orders[0])
	if err != nil {
		c.HandleErr(err)
		return
	}

	// Determine if we have multiple orders
	c.SetState("volunteer/my")
	s, err := orders[0].ToTelegram(c.Db)
	if err != nil {
		c.HandleErr(err)
		return
	}
	c.UpdateMenu(s, myOrdersKeyboard(int(orders[0].ID), false, l > 1, len(*photos)), tb.ModeHTML)
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

	// Photos
	photos, err := getPhotos(c.Db, &orders[0])
	if err != nil {
		return err
	}

	c.UpdateMenu(
		s,
		tb.ModeHTML,
		myOrdersKeyboard(newOID, hasPrevious, hasNext, len(*photos)),
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

// MyPrevious handles the "<-" inline button of the "my orders" menu
func MyPrevious(c *common.Ctx) {
	err := myChangeOrder(c, false)
	if err != nil {
		c.HandleErr(err)
	}
}

type confirmation struct {
	text        string
	yesCbUnique string
	noCbUnique  string
}

func orderOpWithConfirm(c *common.Ctx, confirmationArgs confirmation) (bool, *models.Order, error) {
	parts := strings.Split(c.Callback.Data, "|")
	if len(parts) == 0 {
		return false, nil, common.IllegalPayloadReportableError
	}
	orderID, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, nil, common.IllegalPayloadReportableError
	}
	definitive, err := strconv.ParseBool(parts[1])
	if err != nil {
		return false, nil, common.IllegalPayloadReportableError
	}

	// Make sure the order exists
	order, err := getPendingOrderFor(c.Db, uint(orderID), uint(c.DbUser.TelegramID))
	if err != nil {
		return false, nil, err
	}
	if order == nil {
		return false, nil, common.ReportableError{T: "L'ordine non esiste."}
	}

	if definitive {
		// Run!
		return true, order, nil
	}

	// Confirmation
	c.UpdateMenu(
		confirmationArgs.text,
		&tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					{
						Unique: fmt.Sprintf("%s|%d|1", confirmationArgs.yesCbUnique, orderID),
						Text:   "👍 Sì",
					},
					{
						Unique: confirmationArgs.noCbUnique,
						Text:   "👎 No",
					},
				},
			},
		},
		tb.ModeMarkdown,
	)
	return false, order, nil
}

// MyDone completes an order (handles confirmation as well)
func MyDone(c *common.Ctx) {
	run, order, err := orderOpWithConfirm(c, confirmation{
		text:        "❓ **Sei sicuro di voler segnare questo ordine come completato?**",
		yesCbUnique: "user__my_done",
		noCbUnique:  "user__my_orders",
	})
	if err != nil {
		c.HandleErr(err)
		return
	}
	if !run {
		return
	}

	// Complete for real
	if err := c.Db.Model(&order).Updates(&models.Order{
		Status: models.OrderStatusDone,
	}).Error; err != nil {
		c.HandleErr(err)
		return
	}

	// Log
	s, err := order.ToTelegram(c.Db)
	if err != nil {
		c.HandleErr(err)
		return
	}
	err = c.LogToChan(c.Sign("😁 <b>Ordine completato</b>\n\n"+s), tb.ModeHTML)
	if err != nil {
		c.HandleErr(err)
		return
	}

	c.SetState("volunteer/my/done")
	c.UpdateMenu(
		"**L'ordine è stato segnato come completato!**\nGrazie per aver collaborato! 😄",
		BackReplyMarkup,
		tb.ModeMarkdown,
	)
}

func MyCancel(c *common.Ctx) {
	run, order, err := orderOpWithConfirm(c, confirmation{
		text:        "❓ **Sei sicuro di voler abbandonare questo ordine?**",
		yesCbUnique: "user__my_cancel",
		noCbUnique:  "user__my_orders",
	})
	if err != nil {
		c.HandleErr(err)
		return
	}
	if !run {
		return
	}

	// Log
	s, err := order.ToTelegram(c.Db)
	if err != nil {
		c.HandleErr(err)
		return
	}
	err = c.LogToChan(c.Sign("😞 <b>Ordine abbandonato</b>\n\n"+s), tb.ModeHTML)
	if err != nil {
		c.HandleErr(err)
		return
	}

	// Cancel
	if err := c.Db.Model(&order).Update("assigned_user_id", nil).Error; err != nil {
		c.HandleErr(err)
		return
	}
	c.SetState("volunteer/my/cancelled")
	c.UpdateMenu(
		"😞 **Hai abbandonato l'ordine**",
		BackReplyMarkup,
		tb.ModeMarkdown,
	)
}

func MyPhotos(c *common.Ctx) {
	// Read payload
	orderID, err := strconv.Atoi(c.Callback.Data)
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

	// Get photos
	photos, err := getPhotos(c.Db, order)
	if err != nil {
		c.HandleErr(err)
		return
	}
	if len(*photos) == 0 {
		c.HandleErr(common.ReportableError{T: "Non ci sono foto associate a questo ordine."})
		return
	}

	// Summary
	summary, err := order.ToTelegram(c.Db)
	if err != nil {
		c.HandleErr(err)
		return
	}

	// Build album
	var album tb.Album
	for _, v := range *photos {
		album = append(album, &tb.Photo{
			File: tb.File{FileID: v.FileID},
		})
	}
	msgs, err := c.B.SendAlbum(
		c.TelegramUser(),
		album,
	)
	if err != nil {
		c.HandleErr(err)
		return
	}

	// Reply (for reply markup & info)
	var messageIDs []string
	var replyMarkup *tb.ReplyMarkup
	for _, m := range msgs {
		messageIDs = append(messageIDs, strconv.Itoa(m.ID))
	}
	ids := strings.Join(messageIDs, "|")
	if len(ids) > 60 {
		// Limit is 64 chars 😱!
		replyMarkup = nil
	} else {
		// Limit ok
		replyMarkup = &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					{Text: "🗑 Elimina foto", Unique: "n|" + ids},
				},
				{BackReplyButton},
			},
		}
	}
	c.B.Reply(
		&msgs[0],
		"☝️ Ecco le foto! ☝️\n\n"+summary,
		replyMarkup,
		tb.ModeHTML,
	)
	c.Respond(&tb.CallbackResponse{Text: "👇 Foto inviate! 👇"})
}
