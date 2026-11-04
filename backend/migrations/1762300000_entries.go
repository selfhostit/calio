package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Entries are everything on the household calendar: events, pet sitting and
// trips. Everyone in the household sees and edits them; only the person who
// added an entry, or the owner, can delete it.
//
// All-day entries keep their dates at midnight UTC and are read as plain
// calendar dates; timed entries are real instants.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		entries := core.NewBaseCollection("entries")
		entries.Fields.Add(
			&core.TextField{Name: "title", Required: true, Max: 200},
			&core.SelectField{Name: "kind", Required: true, MaxSelect: 1, Values: []string{"event", "petsit", "trip"}},
			&core.DateField{Name: "start", Required: true},
			&core.DateField{Name: "end"},
			&core.BoolField{Name: "all_day"},
			&core.TextField{Name: "location", Max: 200},
			&core.TextField{Name: "notes", Max: 2000},
			&core.RelationField{Name: "people", CollectionId: users.Id, MaxSelect: 20},
			&core.TextField{Name: "sitter", Max: 120},
			&core.SelectField{Name: "sitter_status", MaxSelect: 1, Values: []string{"needed", "asked", "booked"}},
			&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		entries.AddIndex("idx_entries_start", false, "start", "")
		signedIn := types.Pointer(`@request.auth.id != ""`)
		entries.ListRule, entries.ViewRule, entries.CreateRule, entries.UpdateRule = signedIn, signedIn, signedIn, signedIn
		entries.DeleteRule = types.Pointer(`created_by = @request.auth.id || @request.auth.role = "owner"`)
		return app.Save(entries)
	}, func(app core.App) error {
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		return app.Delete(entries)
	})
}
