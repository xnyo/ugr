package bot

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/privileges"
	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/xnyo/ugr/models"
)

// CommandHandler is a F that accepts a Ctx as a parameter
type CommandHandler func(c *common.Ctx)

// privateOnly is a decorator that runs the telebot handler only if the message is a private message
func privateOnly(f CommandHandler) CommandHandler {
	return func(c *common.Ctx) {
		if !c.Message.Private() {
			return
		}
		f(c)
	}
}

func resolveUser(f CommandHandler) CommandHandler {
	return func(c *common.Ctx) {
		var user models.User

		// Try to get user from db.
		// If it does not exist, a new user with default privileges is created
		if result := c.Db.Where(models.User{
			TelegramID: c.TelegramUser().ID,
		}).Attrs(models.User{
			TelegramID: c.TelegramUser().ID,
			Privileges: privileges.Normal,
		}).FirstOrCreate(&user); result.Error != nil {
			panic(result.Error)
		}
		c.DbUser = &user

		// Ban check
		if user.Privileges&privileges.Normal == 0 {
			c.Reply("Non hai i permessi necessari per utilizzare questo bot.")
			return
		}

		f(c)
	}
}

func handleErrors(f CommandHandler) CommandHandler {
	return func(c *common.Ctx) {
		defer func() {
			if rec := recover(); rec != nil {
				// recover from panic 😱
				var err error
				switch rec := rec.(type) {
				case string:
					err = errors.New(rec)
				case error:
					err = rec
				default:
					err = fmt.Errorf("%v - %#v", rec, rec)
				}
				log.Printf("ERROR !!!\n%v\n%s", err, string(debug.Stack()))
				c.Reply("Si è verificato un errore.")
			}
		}()
		f(c)
	}
}

func protected(f CommandHandler, requiredPrivileges privileges.Privileges) CommandHandler {
	return func(c *common.Ctx) {
		if c.DbUser.Privileges&requiredPrivileges == 0 {
			log.Printf("%v (%v) does not have the required privileges (%v) to trigger %v", c.DbUser.TelegramID, c.DbUser.Privileges, requiredPrivileges, f)
			return
		}
		f(c)
	}
}

func fsm(f CommandHandler, requiredState string) CommandHandler {
	return func(c *common.Ctx) {
		if c.DbUser.State != requiredState {
			log.Printf("%v (%v) does not have the required state (%v) to trigger %v", c.DbUser.TelegramID, c.DbUser.State, requiredState, f)
			return
		}
		f(c)
	}
}

func textPrompt(message string, newState string, options ...interface{}) CommandHandler {
	return func(c *common.Ctx) {
		c.SetState(newState)
		c.Respond()
		c.UpdateMenu("✒️ "+message, options...)
	}
}

// wrapCtxMessage converts a telebot message handler to a custom ugr handler (with ctx)
func wrapCtxMessage(f CommandHandler) func(*tb.Message) {
	return func(m *tb.Message) {
		c := NewCtx(m, nil, nil)
		f(&c)
	}
}

// wrapCtxCallback converts a telebot message handler to a custom ugr callback handler (with ctx)
func wrapCtxCallback(f CommandHandler) func(*tb.Callback) {
	return func(c *tb.Callback) {
		ctx := NewCtx(nil, c, nil)
		f(&ctx)
	}
}

// wrapCtxQuery converts a telebot message handler to a custom ugr Query handler (with ctx)
func wrapCtxQuery(f CommandHandler) func(*tb.Query) {
	return func(q *tb.Query) {
		ctx := NewCtx(nil, nil, q)
		f(&ctx)
	}
}

type Handler struct {
	F CommandHandler
	P privileges.Privileges
	S string
}

func (h Handler) wrap() {
	if h.P > 0 {
		h.F = protected(h.F, h.P)
	}
	if h.S != "" {
		h.F = fsm(h.F, h.S)
	}
}

func (h Handler) BaseWrap() func(*tb.Message) {
	h.wrap()
	return wrapCtxMessage(handleErrors(resolveUser(privateOnly(h.F))))
}

func (h Handler) BaseWrapCb() func(*tb.Callback) {
	h.wrap()
	return wrapCtxCallback(handleErrors(resolveUser(privateOnly(h.F))))
}

func (h Handler) BaseWrapQ() func(*tb.Query) {
	h.wrap()
	return wrapCtxQuery(handleErrors(resolveUser(h.F)))
}

func (h Handler) TextWrap() func(*common.Ctx) {
	h.wrap()
	return h.F
}
