package main

import (
	"API/db"
	"API/handlers"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	departments = "/departments"
	id          = "/{id}"
	employees   = "/employees"
)

func main() {
	r := chi.NewRouter()
	conn, err := db.Connect()
	if err != nil {
		fmt.Printf("Could not establish DB connections %s", err)
		return
	}
	db := &db.Database{}
	db.SetDatabase(conn)

	dh := handlers.DepartmentHandler{}
	dh.SetDatabase(db)
	eh := handlers.EmployeeHandler{}
	eh.SetDatabase(db)

	r.Post(departments, dh.Create)
	r.Post(departments+id+employees, eh.Create)

	r.Get(departments+id, dh.Read)

	r.Patch(departments+id, dh.Update)

	r.Delete(departments+id, dh.Delete)

	http.ListenAndServe(":8090", r)
}
