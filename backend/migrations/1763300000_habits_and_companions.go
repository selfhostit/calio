package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Weekdays a habit can be due on; a habit with none is due every day.
var Weekdays = []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

// Habits and companions: everyone keeps their own small daily habits ("brush
// teeth", "vitamin D") and has their own red panda whose mood follows them.
//
// A habit is due on some weekdays (or every day), a number of times a day
// (brushing teeth: twice), at a part of the day. A check is how many times it
// was done on a day, or that it was skipped on purpose. A day is a plain
// YYYY-MM-DD in the household's time, like all-day entries.
//
// The household sees everyone's habits and companions; only their owner
// changes them.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		signedIn := types.Pointer(`@request.auth.id != ""`)

		habits := core.NewBaseCollection("habits")
		habits.Fields.Add(
			&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, Required: true, CascadeDelete: true},
			&core.TextField{Name: "title", Required: true, Max: 60},
			&core.TextField{Name: "icon", Required: true, Max: 20, Pattern: `^[a-z]+$`},
			&core.NumberField{Name: "times", Required: true, Min: types.Pointer(1.0), Max: types.Pointer(5.0), OnlyInt: true},
			&core.SelectField{Name: "days", MaxSelect: len(Weekdays), Values: Weekdays},
			&core.SelectField{Name: "part", MaxSelect: 1, Values: []string{"morning", "afternoon", "evening"}},
			&core.NumberField{Name: "sort"},
			&core.BoolField{Name: "archived"},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		habits.AddIndex("idx_habits_user", false, "user", "")
		habits.ListRule, habits.ViewRule = signedIn, signedIn
		habits.CreateRule = types.Pointer(`@request.auth.collectionName = "users" && @request.body.user = @request.auth.id`)
		habits.UpdateRule = types.Pointer(`user = @request.auth.id && @request.body.user:isset = false`)
		habits.DeleteRule = types.Pointer(`user = @request.auth.id`)
		if err := app.Save(habits); err != nil {
			return err
		}

		checks := core.NewBaseCollection("habit_checks")
		checks.Fields.Add(
			&core.RelationField{Name: "habit", CollectionId: habits.Id, MaxSelect: 1, Required: true, CascadeDelete: true},
			&core.TextField{Name: "day", Required: true, Pattern: `^\d{4}-\d{2}-\d{2}$`},
			&core.NumberField{Name: "count", Min: types.Pointer(0.0), Max: types.Pointer(5.0), OnlyInt: true},
			&core.BoolField{Name: "skipped"},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		checks.AddIndex("idx_habit_checks_habit_day", true, "habit, day", "")
		checks.AddIndex("idx_habit_checks_day", false, "day", "")
		checks.ListRule, checks.ViewRule = signedIn, signedIn
		checks.CreateRule = types.Pointer(`habit.user = @request.auth.id`)
		checks.UpdateRule = types.Pointer(`habit.user = @request.auth.id && @request.body.habit:isset = false && @request.body.day:isset = false`)
		checks.DeleteRule = types.Pointer(`habit.user = @request.auth.id`)
		if err := app.Save(checks); err != nil {
			return err
		}

		companions := core.NewBaseCollection("companions")
		companions.Fields.Add(
			&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, Required: true, CascadeDelete: true},
			&core.TextField{Name: "name", Required: true, Max: 30},
			// more kinds of animal can follow; the web app draws each one
			&core.SelectField{Name: "species", Required: true, MaxSelect: 1, Values: []string{"red_panda"}},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		companions.AddIndex("idx_companions_user", true, "user", "")
		companions.ListRule, companions.ViewRule = signedIn, signedIn
		companions.CreateRule = types.Pointer(`@request.auth.collectionName = "users" && @request.body.user = @request.auth.id`)
		companions.UpdateRule = types.Pointer(`user = @request.auth.id && @request.body.user:isset = false && @request.body.species:isset = false`)
		companions.DeleteRule = nil
		return app.Save(companions)
	}, func(app core.App) error {
		for _, name := range []string{"companions", "habit_checks", "habits"} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			if err := app.Delete(c); err != nil {
				return err
			}
		}
		return nil
	})
}
