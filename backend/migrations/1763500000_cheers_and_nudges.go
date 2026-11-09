package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// The household side of companions.
//
// A companion now knows how grown it is (stage, kept by the server as
// habits are done, so everyone sees it) and whether its person wants an
// evening reminder of the habits left (nudges).
//
// A cheer is one person cheering another on, at most once a day each; the
// cheer route writes it, and the two of them can see it.
func init() {
	m.Register(func(app core.App) error {
		companions, err := app.FindCollectionByNameOrId("companions")
		if err != nil {
			return err
		}
		companions.Fields.Add(
			&core.SelectField{Name: "stage", MaxSelect: 1, Values: []string{"cub", "young", "grown"}},
			&core.BoolField{Name: "nudges"},
		)
		// people change their companion's name and nudges, not its stage
		companions.UpdateRule = types.Pointer(`user = @request.auth.id && @request.body.user:isset = false && @request.body.species:isset = false && @request.body.stage:isset = false`)
		if err := app.Save(companions); err != nil {
			return err
		}
		// how grown each one already is, from the habits done so far (see
		// rewards.Stages)
		if _, err := app.DB().NewQuery(`
			UPDATE companions SET stage = CASE
				WHEN n >= 150 THEN 'grown' WHEN n >= 30 THEN 'young' ELSE 'cub' END
			FROM (SELECT companions.id AS cid, (SELECT count(*) FROM ledger
				WHERE ledger.user = companions.user AND ledger.reason = 'habit') AS n FROM companions)
			WHERE companions.id = cid`).Execute(); err != nil {
			return err
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		cheers := core.NewBaseCollection("cheers")
		cheers.Fields.Add(
			&core.RelationField{Name: "from", CollectionId: users.Id, MaxSelect: 1, Required: true, CascadeDelete: true},
			&core.RelationField{Name: "to", CollectionId: users.Id, MaxSelect: 1, Required: true, CascadeDelete: true},
			&core.TextField{Name: "day", Required: true, Pattern: `^\d{4}-\d{2}-\d{2}$`},
			&core.AutodateField{Name: "created", OnCreate: true},
		)
		cheers.AddIndex("idx_cheers_from_to_day", true, "`from`, `to`, day", "")
		involved := types.Pointer(`from = @request.auth.id || to = @request.auth.id`)
		cheers.ListRule, cheers.ViewRule = involved, involved
		return app.Save(cheers)
	}, func(app core.App) error {
		cheers, err := app.FindCollectionByNameOrId("cheers")
		if err != nil {
			return err
		}
		if err := app.Delete(cheers); err != nil {
			return err
		}
		companions, err := app.FindCollectionByNameOrId("companions")
		if err != nil {
			return err
		}
		companions.Fields.RemoveByName("stage")
		companions.Fields.RemoveByName("nudges")
		companions.UpdateRule = types.Pointer(`user = @request.auth.id && @request.body.user:isset = false && @request.body.species:isset = false`)
		return app.Save(companions)
	})
}
