package main

import (
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
)

type clients struct {
	db *DatabaseConnection
}

func (c *clients) GetClients() []Client {
	return c.db.GetClients()
}

func (c *clients) AddClient(client Client) {
	c.db.AddClient(client)
}

func NewClients(db *DatabaseConnection) *clients {
	return &clients{
		db: db,
	}
}

func (c *clients) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("templates/base.gohtml", "templates/clients.gohtml", "templates/navigation.gohtml")

		clients := c.GetClients()
		if err := t.Execute(w, clients); err != nil {
			log.Println(err)
		}
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
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

		c.AddClient(newClient)

		//// respond with a new empty form and a table update
		//if err := state.templates.ExecuteTemplate(w, "client.gohtml", state.db.GetClients()); err != nil {
		//	log.Println(err)
		//}
	})

	return r
}