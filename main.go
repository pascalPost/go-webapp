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

	//r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	//	t, _ := template.ParseFiles("templates/base.gohtml", "templates/clients.gohtml", "templates/navigation.gohtml")
	//
	//	clients := state.db.GetClients()
	//	if err := t.Execute(w, clients); err != nil {
	//		log.Println(err)
	//	}
	//})
	//
	//r.Post("/client", func(w http.ResponseWriter, r *http.Request) {
	//	if err := r.ParseForm(); err != nil {
	//		log.Println(err)
	//		return
	//	}
	//
	//	log.Printf("received form: %s\n", r.PostForm)
	//
	//	// parse reminder month
	//	month, err := NewMonth(r.FormValue("reminderMonth"))
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//
	//	// parse reminder frequency
	//	var frequency ReminderFrequency
	//	if f := r.FormValue("reminderFrequency"); f == "yearly" {
	//		frequency = YEAR
	//	} else if f == "halfYearly" {
	//		frequency = HALFYEAR
	//	} else {
	//		log.Printf("Invalid reminder frequency %s (only yearly and halfYearly allowed)\n", f)
	//		return
	//	}
	//
	//	newClient := Client{
	//		FirstName:         r.FormValue("firstname"),
	//		LastName:          r.FormValue("lastname"),
	//		Email:             r.FormValue("email"),
	//		ReminderMonth:     month,
	//		ReminderFrequency: frequency,
	//	}
	//
	//	state.db.AddClient(newClient)
	//
	//	// respond with a new empty form and a table update
	//	if err := state.templates.ExecuteTemplate(w, "client.gohtml", state.db.GetClients()); err != nil {
	//		log.Println(err)
	//	}
	//})

	r.Mount("/clients", NewClients(state.db).Routes())
	r.Mount("/settings", NewSettings(state.db).Routes(state.templates))

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
