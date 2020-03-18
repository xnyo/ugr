package admin

import (
	"fmt"
	"strings"

	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	"github.com/xnyo/ugr/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

// AddAreaName asks for the name of the area that will be added
func AddAreaName(c *common.Ctx) {
	if strings.ContainsRune(c.Message.Text, '\n') {
		c.Reply(text.W("Il nome della zona deve essere contenuto in una sola riga."), tb.ModeMarkdown)
		return
	}
	s := strings.TrimSpace(c.Message.Text)
	c.Db.Create(&models.Area{
		Name:    s,
		Visible: true,
	})
	c.UpdateMenu("✅ **Zona aggiunta**", BackReplyMarkup, tb.ModeMarkdown)
}

// Areas returns the list of all available areas
func Areas(c *common.Ctx) {
	s := "🌆 **Zone disponibili:**\n"
	var results []models.Area
	c.Db.Find(&results)
	for _, v := range results {
		var visible string
		if v.Visible {
			visible = "👁"
		} else {
			visible = "❌"
		}
		s += fmt.Sprintf("\n🔸 %s (%s)", v.String(), visible)
	}
	c.UpdateMenu(s, BackReplyMarkup, tb.ModeMarkdown)
}
