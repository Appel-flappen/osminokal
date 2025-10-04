package config

import (
	"slices"
	"strings"
)

type CalendarType string

// New Calendar Location
const (
	CalendarCaldav CalendarType = "caldav"
)

// New Calendar Location
var AllCalendars = []CalendarType{
	CalendarCaldav,
}

func (c CalendarType) IsValid() bool {
	return slices.Contains(AllCalendars, c)
}

func AllowedCalendarsString() string {
	strCalendars := make([]string, len(AllCalendars))
	for i, s := range AllCalendars {
		strCalendars[i] = string(s)
	}
	return strings.Join(strCalendars, ", ")
}
