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

// Db is the gorm db reference
var Db *gorm.DB

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
	/*B.Handle(
		"/aggordine", base(
			protected(commands.AdminNewOrder, privileges.AdminAddOrder),
		),
	)*/
	B.Handle(
		"/admin", base(
			protected(commands.AdminMenu, privileges.Admin),
		),
	)
	B.Handle(
		"/aggarea", base(
			protected(commands.AdminNewArea, privileges.AdminAddArea),
		),
	)
	B.Handle(
		"\fTest", baseCallback(
			protected(func(c *common.Ctx) {
				c.Answer(&tb.CallbackResponse{
					Text: "Ok!",
				})
			}, privileges.Admin),
		),
	)

	// Start the bot (blocks the current goroutine)
	log.Println("UGR")
	defer log.Println("Disposing bot")
	B.Start()
}
