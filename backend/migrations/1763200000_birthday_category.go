package migrations

import (
	"regexp"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// birthdayName matches app.IsBirthday (backend/app/birthdays.go).
var birthdayName = regexp.MustCompile(`(?i)birth ?days?|\bb-?days?\b|verjaardag|geburtstag|cumpleaños|anniversaire|γενέθλια`)

// Birthdays get a category of their own (pink, with a cake), switched on, so
// a birthday calendar can be imported into it and anything titled "...
// birthday" lands there by itself (app.BirthdayCategory). A household that
// already made one (named Birthdays, Verjaardagen, Γενέθλια... or with the cake
// icon) keeps theirs and gets no second one.
//
// Birthdays already on the calendar in the plain Event category move to it.
func init() {
	m.Register(func(app core.App) error {
		categories, err := app.FindCollectionByNameOrId("categories")
		if err != nil {
			return err
		}
		all, err := app.FindRecordsByFilter("categories", "", "sort,created", 0, 0)
		if err != nil {
			return err
		}
		// one the household made already, by its name (in a few languages, as
		// app.IsBirthday reads it) or its cake icon
		var birthday *core.Record
		for _, c := range all {
			if birthdayName.MatchString(c.GetString("name")) || c.GetString("icon") == "cake" {
				birthday = c
				break
			}
		}
		if birthday == nil {
			sort := 0
			if len(all) > 0 {
				sort = all[len(all)-1].GetInt("sort") + 1
			}
			birthday = core.NewRecord(categories)
			birthday.Load(map[string]any{
				"name": "Birthday", "color": "pink", "icon": "cake", "kind": "event", "enabled": true, "sort": sort,
			})
			if err := app.Save(birthday); err != nil {
				return err
			}
		}
		if birthday.GetString("kind") != "event" {
			return nil
		}

		events, err := app.FindRecordsByFilter("categories", "name = 'Event' && kind = 'event'", "", 0, 0)
		if err != nil {
			return err
		}
		for _, c := range events {
			if _, err := app.DB().NewQuery(
				"UPDATE entries SET category = {:birthday} WHERE category = {:event} AND title LIKE '%birthday%'").
				Bind(map[string]any{"birthday": birthday.Id, "event": c.Id}).Execute(); err != nil {
				return err
			}
		}
		return nil
	}, nil)
}
