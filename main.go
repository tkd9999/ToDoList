package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

var db *sql.DB
var tpl *template.Template

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://project1:password@localhost/todolist?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")

	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

type Todo struct {
	Id       string
	Name     string
	Memo     string
	Deadline string
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/todo", todoIndex)
	http.HandleFunc("/todo/show", todoShow)
	http.HandleFunc("/todo/create", todoCreateForm)
	http.HandleFunc("/todo/create/process", todoCreateProcess)
	http.HandleFunc("/todo/update", todoUpdateForm)
	http.HandleFunc("/todo/update/process", todoUpdateProcess)
	http.HandleFunc("/todo/delete/process", todoDeleteProcess)
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/todo", http.StatusSeeOther)
}

func todoIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("select * from todos")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	tds := make([]Todo, 0)
	for rows.Next() {
		td := Todo{}
		err := rows.Scan(&td.Id, &td.Name, &td.Memo, &td.Deadline) // order matters
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		tds = append(tds, td)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tpl.ExecuteTemplate(w, "todo.gohtml", tds)
}

func todoShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM todos WHERE id = $1", id)

	td := Todo{}
	err := row.Scan(&td.Id, &td.Name, &td.Memo, &td.Deadline)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(w, "show.gohtml", td)
}

func todoCreateForm(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "create.gohtml", nil)
}

func todoCreateProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	td := Todo{}
	td.Name = r.FormValue("name")
	td.Memo = r.FormValue("memo")
	td.Deadline = r.FormValue("deadline")
	sID, _ := uuid.NewV4()
	td.Id = sID.String()

	if td.Name == "" || td.Deadline == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("insert into todos (id,name,memo, deadline) values ($1, $2, $3, $4)", td.Id, td.Name, td.Memo, td.Deadline)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(w, "created.gohtml", td)
}

func todoUpdateForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("select * from todos where id = $1", id)

	td := Todo{}
	err := row.Scan(&td.Id, &td.Name, &td.Memo, &td.Deadline)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "update.gohtml", td)
}

func todoUpdateProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	td := Todo{}
	td.Id = r.FormValue("id")
	td.Name = r.FormValue("name")
	td.Memo = r.FormValue("memo")
	td.Deadline = r.FormValue("deadline")

	if td.Id == "" || td.Name == "" || td.Deadline == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("update todos set name = $2, memo=$3, deadline = $4 where id=$1;", td.Id, td.Name, td.Memo, td.Deadline)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(w, "updated.gohtml", td)
}

func todoDeleteProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("delete from todos where id=$1;", id)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/todo", http.StatusSeeOther)
}
