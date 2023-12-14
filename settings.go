package main

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type encryptionType uint8

const (
	SSL encryptionType = iota
	STARTTLS
)

func (e encryptionType) Ssl() bool {
	return e == SSL
}

func (e encryptionType) Starttls() bool {
	return e == STARTTLS
}

func (e encryptionType) String() string {
	if e == SSL {
		return "SSL"
	} else if e == STARTTLS {
		return "STARTTLS"
	}

	return "invalid encryptionType"
}

func newEncryptionType(s string) (encryptionType, error) {
	s = strings.ToUpper(s)
	if s == "SSL" {
		return SSL, nil
	} else if s == "STARTTLS" {
		return STARTTLS, nil
	}

	return 0, errors.New("invalid encryptionType")
}

func (e encryptionType) DefaultPort() uint16 {
	if e == SSL {
		return 465
	} else if e == STARTTLS {
		return 587
	}

	return 0
}

type Encryption struct {
	t encryptionType
}

func newEncryption(s string) (Encryption, error) {
	t, err := newEncryptionType(s)
	return Encryption{t}, err
}

func (e Encryption) Ssl() bool {
	return e.t.Ssl()
}

func (e Encryption) Starttls() bool {
	return e.t.Starttls()
}

func (e Encryption) String() string {
	return e.t.String()
}

type Settings struct {
	smtpAddress    string
	smtpUsername   string
	smtpPassword   string
	smtpPort       uint16
	smtpEncryption Encryption
	emailFrom      string
	emailFromName  string
	emailSubject   string
	emailBody      string
}

func (s *Settings) Set(newSettings *Settings) {
	s.smtpAddress = newSettings.smtpAddress
	s.smtpUsername = newSettings.smtpUsername
	s.smtpPassword = newSettings.smtpPassword
	s.smtpPort = newSettings.smtpPort
	s.smtpEncryption = newSettings.smtpEncryption
	s.emailFrom = newSettings.emailFrom
	s.emailFromName = newSettings.emailFromName
	s.emailSubject = newSettings.emailSubject
	s.emailBody = newSettings.emailBody
}

func (s *Settings) SmtpAddress() string {
	return s.smtpAddress
}

func (s *Settings) SmtpPort() uint16 {
	return s.smtpPort
}

func (s *Settings) SmtpEncryption() Encryption {
	return s.smtpEncryption
}

func (s *Settings) SmtpUsername() string {
	return s.smtpUsername
}

func (s *Settings) SmtpPassword() string {
	return s.smtpPassword
}

func (s *Settings) EmailFrom() string {
	return s.emailFrom
}

func (s *Settings) EmailFromName() string {
	return s.emailFromName
}

func (s *Settings) EmailSubject() string {
	return s.emailSubject
}

func (s *Settings) EmailBody() string {
	return s.emailBody
}

func NewSettings(db *DatabaseConnection) *Settings {
	settings := db.GetSettings()
	return settings
}

func (s *Settings) Routes(db *DatabaseConnection) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		t, err := template.ParseFiles("templates/base.gohtml", "templates/settings.gohtml", "templates/navigation.gohtml")
		if err != nil {
			log.Println(err)
			return
		}
		if err := t.Execute(w, s); err != nil {
			log.Println(err)
		}
	})

	//r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
	//	//testClient := parseSettings(r)
	//	//testClient.Send(Test)
	//})

	r.Put("/", func(w http.ResponseWriter, r *http.Request) {
		s.Set(parseSettings(r))
		fmt.Println(s)
		db.UpdateSettings(s)
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

	encryption, err := newEncryption(r.FormValue("smtp_encryption"))
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
