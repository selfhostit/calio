package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Birthdays get a category of their own (pink, with a cake), switched on, so
// a birthday calendar can be imported into it and anything titled "...
// birthday" lands there by itself (app.BirthdayCategory). A household that
// already made one keeps theirs.
//
// Birthdays already on the calendar in the plain Event category move to it.
func init() {
	m.Register(func(app core.App) error {
		categories, err := app.FindCollectionByNameOrId("categories")
		if err != nil {
			return err
		}
		existing, err := app.FindRecordsByFilter("categories",
			"name ~ 'birthday' || icon = 'cake'", "", 1, 0)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return nil
		}
		last, err := app.FindRecordsByFilter("categories", "", "-sort", 1, 0)
		if err != nil {
			return err
		}
		sort := 0
		if len(last) > 0 {
			sort = last[0].GetInt("sort") + 1
		}
		birthday := core.NewRecord(categories)
		birthday.Load(map[string]any{
			"name": "Birthday", "color": "pink", "icon": "cake", "kind": "event", "enabled": true, "sort": sort,
		})
		if err := app.Save(birthday); err != nil {
			return err
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
