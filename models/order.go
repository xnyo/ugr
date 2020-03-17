package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// OrderStatus is the status of an order
type OrderStatus uint8

// Order statuses
const (
	OrderStatusNeedsData OrderStatus = iota
	OrderStatusPending
	OrderStatusTaken
	OrderStatusDone
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusNeedsData:
		return "in attesa di dati"
	case OrderStatusPending:
		return "in attesa"
	case OrderStatusTaken:
		return "assegnato"
	case OrderStatusDone:
		return "completato"
	}
	return ""
}

// Order represents an order
type Order struct {
	gorm.Model
	Name       string
	Address    string
	Telephone  string
	AreaID     *uint
	Status     OrderStatus `gorm:"default:0"`
	AssignedTo User
	Notes      *string `gorm:"size:512"`
	Photos     []Photo
}

// TableName returns the sql table name
func (Order) TableName() string {
	return "orders"
}

// ToTelegram converts the current Order to a summary string
// for the telegram bot
func (o Order) ToTelegram(db *gorm.DB) (string, error) {
	var notes string
	if o.Notes == nil {
		notes = "Nessuna"
	} else {
		notes = *o.Notes
	}
	var area Area
	if err := db.Where("id = ?", o.AreaID).First(&area).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"🔸 Ordine per: %s\n🔸 Indirizzo: %s\n🔸 Zona: %s\n🔸 Telefono: %s\n🔸 Stato: %s\n🔸 Note:\n<code>%s</code>",
		o.Name,
		o.Address,
		area.Name,
		o.Telephone,
		o.Status.String(),
		notes,
	), nil
}
