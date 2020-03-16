package bot

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

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

func handleText(state string, f CommandHandler) {
	TextHandlers[state] = f
}

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

	// Register handlers
	B.Handle("/start", base(commands.Start))

	// Admin handlers
	B.Handle("/admin", base(protected(commands.AdminMenu, privileges.Admin)))
	B.Handle("\fadmin", baseCallback(protected(commands.AdminMenu, privileges.Admin)))

	// Admin callback queries
	B.Handle(
		"\fadmin__add_area", baseCallback(
			protected(
				fsm(
					textPrompt(
						"admin/add_area",
						"Indica il nome della zona da aggiungere",
						commands.AdminBackReplyMarkup,
					),
					"admin",
				),
				privileges.Admin,
			),
		),
	)
	handleText("admin/add_area", protected(commands.AdminAddAreaName, privileges.Admin))
	B.Handle("\fadmin__areas", baseCallback(protected(commands.AdminAreas, privileges.Admin)))
	B.Handle("\fadmin__add_admin", notAvailable)
	B.Handle("\fadmin__remove_admin", notAvailable)
	B.Handle("\fadmin__ban", notAvailable)
	B.Handle("\fadmin__users", notAvailable)

	// Text dispatcher
	B.Handle(tb.OnText, base(func(c *common.Ctx) {
		if v, ok := TextHandlers[c.DbUser.State]; ok {
			v(c)
		} else {
			log.Printf("Unbound state %v in text dispatcher\n", c.DbUser.State)
		}
	}))

	// Start the bot (blocks the current goroutine)
	log.Println("UGR")
	defer log.Println("Disposing bot")
	B.Start()
}
