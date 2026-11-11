package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// An imported entry remembers the event as its calendar last gave it, as a
// fingerprint ("source"). A refresh or a second import only changes the
// entry when the event itself changed there, so moving an imported entry in
// Calio sticks until someone changes that event in its own calendar.
func init() {
	m.Register(func(app core.App) error {
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		entries.Fields.Add(&core.TextField{Name: "source", Max: 100, Hidden: true})
		return app.Save(entries)
	}, func(app core.App) error {
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		entries.Fields.RemoveByName("source")
		return app.Save(entries)
	})
}
