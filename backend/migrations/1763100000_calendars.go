package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Calendars are other calendars the household subscribes to by their link
// (Google Calendar's secret iCal address, a published Outlook calendar, an
// iCloud public calendar...). The server fetches each one every half hour and
// keeps its events on the calendar, in the calendar's category.
//
// The link is a secret that lets anyone read that calendar, so it is hidden:
// it is set once, by POST /api/app/calendars, and the API never shows it.
//
// Imported entries remember where they came from: "calendar" for a
// subscription (removing it removes them) and "uid", the event's id in its
// calendar, so importing the same file again or refreshing a subscription
// updates them instead of adding them twice.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		categories, err := app.FindCollectionByNameOrId("categories")
		if err != nil {
			return err
		}
		calendars := core.NewBaseCollection("calendars")
		calendars.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 80},
			&core.URLField{Name: "url", Required: true, Hidden: true},
			&core.RelationField{Name: "category", CollectionId: categories.Id, MaxSelect: 1},
			// set by the server after each refresh
			&core.DateField{Name: "synced"},
			&core.TextField{Name: "error", Max: 500},
			&core.NumberField{Name: "events", OnlyInt: true},
			&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		signedIn := types.Pointer(`@request.auth.collectionName = "users"`)
		calendars.ListRule, calendars.ViewRule = signedIn, signedIn
		// added through POST /api/app/calendars, which checks and fetches the link;
		// afterwards its name and category can change
		calendars.UpdateRule = types.Pointer(`@request.auth.collectionName = "users" && @request.body.synced:isset = false && @request.body.error:isset = false && @request.body.events:isset = false && @request.body.created_by:isset = false`)
		calendars.DeleteRule = types.Pointer(`created_by = @request.auth.id || @request.auth.role = "owner"`)
		if err := app.Save(calendars); err != nil {
			return err
		}

		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		entries.Fields.Add(
			&core.RelationField{Name: "calendar", CollectionId: calendars.Id, MaxSelect: 1, CascadeDelete: true},
			&core.TextField{Name: "uid", Max: 500},
		)
		entries.AddIndex("idx_entries_uid", false, "calendar, uid", "")
		return app.Save(entries)
	}, func(app core.App) error {
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		entries.RemoveIndex("idx_entries_uid")
		entries.Fields.RemoveByName("calendar")
		entries.Fields.RemoveByName("uid")
		if err := app.Save(entries); err != nil {
			return err
		}
		calendars, err := app.FindCollectionByNameOrId("calendars")
		if err != nil {
			return err
		}
		return app.Delete(calendars)
	})
}
