package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type encryption uint8

const (
	SSL encryption = iota
	STARTTLS
)

func (e encryption) Ssl() bool {
	return e == SSL
}

func (e encryption) Starttls() bool {
	return e == STARTTLS
}

func (e encryption) String() string {
	if e == SSL {
		return "SSL"
	} else if e == STARTTLS {
		return "STARTTLS"
	}

	return "invalid encryption"
}

func newEncryption(s string) (encryption, error) {
	if s == "SSL" {
		return SSL, nil
	} else if s == "STARTTLS" {
		return STARTTLS, nil
	}

	return 0, errors.New("invalid encryption")
}

func (e encryption) DefaultPort() uint16 {
	if e == SSL {
		return 465
	} else if e == STARTTLS {
		return 587
	}

	return 0
}

type Settings struct {
	smtpAddress    string
	smtpUsername   string
	smtpPassword   string
	smtpPort       uint16
	smtpEncryption encryption
	emailFrom      string
	emailFromName  string
	emailSubject   string
	emailBody      string
}

func (s *Settings) SmtpAddress() string {
	return s.smtpAddress
}

func (s *Settings) SmtpPort() uint16 {
	return s.smtpPort
}

func (s *Settings) SmtpEncryption() encryption {
	return s.smtpEncryption
}

func (s *Settings) SmtpUsername() string {
	return s.smtpUsername
}

func (s *Settings) SmtpPassword() string {
	return s.smtpPassword
}

func NewSettings(db *DatabaseConnection) *Settings {
	settings := db.GetSettings()
	return settings
}

func (s *Settings) Routes(templates *template.Template) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		//if err := templates.ExecuteTemplate(w, "settings.html", s); err != nil {
		//	log.Println(err)
		//}

		t, err := template.ParseFiles("templates/base.gohtml", "templates/settings.gohtml", "templates/navigation.gohtml")
		if err != nil {
			log.Println(err)
			return
		}
		if err := t.Execute(w, s); err != nil {
			log.Println(err)
		}
	})

	r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		//testClient := parseSettings(r)
		//testClient.Send(Test)
	})

	r.Put("/", func(w http.ResponseWriter, r *http.Request) {
		s = parseSettings(r)
	})

	return r
}

func parseSettings(r *http.Request) *Settings {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return nil
	}

	log.Printf("received form: %s\n", r.PostForm)

	port, err := strconv.ParseUint(r.FormValue("smtp_port"), 10, 16)
	if err != nil {
		log.Println(err)
		return nil
	}

	encryption, err := newEncryption(r.FormValue("smtp_tls"))
	if err != nil {
		log.Println(err)
		return nil
	}

	return &Settings{
		smtpAddress:    r.FormValue("smtp_address"),
		smtpUsername:   r.FormValue("smtp_username"),
		smtpPassword:   r.FormValue("smtp_password"),
		smtpPort:       uint16(port),
		smtpEncryption: encryption,
		emailFrom:      r.FormValue("email_from"),
		emailFromName:  r.FormValue("email_from_name"),
		emailSubject:   r.FormValue("email_subject"),
	}
}
