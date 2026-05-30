package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"API/db"
	"API/db/models"
)

type DepartmentHandler struct {
	conn *db.Database
}

func (d *DepartmentHandler) SetDatabase(connection *db.Database) {
	d.conn = connection
}

func (d *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name     string `json:"name"`
		ParentID *int   `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(request.Name) == "" {
		http.Error(w, "Invalid name", http.StatusBadRequest)
		return
	}

	dept := models.Department{
		Name:     request.Name,
		ParentID: request.ParentID,
	}
	created, err := d.conn.CreateDepartment(dept, request.ParentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (d *DepartmentHandler) Read(w http.ResponseWriter, r *http.Request) {

}

func (d *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {

}

func (d *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {

}
