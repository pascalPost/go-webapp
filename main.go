package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"log"
	"net/http"
)

type State struct {
	db        *DatabaseConnection
	templates *template.Template
	settings  *Settings
}

func NewState() *State {
	db := NewDatabaseConnection()
	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	settings := NewSettings(db)

	return &State{
		db:        db,
		templates: templates,
		settings:  settings,
	}
}

func main() {
	state := NewState()
	defer state.db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/clients", http.StatusPermanentRedirect)
	})

	r.Get("/clients", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("templates/base.gohtml", "templates/clients.gohtml", "templates/navigation.gohtml", "templates/clientForm.gohtml", "templates/clientTable.gohtml", "templates/clientTableRow.gohtml")

		clients := state.db.GetClients()
		if err := t.Execute(w, clients); err != nil {
			log.Println(err)
		}
	})

	r.Mount("/client", NewClients(state.db).Routes())
	r.Mount("/settings", NewSettings(state.db).Routes(state.db))

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
