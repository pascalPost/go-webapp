package main

import (
	"fmt"
)

type Client struct {
	FirstName         string
	LastName          string
	Email             string
	ReminderMonth     month
	ReminderFrequency ReminderFrequency
	RegistrationDate  string
}

type ReminderFrequency uint8

const (
	HALFYEAR ReminderFrequency = iota
	YEAR
)

func (r ReminderFrequency) String() string {
	if r == HALFYEAR {
		return "halbjährlich"
	} else if r == YEAR {
		return "jährlich"
	}

	return "invalid reminder frequency"
}

type month uint8

func NewMonth(s string) (month, error) {
	switch s {
	case "Januar":
		return 1, nil
	case "Februar":
		return 2, nil
	case "März":
		return 3, nil
	case "April":
		return 4, nil
	case "Mai":
		return 5, nil
	case "Juni":
		return 6, nil
	case "Juli":
		return 7, nil
	case "August":
		return 8, nil
	case "September":
		return 9, nil
	case "Oktober":
		return 10, nil
	case "November":
		return 11, nil
	case "Dezember":
		return 12, nil
	default:
		return 0, fmt.Errorf("invalid month: %s", s)
	}
}

func (m month) String() string {
	switch m {
	case 1:
		return "Januar"
	case 2:
		return "Februar"
	case 3:
		return "März"
	case 4:
		return "April"
	case 5:
		return "Mai"
	case 6:
		return "Juni"
	case 7:
		return "Juli"
	case 8:
		return "August"
	case 9:
		return "September"
	case 10:
		return "Oktober"
	case 11:
		return "November"
	case 12:
		return "Dezember"
	default:
		return "Invalid month"
	}
}
