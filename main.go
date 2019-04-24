package main

import (
	"bytes"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var templates = template.Must(template.ParseFiles("view/index.html", "view/item.html", "view/edit.html"))

type Entry struct {
	Title string
	Body  string
	Id    int
}

type Entry2 struct {
	Title string
	Body  []byte
}

type HtmlPage struct {
	Content template.HTML
}

// Journal index page content
func Journal(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hello World")
	var tpl bytes.Buffer

	db, err := sql.Open("mysql", "root:29760338@/uecram?charset=utf8")
	checkErr(err)

	// query
	rows, err := db.Query("SELECT title, content FROM chengshair.journal_entry;")
	checkErr(err)
	i := 1
	for rows.Next() {
		var title string
		var content string
		err = rows.Scan(&title, &content)
		checkErr(err)
		templates.ExecuteTemplate(&tpl, "item.html", &Entry{Title: title, Body: content, Id: i})
		i++
	}

	//fmt.Fprintf(w, tpl.String())
	//fmt.Fprintf(w, tpl.String())
	//templates.HTML()
	renderTemplate(w, "index", &HtmlPage{Content: template.HTML(tpl.String())})
}

//JournalEdit edit post
func JournalEdit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db, err := sql.Open("mysql", "root:29760338@/uecram?charset=utf8")
	checkErr(err)

	// query
	rows, err := db.Query("SELECT title, content FROM chengshair.journal_entry WHERE idjournal_entry=?;", vars["Id"])
	checkErr(err)

	rows.Next()
	var title string
	var content string
	err = rows.Scan(&title, &content)
	checkErr(err)

	i, err := strconv.Atoi(vars["Id"])
	templates.ExecuteTemplate(w, "edit.html", &Entry{Title: title, Body: content, Id: i})

}

//JournalUpdate update route
func JournalUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	r.ParseForm()
	db, err := sql.Open("mysql", "root:29760338@/uecram?charset=utf8")
	// update
	stmt, err := db.Prepare("UPDATE chengshair.journal_entry SET content=? WHERE idjournal_entry=?")
	checkErr(err)
	_, err = stmt.Exec(r.Form["body"][0], vars["Id"])
	checkErr(err)
	http.Redirect(w, r, "/journal", 301)
}

func renderTemplate(w http.ResponseWriter, tmpl string, entry *HtmlPage) {

	err := templates.ExecuteTemplate(w, tmpl+".html", entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

//MethodOverride middlewaear to handle put and patch method
func MethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only act on POST requests.
		if r.Method == "POST" {

			// Look in the request body and headers for a spoofed method.
			// Prefer the value in the request body if they conflict.
			method := r.PostFormValue("_method")
			if method == "" {
				method = r.Header.Get("X-HTTP-Method-Override")
			}

			// Check that the spoofed method is a valid HTTP method and
			// update the request object accordingly.
			if method == "PUT" || method == "PATCH" || method == "DELETE" {
				r.Method = method
			}
		}

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/journal", Journal)
	r.HandleFunc("/journal/edit/{Id}/", JournalEdit).Methods("POST")
	r.HandleFunc("/journal/edit/{Id}/", JournalUpdate).Methods("PUT")
	http.Handle("/", r)
	err := http.ListenAndServe(":8080", MethodOverride(r))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
