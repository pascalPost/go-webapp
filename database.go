package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

// dbFile is the name of the database file
const dbFile string = "db.sqlite"
const dbFileTest string = "file::memory:?cache=shared"

// createClientTable is the SQL statement to create the client table
const createClientTable string = `
  CREATE TABLE IF NOT EXISTS client (
  id INTEGER NOT NULL PRIMARY KEY,
  firstname TEXT NOT NULL,
  lastname TEXT NOT NULL,
  email TEXT NOT NULL,
  reminder_frequency TEXT CHECK( reminder_frequency IN ('YEAR','HALFYEAR') ) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
      );`

// DatabaseConnection represents a connection to the database
type DatabaseConnection struct {
	handle *sql.DB
}

// newDatabaseConnection creates a new database connection for the given file
func newDatabaseConnection(file string) *DatabaseConnection {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		panic(err)
	}
	if _, err := db.Exec(createClientTable); err != nil {
		panic(err)
	}
	if err := CreateSettingsTable(db); err != nil {
		panic(err)
	}
	if err := CreateEmailTable(db); err != nil {
		panic(err)
	}
	return &DatabaseConnection{db}
}

// NewDatabaseConnection creates a new database connection
func NewDatabaseConnection() *DatabaseConnection {
	return newDatabaseConnection(dbFile)
}

// NewTestDatabaseConnection creates a new in memory database connection for testing
func NewTestDatabaseConnection() *DatabaseConnection {
	return newDatabaseConnection(dbFileTest)
}

// Close the database connection
func (db *DatabaseConnection) Close() {
	if err := db.handle.Close(); err != nil {
		panic(err)
	}
}

// GetClients returns all clients in the database
func (db *DatabaseConnection) GetClients() []Client {
	rows, err := db.handle.Query("SELECT id, firstname, lastname, email, reminder_frequency, created_at  FROM client")
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
		if err := rows.Scan(&client.Id, &client.FirstName, &client.LastName, &client.Email, &reminderFrequencyStr, &client.RegistrationDate); err != nil {
			log.Println(err)
			return nil
		}
		if reminderFrequency, err := NewReminderFrequency(reminderFrequencyStr); err != nil {
			log.Println(err)
			return nil
		} else {
			client.ReminderFrequency = reminderFrequency
		}

		// get last email
		if lastEmail, err := db.GetLastEmail(client.Id); err == nil && lastEmail != nil {
			client.LastEmail = lastEmail.Time
		} else {
			log.Println(err)
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
	// TODO reduce duplicated code

	row := db.handle.QueryRow("SELECT id, firstname, lastname, email, reminder_frequency, created_at  FROM client WHERE id = ?", id)

	var client Client
	var reminderFrequencyStr string
	if err := row.Scan(&client.Id, &client.FirstName, &client.LastName, &client.Email, &reminderFrequencyStr, &client.RegistrationDate); err != nil {
		return nil, err
	}

	if reminderFrequency, err := NewReminderFrequency(reminderFrequencyStr); err != nil {
		return nil, err
	} else {
		client.ReminderFrequency = reminderFrequency
	}

	// get last email
	if lastEmail, err := db.GetLastEmail(id); err != nil {
		return &client, err
	} else if lastEmail != nil {
		client.LastEmail = lastEmail.Time
	}

	return &client, nil
}

// AddClient adds a new client to the database
func (db *DatabaseConnection) AddClient(client Client) (uint, error) {
	res, err := db.handle.Exec("INSERT INTO client (firstname, lastname, email, reminder_frequency) VALUES (?, ?, ?, ?)", client.FirstName, client.LastName, client.Email, client.ReminderFrequency.String())
	if err != nil {
		return 0, err
	}

	if rows, err := res.RowsAffected(); err != nil {
		return 0, err
	} else if rows == 0 {
		return 0, errors.New("no rows affected on insert")
	} else if rows > 1 {
		return 0, errors.New("more than one row affected on insert")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func (db *DatabaseConnection) UpdateClient(client Client) error {
	_, err := db.handle.Exec("UPDATE client SET firstname = ?, lastname = ?, email = ?, reminder_frequency = ? WHERE id = ?", client.FirstName, client.LastName, client.Email, client.ReminderFrequency.String(), client.Id)
	return err
}

func (db *DatabaseConnection) DeleteClient(id uint) error {
	rowsAffected, err := db.handle.Exec("DELETE FROM client WHERE id = ?", id)
	log.Println("affected rows by delete: ", rowsAffected)
	return err
}

type Email struct {
	Id       uint
	ClientId uint
	Time     time.Time
}

func parseQuery(rows *sql.Rows, err error) []Email {
	if err != nil {
		return nil
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	var emails []Email

	for rows.Next() {
		var e Email
		if err := rows.Scan(&e.Id, &e.ClientId, &e.Time); err != nil {
			log.Println(err)
			return nil
		}
		emails = append(emails, e)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil
	}
	return emails
}

func (db *DatabaseConnection) GetAllEmails() []Email {
	return parseQuery(db.handle.Query("SELECT id, client_id, sent_at FROM email"))
}

func (db *DatabaseConnection) GetEmails(clientId uint) []Email {
	return parseQuery(db.handle.Query("SELECT id, client_id, sent_at FROM email WHERE client_id = ?", clientId))
}

func (db *DatabaseConnection) GetLastEmail(clientId uint) (*Email, error) {
	row := db.handle.QueryRow("SELECT id, client_id, sent_at FROM email WHERE client_id = ? ORDER BY sent_at DESC LIMIT 1", clientId)

	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var e Email
	if err := row.Scan(&e.Id, &e.ClientId, &e.Time); err != nil {
		return nil, err
	}

	return &e, nil
}
