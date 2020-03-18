package admin

import (
	"encoding/json"
	"html"
	"log"
	"strings"
	"time"

	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/statemodels"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

var endDataReplyKeyboardMarkup *tb.ReplyMarkup = common.SingleKeyboardFactory("Fine")
var expireReplyKeyboardMarkup *tb.ReplyMarkup = common.SingleInlineKeyboardFactory(
	tb.InlineButton{Text: "⛔️ Nessuna scadenza", Unique: "admin__add_order__no_expire"},
)

// AddOrderData asks for all required data of the order that will be added
func AddOrderData(c *common.Ctx) {
	parts := strings.SplitN(c.Message.Text, "\n", 4)
	l := len(parts)
	if l < 3 {
		c.Reply(text.W("Non hai specificato tutti i dati richiesti!"), tb.ModeMarkdown)
		return
	}
	var notes *string
	if l > 3 {
		s := html.EscapeString(common.TruncateString(parts[3], 512))
		notes = &s
	}

	// Add order
	order := &models.Order{
		Name:      html.EscapeString(parts[0]),
		Address:   html.EscapeString(parts[1]),
		Telephone: html.EscapeString(parts[2]),
		Notes:     notes,
	}
	if err := c.Db.Create(order).Error; err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}

	// Area
	keyboard, err := common.AreasReplyKeyboard(c.Db)
	if err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	// TODO: tx
	c.SetState("admin/add_order/area")
	c.SetStateData(statemodels.Order{OrderID: order.ID})
	c.UpdateMenu(
		"✅🌆 **Bene!** Ora indica l'area della consegna.",
		&tb.ReplyMarkup{
			ReplyKeyboard: keyboard,
		},
		tb.ModeMarkdown,
	)
}

// AddOrderArea handles the area
func AddOrderArea(c *common.Ctx) {
	var stateData statemodels.Order
	if err := json.Unmarshal([]byte(c.DbUser.StateData), &stateData); err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}

	areaName := strings.TrimSpace(c.Message.Text)
	if len(areaName) == 0 {
		c.Reply(text.W("Nome area non valido!"), tb.ModeMarkdown)
		return
	}

	// Get area id by name
	area, err := models.GetAreaByName(c.Db, areaName)
	if err != nil {
		c.SessionError(err, BackReplyMarkup)
	}
	if area == nil {
		c.Reply(text.W("Non ho trovato nessuna area con quel nome."), tb.ModeMarkdown)
	}

	// Update order.area
	if err := c.Db.Model(&models.Order{}).Where("id = ?", stateData.OrderID).Update("area_id", area.ID).Error; err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}

	// Ok! Ask for expire
	c.SetState("admin/add_order/expire")
	c.UpdateMenu(
		"✅⏰ **Ottimo!** Indica la scadenza di questo ordine nel formato `gg/mm/aaaa hh:mm`",
		expireReplyKeyboardMarkup,
		tb.ModeMarkdown,
	)
}

func expireDone(c *common.Ctx) {
	// Ok! Ask for photos
	c.SetState("admin/add_order/attachments")
	c.UpdateMenu(
		"✅📷 **Ci siamo quasi!** Inviami ora eventuali foto da allegare. _Invia 'Fine' quando hai terminato di inviare gli allegati._",
		endDataReplyKeyboardMarkup,
		tb.ModeMarkdown,
	)
}

// AddOrderExpire adds an expiration date to the order
func AddOrderExpire(c *common.Ctx) {
	var stateData statemodels.Order
	if err := json.Unmarshal([]byte(c.DbUser.StateData), &stateData); err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	t, err := time.Parse("02/01/2006 15:04", c.Message.Text)
	if err != nil {
		// log.Printf("%v\n", err)
		c.Reply(
			text.W("Formato scadenza non valido!** Deve essere del tipo `gg/mm/aaaa hh:mm`"),
			tb.ModeMarkdown,
		)
		return
	}
	if err := c.Db.Model(
		&models.Order{},
	).Where(
		"id = ?", stateData.OrderID,
	).Update(
		"expire", &t,
	).Error; err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	expireDone(c)
}

// AddOrderNoExpire does not add an expire to the current order and skips to the next section
func AddOrderNoExpire(c *common.Ctx) {
	c.Respond(&tb.CallbackResponse{
		Text: "👍 Nessuna scadenza impostata",
	})
	expireDone(c)
}

// AddOrderAttachments processes incoming photos and saves them as attachments
func AddOrderAttachments(c *common.Ctx) {
	if c.Message.Photo == nil {
		c.Reply(text.W("Per favore invia una foto"), tb.ModeMarkdown)
		return
	}
	var stateData statemodels.Order
	if err := json.Unmarshal([]byte(c.DbUser.StateData), &stateData); err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	log.Printf("Received a photo for order %d\n", stateData.OrderID)
	if err := c.Db.Create(&models.Photo{
		FileID:  c.Message.Photo.FileID,
		OrderID: stateData.OrderID,
	}).Error; err != nil {
		panic(err)
	}
	c.UpdateMenu(
		"📸👍 **Foto ricevuta!** Puoi inviare altre foto oppure invia 'Fine' per terminare.",
		endDataReplyKeyboardMarkup,
		tb.ModeMarkdown,
	)
}

// AddOrderAttachmentsEnd processes the attachment end message
func AddOrderAttachmentsEnd(c *common.Ctx) {
	if c.Message.Text != "Fine" {
		return
	}
	var stateData statemodels.Order
	if err := json.Unmarshal([]byte(c.DbUser.StateData), &stateData); err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	if err := c.Db.Model(&models.Order{}).Where(
		"id = ? AND status = ?",
		stateData.OrderID,
		models.OrderStatusNeedsData,
	).Update(
		"status",
		models.OrderStatusPending,
	).Error; err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	var order models.Order
	if err := c.Db.Model(&models.Order{}).Where("id = ?", stateData.OrderID).First(&order).Error; err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	summary, err := order.ToTelegram(c.Db)
	if err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	c.SetState("admin/add_order/end")
	c.UpdateMenu("✅ <b>Ordine memorizzato!</b>\n\n🛒 <u>Riepilogo ordine</u>\n"+summary, BackReplyMarkup, tb.ModeHTML)
}
