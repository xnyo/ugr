package models

import (
	"github.com/jinzhu/gorm"
)

// OrderStatus is the status of an order
type OrderStatus uint8

// Order statuses
const (
	OrderStatusPending OrderStatus = iota
	OrderStatusPendingTaken
	OrderStatusPendingDelivering
	OrderStatusPendingDone
)

// Order represents an order
type Order struct {
	gorm.Model
	Name        string
	Address     string
	Telephone   string
	Area        Area
	Status      OrderStatus
	AssignedTo  User
	GroceryList []Groceries
}

type Groceries struct {
	Name    string
	Done    bool
	OrderID uint
}

// TableName returns the sql table name
func (Order) TableName() string {
	return "orders"
}
