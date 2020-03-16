package models

// Area represents an area where deliveries can occur
type Area struct {
	ID      int
	Name    string
	Visible bool `gorm:"default:'true'"`
}

// TableName returns the sql table name
func (Area) TableName() string {
	return "areas"
}
