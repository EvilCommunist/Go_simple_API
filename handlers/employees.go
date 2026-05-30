package handlers

import (
	"API/db"
	"net/http"
)

type EmployeeHandler struct {
	conn *db.Database
}

func (d *EmployeeHandler) SetDatabase(connection *db.Database) {
	d.conn = connection
}

func (e *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {

}
