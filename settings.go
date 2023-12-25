package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/99designs/keyring"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const serviceName = "reminder"

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

// createSettingsTable is the SQL statement to create the settings table
const createSettingsTable string = `
CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY CHECK (id = 0),
    smtp_address TEXT NOT NULL,
    smtp_port INTEGER NOT NULL,
    smtp_tls TEXT CHECK( smtp_tls IN ('SSL','STARTTLS') ) NOT NULL,
    email_from TEXT NOT NULL,
    email_from_name TEXT NOT NULL,
    email_subject TEXT NOT NULL,
    email_body TEXT NOT NULL
);`

// CreateSettingsTable creates the settings table
func CreateSettingsTable(db *sql.DB) error {
	_, err := db.Exec(createSettingsTable)
	return err
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

// GetSettings returns the settings from the database and the keyring
func GetSettings(db *DatabaseConnection) *Settings {
	rows, err := db.handle.Query("SELECT smtp_address, smtp_port, smtp_tls, email_from, email_from_name, email_subject, email_body FROM config")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	var settings Settings

	for rows.Next() {
		if err := rows.Scan(&settings.smtpAddress, &settings.smtpPort, &settings.smtpEncryption, &settings.emailFrom, &settings.emailFromName, &settings.emailSubject, &settings.emailBody); err != nil {
			log.Println(err)
			return nil
		}
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil
	}

	// get smtp credentials from keyring
	smtp_user, smtp_pass, err := GetSmtpCredentialsFromKeyring(serviceName)
	if err != nil && !errors.Is(err, noKeyFoundErr) {
		log.Println(err)
		return nil
	}

	settings.smtpUsername = smtp_user
	settings.smtpPassword = smtp_pass

	return &settings
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

func SettingRoutes(db *DatabaseConnection, settings *Settings) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		t, err := template.ParseFiles("templates/base.gohtml", "templates/settings.gohtml", "templates/navigation.gohtml")
		if err != nil {
			log.Println(err)
			return
		}
		if err := t.Execute(w, settings); err != nil {
			log.Println(err)
		}
	})

	r.Put("/", func(w http.ResponseWriter, r *http.Request) {
		settings.Set(parseSettings(r))
		fmt.Println(settings)
		UpdateSettings(db, settings)
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

// UpdateSettings updates the settings in the database and the keyring
func UpdateSettings(db *DatabaseConnection, settings *Settings) {
	if _, err := db.handle.Exec("UPDATE config SET smtp_address = ?, smtp_port = ?, smtp_tls = ?, email_from = ?, email_from_name = ?, email_subject = ?, email_body = ? WHERE id = 0", settings.smtpAddress, settings.smtpUsername, settings.smtpPassword, settings.smtpPort, settings.smtpEncryption, settings.emailFrom, settings.emailFromName, settings.emailSubject, settings.emailBody); err != nil {
		log.Println(err)
	}

	if err := SaveSmtpCredentialsInKeyring(serviceName, settings.smtpUsername, settings.smtpPassword); err != nil {
		log.Println(err)
	}
}

// SaveSmtpCredentialsInKeyring saves the smtp credentials in the keyring
func SaveSmtpCredentialsInKeyring(service, username, password string) error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: service,
	})
	if err != nil {
		return err
	}

	if err := ring.Set(keyring.Item{
		Key:  username,
		Data: []byte(password),
	}); err != nil {
		return err
	}

	return nil
}

var noKeyFoundErr = errors.New("no key found in keyring")

// GetSmtpCredentialsFromKeyring returns the smtp credentials from the keyring
func GetSmtpCredentialsFromKeyring(service string) (string, string, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: service,
	})
	if err != nil {
		return "", "", err
	}

	keys, err := ring.Keys()
	if err != nil {
		return "", "", err
	}

	if len(keys) == 0 {
		return "", "", noKeyFoundErr
	} else if len(keys) > 1 {
		return "", "", errors.New("more than one key found in keyring")
	}

	smtp_user := keys[0]

	smtp_pass, err := ring.Get(smtp_user)
	if err != nil {
		return "", "", err
	}

	return string(smtp_user), string(smtp_pass.Data), nil
}
