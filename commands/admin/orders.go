package admin

import (
	"strings"

	"github.com/xnyo/ugr/common"
	"github.com/xnyo/ugr/models"
	tb "gopkg.in/tucnak/telebot.v2"
)

// AddOrderData asks for all required data of the order that will be added
func AddOrderData(c *common.Ctx) {
	parts := strings.SplitN(c.Message.Text, "\n", 4)
	if len(parts) < 4 {
		c.Reply("⚠️ **Non hai specificato abbastanza tutti i dati richiesti!**", tb.ModeMarkdown)
		return
	}
	name, address, phone, groceries := parts[0], parts[1], parts[2], parts[3]

	// Add groceries
	var groceryList []models.Groceries
	for _, v := range strings.Split(groceries, "\n") {
		v = strings.TrimSpace(v)
		groceryList = append(groceryList, models.Groceries{
			Name: v,
			Done: false,
		})
	}
	// Add order & groceries
	if err := c.Db.Create(&models.Order{
		Name:        name,
		Address:     address,
		Telephone:   phone,
		GroceryList: groceryList,
	}).Error; err != nil {
		panic(err)
	}
	c.UpdateMenu("✅ **Ordine aggiunto**", BackReplyMarkup, tb.ModeMarkdown)
}
