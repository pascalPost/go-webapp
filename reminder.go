package main

import (
	"fmt"
	"log"
	"time"
)

type Client struct {
	Id                uint
	FirstName         string
	LastName          string
	Email             string
	ReminderFrequency ReminderFrequency
	RegistrationDate  string
	LastEmail         time.Time
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

func (r ReminderFrequency) Months() uint8 {
	if r == HALFYEAR {
		return 6
	} else if r == YEAR {
		return 12
	} else {
		log.Fatal("invalid reminder frequency")
	}

	return 0
}
