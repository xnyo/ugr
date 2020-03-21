package models

// Photo represents a Photo
type Photo struct {
	FileID  string `gorm:"type:varchar(48);primary_key"`
	OrderID uint
}
