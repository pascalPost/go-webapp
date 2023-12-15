package main

import (
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"strconv"
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
		if f := r.FormValue("reminderFrequency"); f == "1" {
			frequency = HALFYEAR
		} else if f == "2" {
			frequency = YEAR
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

		// respond with a new empty form
		t, _ := template.ParseFiles("templates/clientForm.gohtml")
		if err := t.Execute(w, c.db.GetClients()); err != nil {
			log.Println(err)
		}
	})

	r.Delete("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		log.Printf("deletion of client %v requested\n", idStr)
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			log.Println(err)
			return
		}
		if err := c.db.DeleteClient(uint(id)); err != nil {
			log.Println(err)
			return
		}
	})

	r.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			log.Println(err)
			return
		}

		client, err := c.db.GetClient(uint(id))
		if err != nil {
			log.Println(err)
			return
		}

		t, _ := template.ParseFiles("templates/clientTableRow.gohtml")
		if err := t.Execute(rw, client); err != nil {
			log.Println(err)
		}
	})

	r.Get("/{id}/edit", func(rw http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		log.Printf("edit of client %v requested\n", idStr)
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			log.Println(err)
			return
		}

		client, err := c.db.GetClient(uint(id))
		if err != nil {
			log.Println(err)
			return
		}

		t, _ := template.ParseFiles("templates/clientTableRowEdit.gohtml")
		if err := t.Execute(rw, client); err != nil {
			log.Println(err)
		}
	})

	r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
		// get id from url
		idStr := chi.URLParam(r, "id")
		log.Printf("update of client %v submitted\n", idStr)
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			log.Println(err)
			return
		}

		// parse form
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
		if f := r.FormValue("reminderFrequency"); f == "1" {
			frequency = HALFYEAR
		} else if f == "2" {
			frequency = YEAR
		} else {
			log.Printf("Invalid reminder frequency %s (only yearly and halfYearly allowed)\n", f)
			return
		}

		clientUpdate := Client{
			Id:                uint(id),
			FirstName:         r.FormValue("firstname"),
			LastName:          r.FormValue("lastname"),
			Email:             r.FormValue("email"),
			ReminderMonth:     month,
			ReminderFrequency: frequency,
		}

		if err := c.db.UpdateClient(clientUpdate); err != nil {
			log.Println(err)
			return
		}

		client, err := c.db.GetClient(uint(id))
		if err != nil {
			log.Println(err)
			return
		}

		t, _ := template.ParseFiles("templates/clientTableRow.gohtml")
		if err := t.Execute(w, client); err != nil {
			log.Println(err)
		}
	})

	return r
}
