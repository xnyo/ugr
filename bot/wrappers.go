package bot

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"

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

// handleErrorsBase is the base error handler.
// it recovers from any panic, logs the stack to stdout
// and calls a custom handler (h), which can be used to
// report feedback to the user, for example by sending
// them a message or answering to their callback query
func handleErrorsBase(f CommandHandler, h func(*common.Ctx, error)) CommandHandler {
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
				h(c, err)
			}
		}()
		f(c)
	}
}

// handleErrors returns the base error handler
// with a custom handler that sends a message
// to the user if an error occurs
func handleErrors(f CommandHandler) CommandHandler {
	return handleErrorsBase(
		f,
		func(c *common.Ctx, err error) {
			c.Reply(text.ErrorOccurred)
		},
	)
}

// handleErrorsCb returns the base error handler
// with a custom handler that responds to the
// callback query with an erro message if an error occurs
func handleErrorsCb(f CommandHandler) CommandHandler {
	return handleErrorsBase(
		f,
		func(c *common.Ctx, err error) {
			c.Respond(&tb.CallbackResponse{Text: text.ErrorOccurred, ShowAlert: true})
		},
	)
}

// protected runs the handler only if the user has at least the
// specified privileges, otherwise it does nothing
func protected(f CommandHandler, requiredPrivileges privileges.Privileges) CommandHandler {
	return func(c *common.Ctx) {
		if c.DbUser.Privileges&requiredPrivileges == 0 {
			log.Printf("%v (%v) does not have the required privileges (%v) to trigger %v", c.DbUser.TelegramID, c.DbUser.Privileges, requiredPrivileges, f)
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
	P privileges.Privileges

	// If not zero, add a fsm() decorator
	// with the provided state
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
	return wrapCtxCallback(handleErrorsCb(resolveUser(h.F)))
}

func (h Handler) BaseWrapQ() func(*tb.Query) {
	h.wrap()
	return wrapCtxQuery(handleErrors(resolveUser(h.F)))
}

func (h Handler) TextWrap() func(*common.Ctx) {
	h.wrap()
	return h.F
}
