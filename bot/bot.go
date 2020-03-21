package bot

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/xnyo/ugr/text"

	"github.com/xnyo/ugr/commands/admin"
	"github.com/xnyo/ugr/commands/volunteer"
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
func NewCtx(m *tb.Message, cb *tb.Callback, q *tb.Query) common.Ctx {
	if m == nil && cb != nil {
		m = cb.Message
	}
	return common.Ctx{
		B:            B,
		Db:           Db,
		Message:      m,
		Callback:     cb,
		InlineQuery:  q,
		LogChannelID: logChannelID,
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

var logChannelID string = "-1001449367422"

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

func rawTextHandle(dispatcher map[string]CommandHandler) CommandHandler {
	return func(c *common.Ctx) {
		if c.DbUser != nil && strings.TrimSpace(c.Message.Text) == text.MainMenu {
			// "Go back to main menu" pre-handler
			c.B.Delete(c.Message)
			if strings.HasPrefix(c.DbUser.State, "admin") {
				admin.Menu(c)
			} else if strings.HasPrefix(c.DbUser.State, "volunteer") {
				volunteer.Menu(c)
			}
		} else {
			// Normal dispatcher
			rawDispatch(dispatcher)(c)
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
	Db.LogMode(true)
	Db.AutoMigrate(models.All...)

	// Dummy handlers
	B.Handle("/start", Handler{F: volunteer.Menu, P: privileges.Normal}.BaseWrap())
	B.Handle("\fvolunteer", Handler{F: volunteer.Menu, P: privileges.Normal}.BaseWrapCb())

	// Admin -- menu
	B.Handle("/admin", Handler{F: admin.Menu, P: privileges.Admin}.BaseWrap())
	B.Handle("\fadmin", Handler{F: admin.Menu, P: privileges.Admin}.BaseWrapCb())
	B.Handle(
		"\fdelete_self",
		Handler{
			F: func(c *common.Ctx) {
				c.Respond(&tb.CallbackResponse{Text: "👋 Arrivederci!"})
				c.B.Delete(c.Message)
			},
		}.BaseWrapCb(),
	)
	B.Handle(
		"\fn",
		Handler{
			F: func(c *common.Ctx) {
				// TODO: consider moving album message ids in db,
				// there's a 64 characters limit in the callback query
				messages := strings.Split(c.Callback.Data, "|")
				for _, v := range messages {
					id, err := strconv.Atoi(v)
					if err != nil {
						continue
					}
					c.B.Delete(&tb.Message{ID: id, Chat: c.Message.Chat})

					// So we avoid spamming telegram...
					time.Sleep(time.Millisecond * 500)
				}
				c.Respond(&tb.CallbackResponse{Text: "🗑"})
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
				true,
				admin.BackReplyMarkup,
			),
			P: privileges.Admin,
			S: "admin",
		}.BaseWrapCb(),
	)
	B.Handle("\fadmin__areas", Handler{F: admin.Areas, P: privileges.Admin, S: "admin"}.BaseWrapCb())

	// Admin -- TODO
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
				true,
				admin.BackReplyMarkup,
				tb.ModeHTML,
			),
			P: privileges.Admin,
			S: "admin",
		}.BaseWrapCb(),
	)
	B.Handle("\fadmin__add_order__no_expire", Handler{F: admin.AddOrderNoExpire, P: privileges.Admin, S: "admin/add_order/expire"}.BaseWrapCb())

	// Invite response
	B.Handle("\faccept_admin", Handler{F: commands.AcceptAdminInvite}.BaseWrapCb())
	B.Handle("\faccept_volunteer", Handler{F: commands.AcceptVolunteerInvite}.BaseWrapCb())

	// Admin -- raw text handlers (data input)
	HandleText(Handler{F: admin.AddAreaName, P: privileges.Admin, S: "admin/add_area"})
	HandleText(Handler{F: admin.AddOrderData, P: privileges.Admin, S: "admin/add_order"})
	HandleText(Handler{F: admin.AddOrderArea, P: privileges.Admin, S: "admin/add_order/area"})
	HandleText(Handler{F: admin.AddOrderExpire, P: privileges.Admin, S: "admin/add_order/expire"})
	HandleText(Handler{F: admin.AddOrderAttachmentsEnd, P: privileges.Admin, S: "admin/add_order/attachments"})

	// Admin -- raw photo handlers
	HandlePhoto(Handler{F: admin.AddOrderAttachments, P: privileges.Admin, S: "admin/add_order/attachments"})

	// User -- new order
	B.Handle("\fuser__choose", Handler{F: volunteer.ChooseOrderStart, P: privileges.Normal, S: "volunteer"}.BaseWrapCb())
	HandleText(Handler{F: volunteer.ChooseOrderZone, P: privileges.Normal, S: "volunteer/choose/zone"})
	B.Handle("\fuser__choose_next", Handler{F: volunteer.ChooseNext, P: privileges.Normal, S: "volunteer/choose/order"}.BaseWrapCb())
	B.Handle("\fuser__choose_previous", Handler{F: volunteer.ChoosePrevious, P: privileges.Normal, S: "volunteer/choose/order"}.BaseWrapCb())
	B.Handle("\fuser__choose_confirm", Handler{F: volunteer.ChooseConfirm, P: privileges.Normal, S: "volunteer/choose/order"}.BaseWrapCb())

	// User -- my orders
	B.Handle("\fuser__my_orders", Handler{F: volunteer.MyOrders, P: privileges.Normal, SS: []string{"volunteer", "volunteer/my"}}.BaseWrapCb())
	B.Handle("\fuser__my_previous", Handler{F: volunteer.MyPrevious, P: privileges.Normal, S: "volunteer/my"}.BaseWrapCb())
	B.Handle("\fuser__my_next", Handler{F: volunteer.MyNext, P: privileges.Normal, S: "volunteer/my"}.BaseWrapCb())
	B.Handle("\fuser__my_done", Handler{F: volunteer.MyDone, P: privileges.Normal, S: "volunteer/my"}.BaseWrapCb())
	B.Handle("\fuser__my_cancel", Handler{F: volunteer.MyCancel, P: privileges.Normal, S: "volunteer/my"}.BaseWrapCb())
	B.Handle("\fuser__my_photos", Handler{F: volunteer.MyPhotos, P: privileges.Normal, S: "volunteer/my"}.BaseWrapCb())

	// Inline handler (invites)
	B.Handle(tb.OnQuery, Handler{F: admin.InlineInviteHandler}.BaseWrapQ())

	// Raw text dispatcher (multi-stage states)
	B.Handle(tb.OnText, Handler{F: rawTextHandle(TextHandlers)}.BaseWrap())
	B.Handle(tb.OnPhoto, Handler{F: rawDispatch(PhotoHandlers)}.BaseWrap())

	// Start the bot (blocks the current goroutine)
	log.Println("UGR")
	defer log.Println("Disposing bot")
	B.Start()
}
