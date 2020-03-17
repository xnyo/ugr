package models

// Area represents an area where deliveries can occur
type Area struct {
	ID      uint
	Name    string
	Visible bool `gorm:"default:'true'"`
	Orders  []Order
}

// TableName returns the sql table name
func (Area) TableName() string {
	return "areas"
}

func (a Area) String() string {
	return a.Name
}
