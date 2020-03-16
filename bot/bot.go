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

// Db is the gorm db reference
var Db *gorm.DB

// HandleText registers a new raw text handler
// The handler is a CommandHandler, so it already
// has its Ctx already set. This is because
// the raw dispatcher sets it (because it requires)
// it as well, to determine the FSM status of the
// current telegram user.
func HandleText(state string, f CommandHandler) {
	TextHandlers[state] = f
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

	// Admin handlers
	B.Handle("/admin", Handler{F: admin.Menu, P: privileges.Admin}.BaseWrap())
	B.Handle("\fadmin", Handler{F: admin.Menu, P: privileges.Admin}.BaseWrapCb())

	// Admin callback queries
	B.Handle(
		"\fadmin__add_area",
		Handler{
			F: textPrompt(
				"admin/add_area",
				"Indica il nome della zona da aggiungere",
				admin.BackReplyMarkup,
			),
			P: privileges.Admin,
			S: "admin",
		}.BaseWrapCb(),
	)
	B.Handle("\fadmin__areas", Handler{F: admin.Areas, P: privileges.Admin, S: "admin"}.BaseWrapCb())
	B.Handle("\fadmin__add_admin", notAvailable)
	B.Handle("\fadmin__remove_admin", notAvailable)
	B.Handle("\fadmin__ban", notAvailable)
	B.Handle("\fadmin__users", notAvailable)

	// Raw text dispatcher (multi-stage states)
	B.Handle(tb.OnText, Handler{
		F: func(c *common.Ctx) {
			if v, ok := TextHandlers[c.DbUser.State]; ok {
				v(c)
			} else {
				log.Printf("Unbound state %v in text dispatcher\n", c.DbUser.State)
			}
		},
	}.BaseWrap())

	// Raw text handlers
	HandleText("admin/add_area", Handler{F: admin.AddAreaName, P: privileges.Admin, S: "admin"}.TextWrap())

	// Start the bot (blocks the current goroutine)
	log.Println("UGR")
	defer log.Println("Disposing bot")
	B.Start()
}
