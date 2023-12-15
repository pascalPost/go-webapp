package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

// dbFile is the name of the database file
const dbFile string = "db.sqlite"

// createClientTable is the SQL statement to create the client table
const createClientTable string = `
  CREATE TABLE IF NOT EXISTS client (
  id INTEGER NOT NULL PRIMARY KEY,
  firstname TEXT NOT NULL,
  lastname TEXT NOT NULL,
  email TEXT NOT NULL,
  reminder_month INTEGER NOT NULL,
  reminder_frequency TEXT CHECK( reminder_frequency IN ('YEAR','HALFYEAR') ) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
      );`

const createSettingsTable string = `
CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY CHECK (id = 0),
    smtp_address TEXT NOT NULL,
    smtp_username TEXT NOT NULL,
    smtp_password TEXT NOT NULL,
    smtp_port INTEGER NOT NULL,
    smtp_tls TEXT CHECK( smtp_tls IN ('SSL','STARTTLS') ) NOT NULL,
    email_from TEXT NOT NULL,
    email_from_name TEXT NOT NULL,
    email_subject TEXT NOT NULL,
    email_body TEXT NOT NULL
);`

// DatabaseConnection represents a connection to the database
type DatabaseConnection struct {
	handle *sql.DB
}

// NewDatabaseConnection creates a new database connection
func NewDatabaseConnection() *DatabaseConnection {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}
	if _, err := db.Exec(createClientTable); err != nil {
		panic(err)
	}
	if _, err := db.Exec(createSettingsTable); err != nil {
		panic(err)
	}
	return &DatabaseConnection{db}
}

// Close the database connection
func (db *DatabaseConnection) Close() {
	if err := db.handle.Close(); err != nil {
		panic(err)
	}
}

// GetClients returns all clients in the database
func (db *DatabaseConnection) GetClients() []Client {
	rows, err := db.handle.Query("SELECT id, firstname, lastname, email, reminder_month, reminder_frequency, created_at  FROM client")
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

	var clients []Client

	for rows.Next() {
		var client Client
		var reminderFrequencyStr string
		if err := rows.Scan(&client.Id, &client.FirstName, &client.LastName, &client.Email, &client.ReminderMonth, &reminderFrequencyStr, &client.RegistrationDate); err != nil {
			log.Println(err)
			return nil
		}
		if reminderFrequency, err := NewReminderFrequency(reminderFrequencyStr); err != nil {
			log.Println(err)
			return nil
		} else {
			client.ReminderFrequency = reminderFrequency
		}
		clients = append(clients, client)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil
	}
	return clients
}

func (db *DatabaseConnection) GetClient(id uint) (*Client, error) {
	row := db.handle.QueryRow("SELECT id, firstname, lastname, email, reminder_month, reminder_frequency, created_at  FROM client WHERE id = ?", id)

	var client Client
	var reminderFrequencyStr string
	if err := row.Scan(&client.Id, &client.FirstName, &client.LastName, &client.Email, &client.ReminderMonth, &reminderFrequencyStr, &client.RegistrationDate); err != nil {
		return nil, err
	}

	if reminderFrequency, err := NewReminderFrequency(reminderFrequencyStr); err != nil {
		return nil, err
	} else {
		client.ReminderFrequency = reminderFrequency
	}

	return &client, nil
}

// AddClient adds a new client to the database
func (db *DatabaseConnection) AddClient(client Client) {
	_, err := db.handle.Exec("INSERT INTO client (firstname, lastname, email, reminder_month, reminder_frequency) VALUES (?, ?, ?, ?, ?)", client.FirstName, client.LastName, client.Email, client.ReminderMonth, client.ReminderFrequency.String())
	if err != nil {
		log.Println(err)
	}
}

func (db *DatabaseConnection) UpdateClient(client Client) error {
	_, err := db.handle.Exec("UPDATE client SET firstname = ?, lastname = ?, email = ?, reminder_month = ?, reminder_frequency = ? WHERE id = ?", client.FirstName, client.LastName, client.Email, client.ReminderMonth, client.ReminderFrequency.String(), client.Id)
	return err
}

func (db *DatabaseConnection) DeleteClient(id uint) error {
	rowsAffected, err := db.handle.Exec("DELETE FROM client WHERE id = ?", id)
	log.Println("affected rows by delete: ", rowsAffected)
	return err
}

func (db *DatabaseConnection) GetSettings() *Settings {
	rows, err := db.handle.Query("SELECT smtp_address, smtp_username, smtp_password, smtp_port, smtp_tls, email_from, email_from_name, email_subject, email_body FROM config")
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
		if err := rows.Scan(&settings.smtpAddress, &settings.smtpUsername, &settings.smtpPassword, &settings.smtpPort, &settings.smtpEncryption, &settings.emailFrom, &settings.emailFromName, &settings.emailSubject, &settings.emailBody); err != nil {
			log.Println(err)
			return nil
		}
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil
	}
	return &settings
}

func (db *DatabaseConnection) UpdateSettings(settings *Settings) {
	// TODO: hash password, see https://neverpanic.de/blog/2020/11/18/the-journey-to-storing-smtp-passwords-in-a-database/
	_, err := db.handle.Exec("UPDATE config SET smtp_address = ?, smtp_username = ?, smtp_password = ?, smtp_port = ?, smtp_tls = ?, email_from = ?, email_from_name = ?, email_subject = ?, email_body = ? WHERE id = 0", settings.smtpAddress, settings.smtpUsername, settings.smtpPassword, settings.smtpPort, settings.smtpEncryption, settings.emailFrom, settings.emailFromName, settings.emailSubject, settings.emailBody)
	if err != nil {
		log.Println(err)
	}
}
