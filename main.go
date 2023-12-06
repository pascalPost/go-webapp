package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"log"
	"net/http"
)

type State struct {
	db *DatabaseConnection
}

func NewState() *State {
	return &State{
		db: NewDatabaseConnection(),
	}
}

func main() {
	templates := template.Must(template.ParseGlob("./*.html"))

	state := NewState()
	defer state.db.Close()

	newClient := Client{
		FirstName:         "Max",
		LastName:          "Mustermann",
		Email:             "max.mustermann@mail.com",
		ReminderMonth:     6,
		ReminderFrequency: YEAR,
	}
	state.db.AddClient(newClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		clients := state.db.GetClients()
		//clients := []Client{
		//	{"Max", "Mustermann", "max.mustermann@gmx.de", 1, 1, "2021-01-01"},
		//}

		if err := templates.ExecuteTemplate(w, "home.html", clients); err != nil {
			log.Println(err)
		}
	})

	r.Post("/client", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println(err)
			return
		}

		log.Printf("received form: %s\n", r.PostForm)

		// parse reminder month
		month, err := NewMonth(r.FormValue("reminderMonth"))
		if err != nil {
			log.Println(err)
			return
		}

		// parse reminder frequency
		var frequency ReminderFrequency
		if f := r.FormValue("reminderFrequency"); f == "yearly" {
			frequency = YEAR
		} else if f == "halfYearly" {
			frequency = HALFYEAR
		} else {
			log.Printf("Invalid reminder frequency %s (only yearly and halfYearly allowed)\n", f)
			return
		}

		newClient := Client{
			FirstName:         r.FormValue("firstname"),
			LastName:          r.FormValue("lastname"),
			Email:             r.FormValue("email"),
			ReminderMonth:     month,
			ReminderFrequency: frequency,
		}

		state.db.AddClient(newClient)

		// respond with a new empty form and a table update
		if err := templates.ExecuteTemplate(w, "client.html", state.db.GetClients()); err != nil {
			log.Println(err)
		}
	})

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
