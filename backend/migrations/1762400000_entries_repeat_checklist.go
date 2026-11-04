package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Entries can repeat (every day, week, month or year, until an optional last
// day) and carry a checklist, such as a packing list for a trip. The
// checklist is a list of {"text": "...", "done": false} items.
func init() {
	m.Register(func(app core.App) error {
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		entries.Fields.Add(
			&core.SelectField{Name: "repeat", MaxSelect: 1, Values: []string{"daily", "weekly", "monthly", "yearly"}},
			&core.DateField{Name: "repeat_until"},
			&core.JSONField{Name: "checklist", MaxSize: 64 << 10},
		)
		return app.Save(entries)
	}, func(app core.App) error {
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		for _, name := range []string{"repeat", "repeat_until", "checklist"} {
			entries.Fields.RemoveByName(name)
		}
		return app.Save(entries)
	})
}
