package main

import (
	"fmt"
	"strconv"
)

type Client struct {
	Id                uint
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

func NewReminderFrequency(s string) (ReminderFrequency, error) {
	if s == "HALFYEAR" {
		return HALFYEAR, nil
	} else if s == "YEAR" {
		return YEAR, nil
	}

	return 0, fmt.Errorf("invalid reminder frequency: %s", s)
}

func (r ReminderFrequency) String() string {
	if r == HALFYEAR {
		return "HALFYEAR"
	} else if r == YEAR {
		return "YEAR"
	}

	return "invalid reminder frequency"
}

func (r ReminderFrequency) StringGerman() string {
	if r == HALFYEAR {
		return "halbjährlich"
	} else if r == YEAR {
		return "jährlich"
	}

	return "invalid reminder frequency"
}

type month uint8

func NewMonth(s string) (month, error) {
	m, err := strconv.ParseUint(s, 10, 8)
	if err != nil || m < 1 || m > 12 {
		return 0, fmt.Errorf("invalid month: %s", s)
	}
	return month(m), nil
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
