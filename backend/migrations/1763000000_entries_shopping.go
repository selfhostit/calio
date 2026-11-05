package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// An entry can carry a shopping list, like the checklist: what to buy for
// the dinner or the party, as {"text": "...", "done": false} items. When
// Shoppa is set up, the list can be sent there (/api/app/shoppa).
func init() {
	m.Register(func(app core.App) error {
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		entries.Fields.Add(&core.JSONField{Name: "shopping", MaxSize: 64 << 10})
		return app.Save(entries)
	}, func(app core.App) error {
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		entries.Fields.RemoveByName("shopping")
		return app.Save(entries)
	})
}
