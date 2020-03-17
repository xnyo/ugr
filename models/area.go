package models

// Area represents an area where deliveries can occur
type Area struct {
	ID      uint
	Name    string
	Visible bool `gorm:"default:'true'"`
	Orders  []Order
}

func (a Area) String() string {
	return a.Name
}
