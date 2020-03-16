package common

import (
	"log"

	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/models"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Ctx ...
type Ctx struct {
	Message  *tb.Message
	Callback *tb.Callback
	B        *tb.Bot
	Db       *gorm.DB
	DbUser   *models.User
}

// IsCallback returns True if the current Ctx holds a callback and returns False if it holds a message
func (c *Ctx) IsCallback() bool {
	if c.Callback == nil && c.Message == nil {
		panic("this ctx does not hold anything!")
	}
	return c.Callback != nil
}

// TelegramUser returns the user user who send the callback, or the user who sent the message
func (c *Ctx) TelegramUser() *tb.User {
	if c.Callback != nil {
		return c.Callback.Sender
	}
	return c.Message.Sender
}

// Reply replies to the message held by the current Ctx.
// This works both with callbacks and normal messages.
func (c *Ctx) Reply(what interface{}, options ...interface{}) (*tb.Message, error) {
	var message *tb.Message
	if c.IsCallback() {
		message = c.Callback.Message
	} else {
		message = c.Message
	}
	return c.B.Send(message.Sender, what, options...)
}

// Answer answers to the callback query held by the current ctx
func (c *Ctx) Answer(r ...*tb.CallbackResponse) {
	c.B.Respond(c.Callback, r...)
}

// SetState updates the state of the user held by the current Ctx
func (c *Ctx) SetState(newState string) {
	log.Printf("%v's state -> %v", c.DbUser.TelegramID, newState)
	c.Db.Model(c.DbUser).Updates(map[string]interface{}{"state": newState})
}
