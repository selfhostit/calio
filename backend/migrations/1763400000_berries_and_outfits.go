package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Berries and outfits for the companions.
//
// The ledger holds every berry earned (a habit done, a whole day done, a
// streak) and spent (in the shop). Only the server writes it, and each
// reason is earned once: its ref is unique. People see their own.
//
// The inventory is what someone bought for their companion and whether it's
// wearing it. The shop route writes it; the household sees it.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		ledger := core.NewBaseCollection("ledger")
		ledger.Fields.Add(
			&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, Required: true, CascadeDelete: true},
			&core.NumberField{Name: "amount", Required: true, OnlyInt: true},
			&core.TextField{Name: "reason", Required: true, Max: 20},
			&core.TextField{Name: "ref", Required: true, Max: 120},
			&core.TextField{Name: "day", Pattern: `^\d{4}-\d{2}-\d{2}$`},
			&core.AutodateField{Name: "created", OnCreate: true},
		)
		ledger.AddIndex("idx_ledger_ref", true, "ref", "")
		ledger.AddIndex("idx_ledger_user", false, "user, reason", "")
		own := types.Pointer(`user = @request.auth.id`)
		ledger.ListRule, ledger.ViewRule = own, own
		if err := app.Save(ledger); err != nil {
			return err
		}

		inventory := core.NewBaseCollection("inventory")
		inventory.Fields.Add(
			&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, Required: true, CascadeDelete: true},
			&core.TextField{Name: "item", Required: true, Max: 40},
			&core.BoolField{Name: "equipped"},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		inventory.AddIndex("idx_inventory_user_item", true, "user, item", "")
		signedIn := types.Pointer(`@request.auth.id != ""`)
		inventory.ListRule, inventory.ViewRule = signedIn, signedIn
		return app.Save(inventory)
	}, func(app core.App) error {
		for _, name := range []string{"inventory", "ledger"} {
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
