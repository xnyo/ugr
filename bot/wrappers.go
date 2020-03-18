package bot

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/jinzhu/gorm"

	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/privileges"
	"github.com/xnyo/ugr/text"
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

// resolveUser populates c.DbUser with the models.User
// that corresponds to c.TelegramUser() (user sending
// the message, sending the cb query or the inline
// query request)
func resolveUser(f CommandHandler) CommandHandler {
	return func(c *common.Ctx) {
		var user models.User

		// Try to get user from db
		if err := c.Db.Where(models.User{
			TelegramID: c.TelegramUser().ID,
		}).First(&user).Error; err != nil && err != gorm.ErrRecordNotFound {
			panic(err)
		} else {
			c.DbUser = &user
		}

		// Exist and ban check
		if c.DbUser != nil && c.DbUser.Privileges&privileges.Normal == 0 {
			c.Report("Non hai i permessi necessari per utilizzare questo bot.")
			return
		}

		f(c)
	}
}

// handleErrors is the base error handler.
// it recovers from any panic, logs the stack to stdout
// and reports feedback to the user
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
				c.Report(text.ErrorOccurred)
			}
		}()
		f(c)
	}
}

// protected runs the handler only if the user has at least the
// specified privileges, otherwise it does nothing
func protected(f CommandHandler, requiredPrivileges privileges.Privileges) CommandHandler {
	return func(c *common.Ctx) {
		if c.DbUser == nil || c.DbUser.Privileges&requiredPrivileges == 0 {
			log.Printf("%v (%v) does not have the required privileges (%v) to trigger %v", c.DbUser.TelegramID, c.DbUser.Privileges, requiredPrivileges, f)
			return
		}
		f(c)
	}
}

func guestsOnly(f CommandHandler, requiredPrivileges privileges.Privileges) CommandHandler {
	return func(c *common.Ctx) {
		if c.DbUser != nil {
			log.Printf("%v tried to trigger guest-only handler %v", c.DbUser.TelegramID, f)
			return
		}
		f(c)
	}
}

// fsm runs the handler only if the user is in the specified
// state, otherwise it does nothing
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

// Handler is a telegram handler
type Handler struct {
	// Function to run.
	// Must be wrapped with BaseWrap, BaseWrapCb, etc
	F CommandHandler

	// If not zero, add a protected() decorator
	// with the provided privileges
	// Set to -1 to allow this only for guests
	P privileges.Privileges

	// If not zero, add a fsm() decorator
	// with the provided state
	S string
}

func (h Handler) wrap() {
	if h.P > 0 {
		h.F = protected(h.F, h.P)
	} else if h.P <= 0 {
		h.F = guestsOnly(h.F, h.P)
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
	return wrapCtxCallback(handleErrors(resolveUser(h.F)))
}

func (h Handler) BaseWrapQ() func(*tb.Query) {
	h.wrap()
	return wrapCtxQuery(handleErrors(resolveUser(h.F)))
}

func (h Handler) TextWrap() func(*common.Ctx) {
	h.wrap()
	return h.F
}
