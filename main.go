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

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/clients", http.StatusPermanentRedirect)
	})

	r.Mount("/clients", NewClients(state.db).Routes())
	r.Mount("/settings", NewSettings(state.db).Routes(state.db))

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
