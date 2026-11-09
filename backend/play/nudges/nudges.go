// Package nudges reminds people in the evening of the habits they haven't
// done yet, if they asked for it (companions.nudges).
package nudges

import (
	"fmt"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/selfhostit/calio/jobs"
	"github.com/selfhostit/calio/play/habits"
	"github.com/selfhostit/calio/push"
)

// When the reminder goes out, in the household's time.
const (
	Schedule = "0 20 * * *"
	// a run caught up after the server was off only nudges until then
	lastHour = 22
)

// Register schedules the evening reminder.
func Register(app core.App) {
	jobs.MustAdd(app, "habit-nudges", Schedule, func(app core.App) error {
		if time.Now().Hour() > lastHour {
			return nil
		}
		list, err := Due(app, habits.Today())
		for _, n := range list {
			push.ToUser(app, n.User, n.Message)
		}
		return err
	})
}

// Nudge is a reminder for one person.
type Nudge struct {
	User    string
	Message push.Message
}

// Due is the reminders to send for a day: to everyone who wants them and
// still has habits left.
func Due(app core.App, day string) ([]Nudge, error) {
	companions, err := app.FindAllRecords("companions", dbx.HashExp{"nudges": true})
	if err != nil {
		return nil, err
	}
	var out []Nudge
	for _, c := range companions {
		user := c.GetString("user")
		h, err := habits.Load(app, user, day, day)
		if err != nil {
			return out, err
		}
		left := h.Left(day)
		if left == 0 {
			continue
		}
		body := "One habit left today."
		if left > 1 {
			body = fmt.Sprintf("%d habits left today.", left)
		}
		out = append(out, Nudge{User: user, Message: push.Message{
			Title: c.GetString("name") + " is waiting for you",
			Body:  body,
			Tag:   "habit-nudge",
		}})
	}
	return out, nil
}
