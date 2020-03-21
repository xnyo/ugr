package volunteer

import (
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

// ChooseOrderStart starts the take order procedure, asking for the zone.
func ChooseOrderStart(c *common.Ctx) {
	c.SetState("volunteer/choose/zone")
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

// ChooseOrderZone is the raw text handler to choose the zone
func ChooseOrderZone(c *common.Ctx) {
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
		c.HandleErr(common.ReportableError{T: text.NoMoreOrdersZone})
		return
	}

	// Determine if we have multiple orders
	c.SetState("volunteer/choose/order")
	s, err := orders[0].ToTelegram(area)
	if err != nil {
		c.HandleErr(err)
		return
	}
	oID := int(orders[0].ID)
	aID := int(area.ID)
	keyboard := [][]tb.InlineButton{
		{chooseOrderConfirm(oID)},
		{BackReplyButton},
	}
	if len(orders) > 1 {
		keyboard[0] = append(keyboard[0], chooseOrderNext(aID, oID))
	}
	c.UpdateMenu(
		s,
		tb.ModeHTML,
		&tb.ReplyMarkup{InlineKeyboard: keyboard},
	)
}

// changeOrder changes the current "choose order" menu with
// the next/previous one.
func changeOrder(c *common.Ctx, next bool) error {
	payload := strings.Split(c.Callback.Data, "|")
	if len(payload) != 2 {
		return common.IllegalPayloadReportableError
	}
	payloadAID, err := strconv.Atoi(payload[0])
	if err != nil {
		return err
	}
	payloadOID, err := strconv.Atoi(payload[1])
	if err != nil {
		return err
	}

	var orders []models.Order
	// nil in Where() does not work...
	where := "assigned_user_id IS NULL AND area_id = ?"
	args := []interface{}{payloadAID, payloadOID}
	var order string
	if next {
		where += " AND id > ?"
		order = "id asc"
	} else {
		where += " AND id < ?"
		order = "id desc"
	}
	if err := c.Db.Where(&models.Order{
		Status: models.OrderStatusPending,
	}).Where(
		where, args...,
	).Order(order).Limit(2).Find(&orders).Error; err != nil {
		return err
	}

	// Update menu
	l := len(orders)
	if l == 0 {
		return common.ReportableError{T: text.NoMoreOrdersZone}
	}
	// New order is always in index 0 because we order the results
	newOID := int(orders[0].ID)
	s, err := orders[0].ToTelegram(c.Db)
	if err != nil {
		return err
	}
	var keyboard [][]tb.InlineButton
	if next {
		keyboard = [][]tb.InlineButton{
			{chooseOrderPrevious(payloadAID, newOID), chooseOrderConfirm(newOID)},
			{BackReplyButton},
		}
		if l >= 2 {
			// has next
			keyboard[0] = append(keyboard[0], chooseOrderNext(payloadAID, newOID))
		}
	} else {
		keyboard = [][]tb.InlineButton{
			{chooseOrderConfirm(newOID), chooseOrderNext(payloadAID, newOID)},
			{BackReplyButton},
		}
		if l >= 2 {
			// has prev
			keyboard[0] = append([]tb.InlineButton{chooseOrderPrevious(payloadAID, newOID)}, keyboard[0]...)
		}
	}
	c.UpdateMenu(
		s,
		tb.ModeHTML,
		&tb.ReplyMarkup{InlineKeyboard: keyboard},
	)
	return nil
}

// ChooseNext handles the "->" inline button in the choose order panel
func ChooseNext(c *common.Ctx) {
	err := changeOrder(c, true)
	if err != nil {
		c.HandleErr(err)
	}
}

// ChoosePrevious handles the "<-" inline button in the choose order panel
func ChoosePrevious(c *common.Ctx) {
	err := changeOrder(c, false)
	if err != nil {
		c.HandleErr(err)
	}
}

// ChooseConfirm assigns a order to the current volunteer
func ChooseConfirm(c *common.Ctx) {
	orderID, err := strconv.Atoi(c.Callback.Data)
	if err != nil {
		c.HandleErr(err)
		return
	}
	var order models.Order
	err = c.Db.Transaction(func(tx *gorm.DB) error {
		if err := c.Db.Model(&order).Where(
			"id = ?", orderID,
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
		c.HandleErr(err)
		return
	}
	s, err := order.ToTelegram(c.Db)
	if err != nil {
		c.HandleErr(err)
		return
	}
	var previous, next []models.Order
	var hasPrevious, hasNext bool
	// Try to find previous
	if err := myFind(c.Db, myFindArgs{
		orderID: int(order.ID),
		next:    false,
	}).Find(&previous).Error; err != nil {
		c.HandleErr(err)
		return
	}
	hasPrevious = len(previous) > 1

	// Try to find next
	if err := myFind(c.Db, myFindArgs{
		orderID: int(order.ID),
		next:    true,
	}).Find(&next).Error; err != nil {
		c.HandleErr(err)
		return
	}
	hasNext = len(next) > 1

	// Find photos
	photos, err := getPhotos(c.Db, &order)
	if err != nil {
		c.HandleErr(err)
		return
	}

	c.SetState("volunteer/my")
	c.UpdateMenu(
		`🛍 <b>Hai preso questo ordine!</b>
Ora sarà visibile dalla lista 'I miei ordini'
<i>Non dimenticarti di segnarlo come 'Completato' una volta portato a termine!</i>

`+s,
		tb.ModeHTML,
		myOrdersKeyboard(orderID, hasPrevious, hasNext, len(*photos) > 0),
	)
}
