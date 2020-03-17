package admin

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/statemodels"
	tb "gopkg.in/tucnak/telebot.v2"
)

// AddOrderData asks for all required data of the order that will be added
func AddOrderData(c *common.Ctx) {
	parts := strings.SplitN(c.Message.Text, "\n", 4)
	l := len(parts)
	if l < 3 {
		c.Reply("⚠️ **Non hai specificato tutti i dati richiesti!**", tb.ModeMarkdown)
		return
	}
	var notes *string
	if l > 3 {
		s := common.TruncateString(parts[3], 512)
		notes = &s
	}

	// Add order
	order := &models.Order{
		Name:      parts[0],
		Address:   parts[1],
		Telephone: parts[2],
		Notes:     notes,
	}
	if err := c.Db.Create(order).Error; err != nil {
		panic(err)
	}

	// Ask for photos
	// TODO: tx
	c.SetState("admin/add_order/attachments")
	c.SetStateData(statemodels.Order{OrderID: order.ID})
	c.UpdateMenu(
		"✅📷 **Ci siamo quasi!** Inviami ora eventuali foto da allegare. _Invia 'Fine' quando hai terminato di inviare gli allegati._",
		&tb.ReplyMarkup{
			ReplyKeyboard: [][]tb.ReplyButton{
				[]tb.ReplyButton{
					tb.ReplyButton{
						Text: "Fine",
					},
				},
			},
		},
		tb.ModeMarkdown,
	)
}

// AddOrderAttachments processes incoming photos and saves them as attachments
func AddOrderAttachments(c *common.Ctx) {
	if c.Message.Photo == nil {
		c.Reply("⚠️ **Per favore invia una foto**", tb.ModeMarkdown)
		return
	}
	var stateData statemodels.Order
	if err := json.Unmarshal([]byte(c.DbUser.StateData), &stateData); err != nil {
		c.Reply("⚠️ **Si è verificato un errore nella sessione corrente**. Per favore, ricomincia.", BackReplyMarkup, tb.ModeMarkdown)
		return
	}
	log.Printf("Received a photo for order %d\n", stateData.OrderID)
	if err := c.Db.Create(&models.Photo{
		FileID:  c.Message.Photo.FileID,
		OrderID: stateData.OrderID,
	}).Error; err != nil {
		panic(err)
	}
	c.UpdateMenu("📸👍 **Foto ricevuta!** Puoi inviare altre foto oppure invia 'Fine' per terminare.", tb.ModeMarkdown)
}

// AddOrderAttachmentsEnd processes the attachment end message
func AddOrderAttachmentsEnd(c *common.Ctx) {
	if c.Message.Text != "Fine" {
		return
	}
	var stateData statemodels.Order
	if err := json.Unmarshal([]byte(c.DbUser.StateData), &stateData); err != nil {
		c.Reply("⚠️ **Si è verificato un errore nella sessione corrente**. Per favore, ricomincia.", BackReplyMarkup, tb.ModeMarkdown)
		return
	}
	c.Db.Model(&models.Order{}).Where(
		"id = ? AND status = ?",
		stateData.OrderID,
		models.OrderStatusNeedsData,
	).Update(
		"status",
		models.OrderStatusPending,
	)
	c.SetState("admin/add_order/end")
	c.UpdateMenu("✅ **Ordine memorizzato!**", BackReplyMarkup, tb.ModeMarkdown)
}
