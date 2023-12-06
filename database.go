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
	rows, err := db.handle.Query("SELECT firstname, lastname, email, reminder_month, reminder_frequency, created_at  FROM client")
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
		if err := rows.Scan(&client.FirstName, &client.LastName, &client.Email, &client.ReminderMonth, &client.ReminderFrequency, &client.RegistrationDate); err != nil {
			log.Println(err)
			return nil
		}
		clients = append(clients, client)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil
	}
	return clients
}

// AddClient adds a new client to the database
func (db *DatabaseConnection) AddClient(client Client) {
	_, err := db.handle.Exec("INSERT INTO client (firstname, lastname, email, reminder_month, reminder_frequency) VALUES (?, ?, ?, ?, ?)", client.FirstName, client.LastName, client.Email, client.ReminderMonth, client.ReminderFrequency)
	if err != nil {
		log.Println(err)
	}
}
