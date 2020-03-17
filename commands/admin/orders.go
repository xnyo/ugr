package admin

import (
	"encoding/json"
	"html"
	"log"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/statemodels"
	tb "gopkg.in/tucnak/telebot.v2"
)

var endDataReplyKeyboardMarkup *tb.ReplyMarkup = &tb.ReplyMarkup{
	ReplyKeyboard: [][]tb.ReplyButton{
		[]tb.ReplyButton{
			tb.ReplyButton{
				Text: "Fine",
			},
		},
	},
}

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
	var areas []models.Area
	if err := c.Db.Where("visible = 1").Find(&areas).Error; err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}
	var keyboard [][]tb.ReplyButton
	for _, v := range areas {
		keyboard = append(keyboard, []tb.ReplyButton{tb.ReplyButton{Text: v.Name}})
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
		c.Reply("⚠️ **Nome area non valido!**", tb.ModeMarkdown)
		return
	}

	// Get area id by name
	var area models.Area
	if err := c.Db.Where("name = ?", areaName).First(&area).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No area
			c.Reply("⚠️ Non ho trovato nessuna area con quel nome.", tb.ModeMarkdown)
		} else {
			// Other err
			c.SessionError(err, BackReplyMarkup)
		}
		return
	}

	// Update order.area
	if err := c.Db.Model(&models.Order{}).Where("id = ?", stateData.OrderID).Update("area_id", area.ID).Error; err != nil {
		c.SessionError(err, BackReplyMarkup)
		return
	}

	// Ok! Ask for photos
	c.SetState("admin/add_order/attachments")
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
