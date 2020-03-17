package models

import (
	"github.com/jinzhu/gorm"
)

// OrderStatus is the status of an order
type OrderStatus uint8

// Order statuses
const (
	OrderStatusNeedsData OrderStatus = iota
	OrderStatusPending
	OrderStatusTaken
	OrderStatusDelivering
	OrderStatusDone
)

// Order represents an order
type Order struct {
	gorm.Model
	Name       string
	Address    string
	Telephone  string
	Area       *Area
	Status     OrderStatus `gorm:"default:0"`
	AssignedTo *User
	Notes      *string `gorm:"size:512"`
	Photos     []Photo
}

// TableName returns the sql table name
func (Order) TableName() string {
	return "orders"
}
