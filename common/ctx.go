package common

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Ctx ...
type Ctx struct {
	Message     *tb.Message
	Callback    *tb.Callback
	InlineQuery *tb.Query
	B           *tb.Bot
	Db          *gorm.DB
	DbUser      *models.User
}

// TelegramUser returns the user user who send the callback, or the user who sent the message
func (c *Ctx) TelegramUser() *tb.User {
	if c.Callback != nil {
		return c.Callback.Sender
	}
	if c.InlineQuery != nil {
		return &c.InlineQuery.From
	}
	return c.Message.Sender
}

// Reply replies to the message held by the current Ctx.
// This works both with callbacks and normal messages.
func (c *Ctx) Reply(what interface{}, options ...interface{}) (*tb.Message, error) {
	return c.B.Send(c.TelegramUser(), what, options...)
}

// Report reports a message to the user, in the most appropriate form
// based on the current ctx.
// - If the ctx holds a callback query, it sends a callback response
// - If the ctx holds an inline query, it sends a query result
// - If the ctx holds a message, it sends a message
func (c *Ctx) Report(text string) {
	if c.Callback != nil {
		// Callback query
		c.Respond(&tb.CallbackResponse{
			ShowAlert: true,
			Text:      text,
		})
	} else if c.InlineQuery != nil {
		// Inline query
		results := tb.Results{
			&tb.ArticleResult{
				Title: text,
			},
		}
		// Make sure the id is unique
		results[0].SetResultID(fmt.Sprintf("%s_%s", GetMD5Hash(text), string(c.TelegramUser().ID)))
		results[0].SetContent(&tb.InputTextMessageContent{Text: "⛔"})
		c.B.Answer(c.InlineQuery, &tb.QueryResponse{Results: results})
	} else {
		// Message
		c.Reply(text)
	}
}

// UpdateMenu updates a callback-based menu by deleting the message
// the callback originated from and sending a new one
func (c *Ctx) UpdateMenu(what interface{}, options ...interface{}) (*tb.Message, error) {
	if c.Callback != nil {
		// Delete callback query original message
		c.B.Delete(c.Callback.Message)
	} else {
		// Delete latest message stored in db
		c.B.Delete(c.DbUser)
	}

	// New message
	msg, err := c.Reply(what, options...)

	// Update message id in db
	c.Db.Model(c.DbUser).Update("latest_bot_msg_id", msg.ID)
	return msg, err
}

// Respond responds to the callback query held by the current ctx
func (c *Ctx) Respond(r ...*tb.CallbackResponse) {
	c.B.Respond(c.Callback, r...)
}

// SetState updates the state of the user held by the current Ctx
func (c *Ctx) SetState(newState string) {
	log.Printf("%v's state -> %v", c.DbUser.TelegramID, newState)
	if err := c.Db.Model(c.DbUser).Update("state", newState).Error; err != nil {
		panic(err)
	}
}

/* func (c *Ctx) AddStateData(key string, value interface{}) {
	err := c.Db.Transaction(func(tx *gorm.DB) error {
		var user *models.User
		if err := tx.Where(c.DbUser).First(&user).Error; err != nil {
			return err
		}

		stateData := make(map[string]interface{})
		err := json.Unmarshal([]byte(user.StateData), stateData)
		if err != nil {
			panic(err)
		}
		stateData[key] = value
		jsonData, err := json.Marshal(stateData)
		if err != nil {
			panic(err)
		}
		if err := tx.Model(&user).Update("state_data", string(jsonData)).Error; err != nil {
			return err
		}
		c.DbUser = user
		return nil
	})
	if err != nil {
		panic(err)
	}
} */

func (c *Ctx) ClearStateData() {
	if err := c.Db.Model(&c.DbUser).Update("state_data", "{}").Error; err != nil {
		panic(err)
	}
	c.DbUser.StateData = "{}"
}

func (c *Ctx) SetStateData(data interface{}) {
	j, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	s := string(j)
	if err := c.Db.Model(&c.DbUser).Update("state_data", s).Error; err != nil {
		panic(err)
	}
	c.DbUser.StateData = s
}

func sessionError(c *Ctx, err error) {
	c.SetState("error")
	c.HandleErr(err)
}

func (c *Ctx) SessionError(err error, replyMarkup *tb.ReplyMarkup) {
	sessionError(c, err)
	c.Reply(text.SessionError, replyMarkup, tb.ModeMarkdown)
}

func (c *Ctx) RespondSessionError(err error) {
	sessionError(c, err)
	c.Respond(&tb.CallbackResponse{
		Text:      text.SessionError,
		ShowAlert: true,
	})
}

// HandleErr reports an error to sentry
func (c *Ctx) HandleErr(err error) {
	// TODO: Sentry
	panic(err)
}
