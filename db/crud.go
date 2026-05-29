package db

import (
	"API/db/models"
	"errors"

	"gorm.io/gorm"
)

type Database struct {
	conn *gorm.DB
}

func (db *Database) SetDatabase(connection *gorm.DB) {
	db.conn = connection
}

// Department

func (db *Database) CreateDepartment(dep models.Department, parent *int) (*models.Department, error) {
	if len(dep.Name) > 200 {
		return nil, errors.New("Name is longer than 200 symbols")
	}

	var count int64
	query := db.conn.Model(&models.Department{}).Where("name = ?", dep.Name)
	if dep.ParentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *dep.ParentID)
	}
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("Same department already exists with this parent")
	}

	dept := &models.Department{
		Name:     dep.Name,
		ParentID: dep.ParentID,
	}
	if err := db.conn.Create(dept).Error; err != nil {
		return nil, err
	}
	return dept, nil
}

func (db *Database) UpdateDepartment() {

}

// Employees

func (db *Database) CreateEmployee(emp models.Employee, department int) (*models.Employee, error) {
	var dept models.Department
	if err := db.conn.First(&dept, department).Error; err != nil {
		return nil, errors.New("department not found")
	}
	if len(emp.FullName) > 200 {
		return nil, errors.New("Full name is longer than 200 symbols")
	}
	if len(emp.Position) > 200 {
		return nil, errors.New("Position is longer than 200 symbols")
	}

	empl := &models.Employee{
		DepartmentID: department,
		FullName:     emp.FullName,
		Position:     emp.Position,
		HiredAt:      emp.HiredAt,
	}
	if err := db.conn.Create(emp).Error; err != nil {
		return nil, err
	}
	return empl, nil
}
