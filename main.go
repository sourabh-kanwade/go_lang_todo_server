package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	Id     int    `json:"Id"`
	Name   string `json:"Name"`
	Status bool   `json:"Status"`
}

var db *sql.DB

func getTodoList(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM tasks;")
	if err != nil {
		log.Fatal("no tasks")
	}
	var tasks []Task
	for rows.Next() {
		var id int
		var name string
		var status bool
		err := rows.Scan(&id, &name, &status)
		if err != nil {
			log.Fatal("no tasks")
		}
		tasks = append(tasks, Task{Id: id, Name: name, Status: status})
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, insertErr := db.Exec("INSERT INTO tasks (Name,Status) VALUES (?,?)", task.Name, task.Status)
	if insertErr != nil {
		http.Error(w, insertErr.Error(), http.StatusInternalServerError)
	}
	insertedId, idErr := rows.LastInsertId()
	if idErr != nil {
		http.Error(w, idErr.Error(), http.StatusInternalServerError)
	}
	insertedRow := db.QueryRow("SELECT * FROM tasks WHERE Id=?", insertedId)

	var id int
	var name string
	var status bool
	scanErr := insertedRow.Scan(&id, &name, &status)
	if scanErr != nil {
		http.Error(w, "scan error", http.StatusNoContent)
	}
	newTask := Task{Id: id, Name: name, Status: status}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newTask)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	var task Task
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	decodeErr := json.NewDecoder(r.Body).Decode(&task)
	if decodeErr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("UPDATE tasks SET name=?, status=? WHERE id=?", task.Name, task.Status, id)
	if err != nil {
		log.Fatal(err)
	}
	updatedRow := db.QueryRow("SELECT * FROM tasks WHERE Id=?", id)

	var updatedTask Task
	getErr := updatedRow.Scan(&updatedTask.Id, &updatedTask.Name, &updatedTask.Status)

	if getErr == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	} else if getErr != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error scanning row:", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)

}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"id": id})

}

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./todos.db")

	if err != nil {
		panic(err)
	}
	_, qErr := db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		status INTEGER NOT NULL
)`)

	if qErr != nil {
		panic(qErr)
	}

}
func main() {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/todos", getTodoList).Methods("GET")
	r.HandleFunc("/todos", addTodo).Methods("POST")
	r.HandleFunc("/todos/{id}", updateTodo).Methods("PATCH")
	r.HandleFunc("/todos/{id}", deleteTodo).Methods("DELETE")
	fmt.Println("Server is Running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", r))
	// db.Close()
}
