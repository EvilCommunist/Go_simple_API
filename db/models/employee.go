package models

import (
	"time"
)

type Employee struct {
	ID           int        `gorm:"primaryKey"`
	DepartmentID int        `gorm:"not null;index"`
	Department   Department `gorm:"foreignKey:DepartmentID"`
	FullName     string     `gorm:"size:200;not null"`
	Position     string     `gorm:"size:200;not null"`
	HiredAt      *time.Time `gorm:"type:date"`
	CreatedAt    time.Time
}

func (Employee) TableName() string {
	return "employees"
}
