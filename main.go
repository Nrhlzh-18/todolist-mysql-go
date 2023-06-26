package main

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"path"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var db, _ = gorm.Open("mysql", "root:@/todolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct {
	Id        int `gorm:"primary_key"`
	Deskripsi string
	Completed bool
}

func Healthz(w http.ResponseWriter, r *http.Request) {

	var filepath = path.Join("view", "index.html")
	var tmpl, err = template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data = map[string]interface{}{
		"title": "Learning Golang Web",
		"name":  "Batman",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	description := r.FormValue("description")
	log.WithFields(log.Fields{"description": description}).Info("Add new TodoItem. Saving to database.")
	todo := &TodoItemModel{Deskripsi: description, Completed: false}
	db.Create(&todo)
	return
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	//test if the todoitem exist in DB
	err := GetItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "record not found"}`)
	} else {
		completed, _ := strconv.ParseBool(r.FormValue("completed"))
		log.WithFields(log.Fields{"id": id, "completed": completed}).Info("update TodoItem")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		db.Save(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": true}`)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	// get url param from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	//test if todo item exist in DB
	err := GetItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": false, "error": "record not found"}`)
	} else {
		log.WithFields(log.Fields{"id": id}).Info("deleting TodoItem")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		db.Delete(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
	}
}

func GetItemByID(id int) bool {
	todo := &TodoItemModel{}
	result := db.First(&todo, id)
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		return false
	}

	return true
}

func GetCompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("get completed TodoItems")
	completedTodoItems := GetTodoItems(true)
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(completedTodoItems)
}

func GetIncompleteditems(w http.ResponseWriter, r *http.Request) {
	log.Info("get incompleted TodoItems")
	incompletedTodoItems := GetTodoItems(false)
	json.NewEncoder(w).Encode(incompletedTodoItems)
}

func GetTodoItems(completed bool) interface{} {
	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func main() {
	defer db.Close()

	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})

	log.Info("Starting Todolist API server")
	router := mux.NewRouter()
	router.HandleFunc("/", Healthz).Methods("GET")
	router.HandleFunc("/todo-completed", GetCompletedItems).Methods("GET")
	router.HandleFunc("/todo-incompleted", GetIncompleteditems).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo/:id", UpdateItem).Methods("PUT")
	router.HandleFunc("/todo/:id", DeleteItem).Methods("DELET")

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)

	http.ListenAndServe(":8000", handler)
}
