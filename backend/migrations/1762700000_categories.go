package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// CategoryColors are the colors a category can have. The web app spells out
// the classes for each one (web/src/lib/look.ts).
var CategoryColors = []string{
	"indigo", "amber", "teal", "rose", "sky", "emerald", "violet", "orange", "lime", "pink", "cyan", "slate",
}

// Categories are what the household puts on its calendar: events, pet
// sitting, trips, babysitting, or anything they add, each with a name, a
// color and an icon. A category can be switched off so it isn't offered for
// new entries; its entries stay on the calendar.
//
// A category's kind is how its entries behave: "petsit" entries need a sitter
// (pets, kids, the house) and "trip" entries get a packing list. An entry's
// kind follows its category.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		categories := core.NewBaseCollection("categories")
		categories.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 40},
			&core.SelectField{Name: "color", Required: true, MaxSelect: 1, Values: CategoryColors},
			&core.TextField{Name: "icon", Required: true, Max: 20, Pattern: `^[a-z]+$`},
			&core.SelectField{Name: "kind", Required: true, MaxSelect: 1, Values: []string{"event", "petsit", "trip"}},
			&core.BoolField{Name: "enabled"},
			&core.NumberField{Name: "sort"},
			&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		signedIn := types.Pointer(`@request.auth.id != ""`)
		categories.ListRule, categories.ViewRule, categories.CreateRule, categories.UpdateRule = signedIn, signedIn, signedIn, signedIn
		categories.DeleteRule = types.Pointer(`created_by = @request.auth.id || @request.auth.role = "owner"`)
		if err := app.Save(categories); err != nil {
			return err
		}

		// the categories every household starts with; babysitting is there to
		// switch on
		ids := map[string]string{}
		for i, c := range []struct {
			name, color, icon, kind string
			enabled                 bool
		}{
			{"Event", "indigo", "sparkle", "event", true},
			{"Pet sitting", "amber", "paw", "petsit", true},
			{"Trip", "teal", "plane", "trip", true},
			{"Babysitting", "rose", "baby", "petsit", false},
		} {
			r := core.NewRecord(categories)
			r.Load(map[string]any{"name": c.name, "color": c.color, "icon": c.icon, "kind": c.kind, "enabled": c.enabled, "sort": i})
			if err := app.Save(r); err != nil {
				return err
			}
			if _, ok := ids[c.kind]; !ok {
				ids[c.kind] = r.Id
			}
		}

		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		entries.Fields.Add(&core.RelationField{Name: "category", CollectionId: categories.Id, MaxSelect: 1})
		if err := app.Save(entries); err != nil {
			return err
		}
		// what is already on the calendar goes in the category of its kind
		for kind, id := range ids {
			if _, err := app.DB().NewQuery("UPDATE entries SET category = {:id} WHERE kind = {:kind}").
				Bind(map[string]any{"id": id, "kind": kind}).Execute(); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		entries.Fields.RemoveByName("category")
		if err := app.Save(entries); err != nil {
			return err
		}
		categories, err := app.FindCollectionByNameOrId("categories")
		if err != nil {
			return err
		}
		return app.Delete(categories)
	})
}
