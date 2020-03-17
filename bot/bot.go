package bot

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/xnyo/ugr/commands/admin"
	"github.com/xnyo/ugr/privileges"

	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/commands"
	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"

	// Register sqlite gorm dialect
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	tb "gopkg.in/tucnak/telebot.v2"
)

// NewCtx creates a new Ctx from a tb.Message pointer
func NewCtx(m *tb.Message, cb *tb.Callback) common.Ctx {
	if m == nil && cb != nil {
		m = cb.Message
	}
	return common.Ctx{
		B:        B,
		Db:       Db,
		Message:  m,
		Callback: cb,
	}
}

// B is the telebot bot reference
var B *tb.Bot

// TextHandlers contains the handlers for tb.OnText
var TextHandlers map[string]CommandHandler = make(map[string]CommandHandler)

// PhotoHandlers contains the handlers for tb.OnPhoto
var PhotoHandlers map[string]CommandHandler = make(map[string]CommandHandler)

// Db is the gorm db reference
var Db *gorm.DB

// HandleText registers a new raw text handler
// The handler is a CommandHandler, so it already
// has its Ctx already set. This is because
// the raw dispatcher sets it (because it requires)
// it as well, to determine the FSM status of the
// current telegram user.
func HandleText(h Handler) {
	TextHandlers[h.S] = h.TextWrap()
}

// HandlePhoto registers a new raw photo handler
func HandlePhoto(h Handler) {
	PhotoHandlers[h.S] = h.TextWrap()
}

// rawDispatch dispatches messages based on the state of the user
// held by the current ctx, based on a provided dispatcher map
func rawDispatch(dispatcher map[string]CommandHandler) CommandHandler {
	return func(c *common.Ctx) {
		if v, ok := dispatcher[c.DbUser.State]; ok {
			v(c)
		} else {
			log.Printf("Unbound state %v in text dispatcher\n", c.DbUser.State)
		}
	}
}

// notAvailable is a raw telebot handler that
// responds to the callback query with a
// "feature not available" message.
func notAvailable(c *tb.Callback) {
	B.Respond(c, &tb.CallbackResponse{
		Text:      "😔 Funzionalità non ancora disponibile.",
		ShowAlert: true,
	})
}

// Initialize initizliaes the bot
func Initialize(token string) error {
	if B != nil {
		return errors.New("Bot already initialized")
	}
	var err error
	B, err = tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	return err
}

// Start starts the bot
func Start() {
	defer log.Println("Bot disposed, bye!")

	// SIGINT handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			B.Stop()
		}
	}()

	// Db
	var err error
	Db, err = gorm.Open("sqlite3", "ugr.db")
	if err != nil {
		log.Fatal(err)
	}
	defer Db.Close()
	Db.AutoMigrate(models.All...)

	// Dummy handlers
	B.Handle("/start", Handler{F: commands.Start}.BaseWrap())

	// Admin -- menu
	B.Handle("/admin", Handler{F: admin.Menu, P: privileges.Normal}.BaseWrap())
	B.Handle("\fadmin", Handler{F: admin.Menu, P: privileges.Normal}.BaseWrapCb())
	B.Handle(
		"\fdelete_self",
		Handler{
			F: func(c *common.Ctx) {
				c.Respond(&tb.CallbackResponse{Text: "👋 Arrivederci!"})
				c.B.Delete(c.Message)
			},
		}.BaseWrapCb(),
	)

	// Admin -- areas
	B.Handle(
		"\fadmin__add_area",
		Handler{
			F: textPrompt(
				"Indica il nome della zona da aggiungere",
				"admin/add_area",
				admin.BackReplyMarkup,
			),
			P: privileges.Normal,
			S: "admin",
		}.BaseWrapCb(),
	)
	B.Handle("\fadmin__areas", Handler{F: admin.Areas, P: privileges.Normal, S: "admin"}.BaseWrapCb())

	// Admin -- TODO
	B.Handle("\fadmin__add_admin", notAvailable)
	B.Handle("\fadmin__remove_admin", notAvailable)
	B.Handle("\fadmin__ban", notAvailable)
	B.Handle("\fadmin__users", notAvailable)

	// Admin -- orders
	B.Handle(
		"\fadmin__add_order",
		Handler{
			F: textPrompt(
				`<b>Invia un messaggio con i seguenti dati, uno per riga:</b>

<code>cognome (e nome) destinatario
indirizzo destinatario
numero di telefono destinatario
note aggiuntive (anche più righe)</code>`,
				"admin/add_order",
				admin.BackReplyMarkup,
				tb.ModeHTML,
			),
			P: privileges.Normal,
			S: "admin",
		}.BaseWrapCb(),
	)
	B.Handle("\fadmin__add_order__no_expire", Handler{F: admin.AddOrderNoExpire, P: privileges.Normal, S: "admin/add_order/expire"}.BaseWrapCb())

	// Admin -- raw text handlers (data input)
	HandleText(Handler{F: admin.AddAreaName, P: privileges.Normal, S: "admin/add_area"})
	HandleText(Handler{F: admin.AddOrderData, P: privileges.Normal, S: "admin/add_order"})
	HandleText(Handler{F: admin.AddOrderArea, P: privileges.Normal, S: "admin/add_order/area"})
	HandleText(Handler{F: admin.AddOrderExpire, P: privileges.Normal, S: "admin/add_order/expire"})
	HandleText(Handler{F: admin.AddOrderAttachmentsEnd, P: privileges.Normal, S: "admin/add_order/attachments"})

	// Admin -- raw photo handlers
	HandlePhoto(Handler{F: admin.AddOrderAttachments, P: privileges.Normal, S: "admin/add_order/attachments"})

	// Raw text dispatcher (multi-stage states)
	B.Handle(tb.OnText, Handler{F: rawDispatch(TextHandlers)}.BaseWrap())
	B.Handle(tb.OnPhoto, Handler{F: rawDispatch(PhotoHandlers)}.BaseWrap())

	// Start the bot (blocks the current goroutine)
	log.Println("UGR")
	defer log.Println("Disposing bot")
	B.Start()
}
