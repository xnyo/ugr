package common

import tb "gopkg.in/tucnak/telebot.v2"

// SingleKeyboardFactory returns a ReplyMarkup with a single button
// containing the provided text
func SingleKeyboardFactory(s string) *tb.ReplyMarkup {
	return &tb.ReplyMarkup{
		ReplyKeyboard: [][]tb.ReplyButton{
			{{Text: s}},
		},
	}
}

// SingleInlineKeyboardFactory returns a ReplyMarkup with a single button
// containing the provided button (inline version)
func SingleInlineKeyboardFactory(btn tb.InlineButton) *tb.ReplyMarkup {
	return &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{{btn}},
	}
}
