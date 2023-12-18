package main

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"time"
)

const createEmailTable string = `
CREATE TABLE IF NOT EXISTS email (
    id INTEGER PRIMARY KEY,
    client_id INTEGER NOT NULL,
    sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (client_id) REFERENCES client (id)
);`

func CreateEmailTable(db *sql.DB) error {
	_, err := db.Exec(createEmailTable)
	return err
}

func (db *DatabaseConnection) AddEmailAtDate(clientId uint, date time.Time) error {
	_, err := db.handle.Exec("INSERT INTO email (client_id, sent_at) VALUES (?, ?)", clientId, date)
	return err
}

func (db *DatabaseConnection) AddEmail(clientId uint) error {
	_, err := db.handle.Exec("INSERT INTO email (client_id) VALUES (?)", clientId)
	return err
}

type emailHistoryEntry struct {
	ClientId      uint
	FirstName     string
	LastName      string
	Email         string
	LastEmailTime time.Time
	NextEmailTime time.Time
	PendingSince  time.Duration
}

func (db *DatabaseConnection) emailHistory() ([]emailHistoryEntry, error) {
	rows, err := db.handle.Query(`
SELECT client.id, client.firstname, client.lastname, client.email, client.reminder_frequency, email.sent_at
FROM client
INNER JOIN email ON client.id = email.client_id
ORDER BY email.sent_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	var entries []emailHistoryEntry
	var freqStr string
	for rows.Next() {
		var entry emailHistoryEntry
		if err := rows.Scan(&entry.ClientId, &entry.FirstName, &entry.LastName, &entry.Email, &freqStr, &entry.LastEmailTime); err != nil {
			return nil, err
		}

		freq, err := NewReminderFrequency(freqStr)
		if err != nil {
			return nil, err
		}

		entry.NextEmailTime = entry.LastEmailTime.AddDate(0, int(freq.Months()), 0)
		entry.PendingSince = time.Now().Sub(entry.NextEmailTime)

		entries = append(entries, entry)
	}

	return entries, nil
}

func EmailRoutes(db *DatabaseConnection) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		emailHist, err := db.emailHistory()

		// compute pending emails
		var pendingEmails []*emailHistoryEntry
		for _, entry := range emailHist {
			if entry.PendingSince > 0 {
				pendingEmails = append(pendingEmails, &entry)
			}
		}

		if err != nil {
			log.Println(err)
			return
		}

		t, _ := template.ParseFiles("templates/base.gohtml", "templates/emails.gohtml", "templates/navigation.gohtml")

		if err := t.Execute(w, struct {
			Pending []*emailHistoryEntry
			History []emailHistoryEntry
		}{
			Pending: pendingEmails,
			History: emailHist,
		}); err != nil {
			log.Println(err)
		}
	})

	return r
}
