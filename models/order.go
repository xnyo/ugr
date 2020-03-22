package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// OrderStatus is the status of an order
type OrderStatus uint8

// Order statuses
const (
	OrderStatusNeedsData OrderStatus = iota
	OrderStatusPending
	OrderStatusDone
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusNeedsData:
		return "in attesa di dati"
	case OrderStatusPending:
		return "in attesa"
	case OrderStatusDone:
		return "completato"
	}
	return ""
}

// Order represents an order
type Order struct {
	gorm.Model
	Name           string
	Address        string
	Telephone      string
	AreaID         *uint
	Expire         *time.Time
	Status         OrderStatus `gorm:"default:0"`
	AssignedUserID *uint
	Notes          *string `gorm:"size:512"`
	Photos         []Photo
}

// ToTelegram converts the current Order to a summary string
// for the telegram bot
// "what" can be either:
// - a *gorm.DB, in that case the area will be fetched from the db
// - a (*)models.Area, in case the area has already been fetched
func (o Order) ToTelegram(dbOrArea interface{}) (string, error) {
	// Fetch area from db or get it from the parameter
	var a Area
	switch dbOrArea.(type) {
	case *gorm.DB:
		{
			db := dbOrArea.(*gorm.DB)
			if err := db.Where("id = ?", o.AreaID).First(&a).Error; err != nil {
				return "", err
			}
		}
	case Area:
		a = dbOrArea.(Area)
	case *Area:
		a = *dbOrArea.(*Area)
	default:
		return "", fmt.Errorf("unsupported type supplied: %T. Expected *gorm.DB, models.Area or a ptr to it", dbOrArea)
	}

	// Format notes (nullable)
	var notes string
	if o.Notes == nil {
		notes = "Nessuna"
	} else {
		notes = *o.Notes
	}

	// Fetch expire (nullable)
	var expire, expiredWarning string
	if o.Expire == nil {
		expire = "Nessuna"
	} else {
		expire = o.Expire.In(time.Local).Format("02/01/2006 15:04")
		if o.Expire.Before(time.Now().UTC()) {
			expiredWarning = "⚠️ <b>Attenzione! Questo ordine è scaduto.</b>\n\n"
		}
	}

	// Format string
	return fmt.Sprintf(
		"%s🔸 Ordine per: %s\n🔸 Indirizzo: %s\n🔸 Zona: %s\n🔸 Telefono: %s\n🔸 Scadenza: %s\n🔸 Stato: %s\n🔸 Note:\n<code>%s</code>",
		expiredWarning,
		o.Name,
		o.Address,
		a.Name,
		o.Telephone,
		expire,
		o.Status.String(),
		notes,
	), nil
}
