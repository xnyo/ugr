package common

import tb "gopkg.in/tucnak/telebot.v2"

// SingleKeyboardFactory returns a ReplyMarkup with a single button
// containing the provided text
func SingleKeyboardFactory(s string) *tb.ReplyMarkup {
	return &tb.ReplyMarkup{
		ReplyKeyboard: [][]tb.ReplyButton{
			[]tb.ReplyButton{
				tb.ReplyButton{
					Text: s,
				},
			},
		},
	}
}
