package models

import "github.com/jinzhu/gorm"

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

// GetAreaByName returns an Area pointer by its exact name, nil otherwise
func GetAreaByName(db *gorm.DB, areaName string) (*Area, error) {
	var area Area
	if err := db.Where("name = ?", areaName).First(&area).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No area
			return nil, nil
		}
		// Other err
		return nil, err
	}
	return &area, nil
}

// GetVisibleAreas returns all visible areas, as a slice
func GetVisibleAreas(db *gorm.DB) ([]Area, error) {
	var areas []Area
	if err := db.Where("visible = 1").Find(&areas).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []Area{}, nil
		}
		return nil, err
	}
	return areas, nil
}
