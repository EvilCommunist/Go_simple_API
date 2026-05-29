package db

import (
	"API/db/models"
	"errors"
	"strings"

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
	if strings.TrimSpace(dep.Name) == "" {
		return nil, errors.New("name cannot be empty")
	}
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

func (db *Database) UpdateDepartment(depID int, name *string, parentID *int) (*models.Department, error) {
	var dept models.Department
	if err := db.conn.First(&dept, depID).Error; err != nil {
		return nil, errors.New("department not found")
	}

	if name != nil {
		if *name == "" {
			return nil, errors.New("name cannot be empty")
		}
		if len(*name) > 200 {
			return nil, errors.New("name too long (max 200)")
		}

		var effectiveParentID *int = dept.ParentID
		if parentID != nil && *parentID != 0 {
			effectiveParentID = parentID
		} else if parentID != nil && *parentID == 0 {
			effectiveParentID = nil
		}

		var count int64
		query := db.conn.Model(&models.Department{}).Where("name = ? AND id != ?", *name, dept.ID)
		if effectiveParentID == nil {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", *effectiveParentID)
		}
		if err := query.Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("department with this name already exists under the target parent")
		}
		dept.Name = *name
	}

	if parentID != nil {
		if *parentID == depID {
			return nil, errors.New("department cannot be its own parent")
		}
		if *parentID == 0 {
			dept.ParentID = nil
		} else {
			var parent models.Department
			if err := db.conn.First(&parent, *parentID).Error; err != nil {
				return nil, errors.New("parent department not found")
			}
			if db.hasCycle(dept.ID, *parentID) {
				return nil, errors.New("cannot move department under its own descendant")
			}
			dept.ParentID = &parent.ID
		}
	}

	if err := db.conn.Save(&dept).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}

func (db *Database) hasCycle(deptID, newParentID int) bool {
	current := newParentID
	for {
		if current == deptID {
			return true
		}
		var parentID *int
		err := db.conn.Model(&models.Department{}).
			Select("parent_id").
			Where("id = ?", current).
			Scan(&parentID).Error
		if err != nil || parentID == nil {
			break
		}
		current = *parentID
	}
	return false
}

func (db *Database) GetAllChildrenIDs(rootID int) ([]int, error) {
	var ids []int
	err := db.conn.Raw(`
		WITH RECURSIVE subtree AS (
			SELECT id FROM departments WHERE id = ?
			UNION ALL
			SELECT d.id FROM departments d
			INNER JOIN subtree s ON d.parent_id = s.id
		)
		SELECT id FROM subtree
	`, rootID).Scan(&ids).Error
	return ids, err
}

func (db *Database) GetDepartment(id int, depth int, includeEmployees bool) (*models.Department, error) {
	var dept models.Department
	if err := db.conn.First(&dept, id).Error; err != nil {
		return nil, errors.New("department not found")
	}

	if includeEmployees {
		if err := db.conn.Model(&dept).Order("created_at ASC").Association("Employees").Find(&dept.Employees); err != nil {
			return nil, err
		}
	}

	if depth > 0 {
		if err := db.loadChildren(&dept, depth, includeEmployees); err != nil {
			return nil, err
		}
	}
	return &dept, nil
}

func (db *Database) loadChildren(dept *models.Department, remainingDepth int, includeEmployees bool) error {
	if remainingDepth <= 1 {
		return nil
	}
	var children []models.Department
	if err := db.conn.Where("parent_id = ?", dept.ID).Find(&children).Error; err != nil {
		return err
	}
	for i := range children {
		if includeEmployees {
			if err := db.conn.Model(&children[i]).Order("created_at ASC").Association("Employees").Find(&children[i].Employees); err != nil {
				return err
			}
		}
		if remainingDepth > 1 {
			if err := db.loadChildren(&children[i], remainingDepth-1, includeEmployees); err != nil {
				return err
			}
		}
	}
	dept.Children = children
	return nil
}

func (db *Database) DeleteDepartment(id int, mode string, reassignToID *int) error {
	var dept models.Department
	if err := db.conn.First(&dept, id).Error; err != nil {
		return errors.New("department not found")
	}

	switch mode {
	case "cascade":
		childrenIDs, err := db.GetAllChildrenIDs(id)
		if err != nil {
			return err
		}
		if err := db.conn.Where("department_id IN ?", childrenIDs).Delete(&models.Employee{}).Error; err != nil {
			return err
		}
		return db.conn.Where("id IN ?", childrenIDs).Delete(&models.Department{}).Error

	case "reassign":
		if reassignToID == nil {
			return errors.New("reassign_to_department_id is required when mode=reassign")
		}
		var targetDept models.Department
		if err := db.conn.First(&targetDept, *reassignToID).Error; err != nil {
			return errors.New("target department not found")
		}
		tx := db.conn.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		if err := tx.Model(&models.Employee{}).Where("department_id = ?", id).Update("department_id", *reassignToID).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(&models.Department{}).Where("parent_id = ?", id).Update("parent_id", *reassignToID).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Delete(&models.Department{}, id).Error; err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit().Error

	default:
		return errors.New("invalid mode: must be 'cascade' or 'reassign'")
	}
}

// Employees

func (db *Database) CreateEmployee(emp models.Employee, department int) (*models.Employee, error) {
	var dept models.Department
	if err := db.conn.First(&dept, department).Error; err != nil {
		return nil, errors.New("department not found")
	}
	if len(emp.FullName) > 200 {
		return nil, errors.New("full name is longer than 200 symbols")
	}
	if len(emp.Position) > 200 {
		return nil, errors.New("position is longer than 200 symbols")
	}
	if strings.TrimSpace(emp.FullName) == "" {
		return nil, errors.New("full_name cannot be empty")
	}
	if strings.TrimSpace(emp.Position) == "" {
		return nil, errors.New("position cannot be empty")
	}

	empl := &models.Employee{
		DepartmentID: department,
		FullName:     emp.FullName,
		Position:     emp.Position,
		HiredAt:      emp.HiredAt,
	}
	if err := db.conn.Create(empl).Error; err != nil {
		return nil, err
	}
	return empl, nil
}

func (db *Database) DeleteEmployee(id int) error {
	result := db.conn.Delete(&models.Employee{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("employee not found")
	}
	return nil
}
