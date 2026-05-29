package models

import "time"

type Department struct {
	ID        int          `gorm:"PrimaryKey"`
	Name      string       `gorm:"size:200;not null"`
	ParentID  *int         `gorm:"index"`
	Parent    *Department  `gorm:"foreignKey:ParentID"`
	Children  []Department `gorm:"foreignKey:ParentID"`
	Employees []Employee
	CreatedAt time.Time
}

func (Department) TableName() string {
	return "departments"
}
