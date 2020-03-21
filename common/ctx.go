package common

import (
	"encoding/json"
	"fmt"
	"html"
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
	NoMenu      bool

	LogChannelID string
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
// TODO: make this private and replace all calls to
// 	it with c.HandleErr(common.ReportableError{T:...})
func (c *Ctx) Reply(what interface{}, options ...interface{}) (*tb.Message, error) {
	return c.B.Send(c.TelegramUser(), what, options...)
}

// Report reports a message to the user, in the most appropriate form
// based on the current ctx.
// - If the ctx holds a callback query, it sends a callback response
// - If the ctx holds an inline query, it sends a query result
// - If the ctx holds a message, it sends a message
func (c *Ctx) Report(txt string) {
	if c.Callback != nil {
		// Callback query
		c.Respond(&tb.CallbackResponse{
			ShowAlert: true,
			Text:      txt,
		})
	} else if c.InlineQuery != nil {
		// Inline query
		results := tb.Results{&tb.ArticleResult{Title: txt}}
		results[0].SetContent(&tb.InputTextMessageContent{Text: "⛔"})
		c.AnswerNoCache(&tb.QueryResponse{Results: results})
	} else if c.NoMenu {
		// Message -- no menu
		c.Reply(txt)
	} else {
		// Message -- menu
		c.UpdateMenu(
			txt,
			&tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{{{
					Unique: "volunteer",
					Text:   text.MainMenu,
				}}},
			},
		)
	}
}

// UpdateMenu updates a callback-based menu by deleting the message
// the callback originated from and sending a new one
func (c *Ctx) UpdateMenu(what interface{}, options ...interface{}) (*tb.Message, error) {
	if c.Callback != nil {
		// Delete callback query original message
		c.B.Delete(c.Callback.Message)
	} else if c.DbUser.LatestBotMsgID > 0 {
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
func (c *Ctx) Respond(r ...*tb.CallbackResponse) error {
	return c.B.Respond(c.Callback, r...)
}

// AnswerNoCache answers to an inline query, with no cache
func (c *Ctx) AnswerNoCache(r *tb.QueryResponse) error {
	r.CacheTime = 1
	r.IsPersonal = true
	return c.B.Answer(c.InlineQuery, r)
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

// HandleErr reports an error to sentry
func (c *Ctx) HandleErr(err error) {
	if err == nil {
		return
	}
	switch v := err.(type) {
	case ReportableError:
		v.Report(c)
	default:
		{
			if c.InlineQuery != nil {
				// Inline query, we cannot use menu!
				// We must report the error via inline results
				// NOT through the bot chat.
				results := tb.Results{&tb.ArticleResult{Title: text.SessionError}}
				results[0].SetContent(&tb.InputTextMessageContent{Text: "⛔"})
				c.AnswerNoCache(&tb.QueryResponse{Results: results})
			} else {
				// Unhandled error
				if c.DbUser != nil {
					c.SetState("error")
				}
				c.Report(text.SessionError)
				/* if c.NoMenu {
					// No menu
					c.Report(text.SessionError)
				} else {
					// Normal
					c.UpdateMenu(
						text.SessionError,
						&tb.ReplyMarkup{
							InlineKeyboard: [][]tb.InlineButton{{{
								Unique: "volunteer",
								Text:   text.MainMenu,
							}}},
						},
						tb.ModeMarkdown,
					)
				} */
			}

			// TODO: Sentry
			panic(err)
		}
	}
}

// LogToChan sends a message to the log channel
func (c *Ctx) LogToChan(what interface{}, options ...interface{}) error {
	channel, err := c.B.ChatByID(c.LogChannelID)
	if err != nil {
		return err
	}
	_, err = c.B.Send(channel, what, options...)
	return err
}

// Signature returns a string that contains
// the name and username of the user held
// by the current ctx
func (c *Ctx) Signature() string {
	return fmt.Sprintf(
		"✒️ <code>%s %s</code> (@%s)",
		html.EscapeString(c.TelegramUser().FirstName),
		html.EscapeString(c.TelegramUser().LastName),
		html.EscapeString(c.TelegramUser().Username),
	)
}

// Sign takes msg and appends c.Signature() to it
func (c *Ctx) Sign(msg string) string {
	return msg + "\n\n" + c.Signature()
}
