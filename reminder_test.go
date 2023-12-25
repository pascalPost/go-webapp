package main

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

// TestKeyring tests the keyring which is used to store the SMTP credentials
func TestKeyring(t *testing.T) {
	service := "reminder_test"
	username := "test"
	password := "test_pw"

	err := SaveSmtpCredentialsInKeyring(service, username, password)
	assert.NoError(t, err)

	smtp_user, smtp_pass, err := GetSmtpCredentialsFromKeyring(service)
	assert.NoError(t, err)
	assert.Equal(t, username, smtp_user)
	assert.Equal(t, password, smtp_pass)
}

func TestAddAndGetClient(t *testing.T) {
	db := NewTestDatabaseConnection()
	defer db.Close()

	newClient := Client{
		FirstName:         "Test",
		LastName:          "Client",
		Email:             "test@client.com",
		ReminderFrequency: YEAR,
	}

	if id, err := db.AddClient(newClient); err != nil {
		t.Error(err)
	} else {
		client, err := db.GetClient(id)

		assert.NoError(t, err)
		assert.Equal(t, id, client.Id)
		assert.Equal(t, newClient.FirstName, client.FirstName)
		assert.Equal(t, newClient.LastName, client.LastName)
		assert.Equal(t, newClient.Email, client.Email)
		assert.Equal(t, newClient.ReminderFrequency, client.ReminderFrequency)
	}
}

func isReminderEmailDue(currentMonth time.Month, reminderMonth time.Month, frequency ReminderFrequency) bool {
	if frequency == YEAR {
		return currentMonth == reminderMonth
	} else if frequency == HALFYEAR {
		halfReminder := reminderMonth + 6

		if halfReminder > 12 {
			halfReminder -= 12
		}

		return currentMonth == reminderMonth || currentMonth == halfReminder
	} else {
		return false
	}
}

func TestMarkEmailsAsDue(t *testing.T) {
	log.Println(time.Now())

	// assume it is January
	currentMonth := time.January
	reminderMonth := time.January
	frequency := YEAR

	assert.True(t, isReminderEmailDue(currentMonth, reminderMonth, frequency))

	reminderMonth = time.February
	assert.False(t, isReminderEmailDue(currentMonth, reminderMonth, frequency))

	reminderMonth = time.June
	assert.False(t, isReminderEmailDue(currentMonth, reminderMonth, frequency))

	reminderMonth = time.June
	currentMonth = time.January
	frequency = HALFYEAR
	assert.False(t, isReminderEmailDue(currentMonth, reminderMonth, frequency))

	reminderMonth = time.January
	currentMonth = time.July
	frequency = HALFYEAR
	assert.True(t, isReminderEmailDue(currentMonth, reminderMonth, frequency))

	reminderMonth = time.December
	currentMonth = time.June
	frequency = HALFYEAR
	assert.True(t, isReminderEmailDue(currentMonth, reminderMonth, frequency))

	reminderMonth = time.June
	currentMonth = time.December
	frequency = HALFYEAR
	assert.True(t, isReminderEmailDue(currentMonth, reminderMonth, frequency))

	reminderMonth = time.June
	currentMonth = time.November
	frequency = HALFYEAR
	assert.False(t, isReminderEmailDue(currentMonth, reminderMonth, frequency))

	reminderMonth = time.May
	currentMonth = time.November
	frequency = HALFYEAR
	assert.True(t, isReminderEmailDue(currentMonth, reminderMonth, frequency))
}

//type PendingMail struct {
//	ClientId uint
//	time.Duration
//}
//
//func GetPendingEmails(db *DatabaseConnection) []PendingMail {
//	var pending []PendingMail
//
//	currentMonth := time.Now().Month()
//
//	clients := db.GetClients()
//
//	for _, client := range clients {
//		if email, err := db.GetLastEmail(client.Id); err != nil || email == nil {
//			if err != nil {
//				log.Println(err)
//			}
//			// no email found or error (treated as no email found)
//			if isReminderEmailDue(currentMonth, time.Month(client.ReminderMonth), client.ReminderFrequency){
//				pending = append(pending, PendingMail{
//					ClientId: client.Id,
//					Reason:   "initial email due",
//					DueMonths: 0,
//				})
//			}
//		} else {
//			// email found
//			diff := time.Now().Sub(email.Time)
//
//			diffMonth := diff.Hours() / 24 / 30
//
//
//			var dueNumberOfMonths uint
//			if client.ReminderFrequency == YEAR {
//				dueNumberOfMonths =
//			} else if client.ReminderFrequency == HALFYEAR {
//
//			} else {
//				log.Println("invalid reminder frequency encountered")
//			}
//
//			if (client.ReminderFrequency == YEAR && diffMonth >= 12) || (client.ReminderFrequency == HALFYEAR && diffMonth >= 6) {
//				// send email
//
//			}
//		}
//
//	}
//
//	// all clients w/o an email sent in the last 6 months
//
//	// duration since the last email sent
//
//	// loop all clients
//
//	// find the last email sent
//
//}
//
////func TestEmail(t *testing.T) {
////	db := NewTestDatabaseConnection()
////	defer db.Close()
////
////	client1 := Client{
////		FirstName:         "Test",
////		LastName:          "Client",
////		Email:             "test@client.com",
////		ReminderMonth:     6,
////		ReminderFrequency: YEAR,
////	}
////
////	client2 := Client{
////		FirstName:         "Test2",
////		LastName:          "Client",
////		Email:             "test@client.com",
////		ReminderMonth:     4,
////		ReminderFrequency: YEAR,
////	}
////
////	if _, err := db.AddClient(client1); err != nil {
////		t.Error(err)
////	}
////
////}
