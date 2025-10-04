package osminokal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Appel-flappen/osminokal/config"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

type Calendar interface {
	PutSessions(deps Deps, ctx context.Context, sessions []FreeEnergySession) error
}

type CalendarFactory func(deps Deps, c *http.Client, cfg *config.CalendarConfig) (Calendar, error)

// New Source Location
var CalendarRegistry = map[config.CalendarType]CalendarFactory{
	config.CalendarCaldav: NewCaldavClient,
}

func NewCalendar(deps Deps, c *http.Client, cfg *config.CalendarConfig) (Calendar, error) {
	factoryFunc, ok := CalendarRegistry[config.CalendarType(cfg.Type)]
	if !ok {
		return nil, fmt.Errorf("unsupported calendar type: %s, for calendar name: %s", cfg.Type, cfg.Name)
	}

	return factoryFunc(deps, c, cfg)
}

type CaldavClient struct {
	CaldavClient *caldav.Client
	Endpoint     string
	Name         string
}

func NewCaldavClient(deps Deps, c *http.Client, cfg *config.CalendarConfig) (Calendar, error) {
	authClient := webdav.HTTPClientWithBasicAuth(c, cfg.Username, cfg.Password)
	cdClient, err := caldav.NewClient(authClient, cfg.Endpoint)
	if err != nil {
		deps.Logger.Error("error creating caldav client", "err", err, "calendar_name", cfg.Name)
	} else {
		deps.Logger.Info("created caldav client", "calendar_name", cfg.Name)
	}
	return &CaldavClient{
		CaldavClient: cdClient,
		Endpoint:     cfg.Endpoint,
		Name:         cfg.Name,
	}, nil
}

func (c CaldavClient) PutSessions(deps Deps, ctx context.Context, sessions []FreeEnergySession) error {
	// every session should actually be its own calendar
	// this makes it easier to have old ones not be overwritten but new ones will get added.
	reminders := []string{
		"-PT1H", // P: Period, T: Time, 1H: 1 Hour
		"-PT6H",
		"-PT12H",
	}

	for i := range sessions {
		cal := ical.NewCalendar()
		cal.Props.SetText(ical.PropVersion, "2.0")
		cal.Props.SetText(ical.PropProductID, "-//appel-flappen//osminokal//EN")

		event := ical.NewEvent()
		event.Props.SetText(ical.PropUID, sessions[i].ID.String())
		event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now())
		event.Props.SetDateTime(ical.PropDateTimeStart, sessions[i].Start)
		event.Props.SetDateTime(ical.PropDateTimeEnd, sessions[i].End)
		event.Props.SetText(ical.PropSummary, "Octopus Energy free energy session")

		// alarms
		for _, duration := range reminders {
			alarm := ical.NewComponent(ical.CompAlarm)

			// ACTION: Specifies what the alarm does (DISPLAY is common)
			alarm.Props.SetText(ical.PropAction, "DISPLAY")

			// DESCRIPTION: The text shown in the reminder box
			alarm.Props.SetText(ical.PropDescription, "reminder to plug in!")

			// TRIGGER: The key property that sets the timing.
			// PARAMETER: The related property sets it relative to START time.
			triggerProp := ical.NewProp(ical.PropTrigger)
			triggerProp.SetValueType(ical.ValueDuration)
			triggerProp.Value = duration
			triggerProp.Params.Add(ical.ParamRelated, "START")
			alarm.Props.Set(triggerProp)

			event.Children = append(event.Children, alarm)
			// DURATION and REPEAT can be added for repeating alarms, but are omitted here.
		}

		cal.Children = append(cal.Children, event.Component)
		eventFilename := sessions[i].ID.String() + "-osminokal-event.ics"
		// we expect user to provide path to calendar in the endpoint.
		calendarPath := ""
		eventPath := calendarPath + eventFilename

		calendarObject, err := c.CaldavClient.PutCalendarObject(ctx, eventPath, cal)
		if err != nil {
			deps.Logger.Error("error creating event in calendar", "calendar_name", c.Name, "err", err)
		} else {
			deps.Logger.Debug("event created successfully", "time", calendarObject.ModTime.String(), "path", calendarObject.Path, "calendar_name", c.Name)
		}
	}

	return nil
}
