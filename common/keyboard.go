package common

import (
	"github.com/jinzhu/gorm"
	"github.com/xnyo/ugr/models"
	tb "gopkg.in/tucnak/telebot.v2"
)

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

// AreasReplyKeyboard returns a reply keyboard with all visible areas
func AreasReplyKeyboard(db *gorm.DB) ([][]tb.ReplyButton, error) {
	areas, err := models.GetVisibleAreas(db)
	if err != nil {
		return nil, err
	}
	var keyboard [][]tb.ReplyButton
	for _, v := range areas {
		keyboard = append(keyboard, []tb.ReplyButton{tb.ReplyButton{Text: v.Name}})
	}
	return keyboard, nil
}
