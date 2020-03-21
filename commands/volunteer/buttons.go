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

func myOrdersPrevious(orderID int) tb.InlineButton {
	return tb.InlineButton{
		Text:   "⬅️",
		Unique: fmt.Sprintf("user__my_previous|%d", orderID),
	}
}

func myOrdersNext(orderID int) tb.InlineButton {
	return tb.InlineButton{
		Text:   "➡️",
		Unique: fmt.Sprintf("user__my_next|%d", orderID),
	}
}

func myOrdersKeyboard(orderID int, hasPrevious, hasNext bool, photos int) *tb.ReplyMarkup {
	k := &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{},
			{
				{Text: "✅ Completato", Unique: fmt.Sprintf("user__my_done|%d|0", orderID)},
				{Text: "😞 Rinuncia", Unique: fmt.Sprintf("user__my_cancel|%d|0", orderID)},
			},
			{BackReplyButton},
		},
	}
	if hasPrevious {
		k.InlineKeyboard[0] = append(k.InlineKeyboard[0], myOrdersPrevious(orderID))
	}
	if hasNext {
		k.InlineKeyboard[0] = append(k.InlineKeyboard[0], myOrdersNext(orderID))
	}
	if photos > 0 {
		k.InlineKeyboard = append(
			k.InlineKeyboard[:1],
			append(
				[][]tb.InlineButton{
					{
						{
							Text:   fmt.Sprintf("📷 Visualizza foto (%d)", photos),
							Unique: fmt.Sprintf("user__my_photos|%d", orderID),
						},
					},
				},
				k.InlineKeyboard[1:]...,
			)...,
		)
	}
	if !hasPrevious && !hasNext {
		k.InlineKeyboard = k.InlineKeyboard[1:]
	}
	return k
}
