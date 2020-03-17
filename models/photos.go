package models

// Photo represents a Photo
type Photo struct {
	FileID  string `gorm:"primary_key"`
	OrderID uint
}
