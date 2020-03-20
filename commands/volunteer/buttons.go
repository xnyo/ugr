package volunteer

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

func chooseOrderPrevious(areaID, orderID int) tb.InlineButton {
	return tb.InlineButton{
		Text:   "⬅️",
		Unique: fmt.Sprintf("user__choose_previous|%d|%d", areaID, orderID),
	}
}

func chooseOrderNext(areaID, orderID int) tb.InlineButton {
	return tb.InlineButton{
		Text:   "➡️",
		Unique: fmt.Sprintf("user__choose_next|%d|%d", areaID, orderID),
	}
}

func chooseOrderConfirm(orderID int) tb.InlineButton {
	return tb.InlineButton{
		Text:   "✅",
		Unique: fmt.Sprintf("user__choose_confirm|%d", orderID),
	}
}
