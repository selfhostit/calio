package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Contacts are the people and places the household relies on: the vet, pet
// sitters, babysitters, the doctor. Everyone sees and edits them; only the
// person who added one, or the owner, can delete it.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		contacts := core.NewBaseCollection("contacts")
		contacts.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 120},
			&core.SelectField{Name: "role", Required: true, MaxSelect: 1,
				Values: []string{"vet", "petsitter", "babysitter", "doctor", "neighbour", "other"}},
			&core.TextField{Name: "phone", Max: 40},
			&core.EmailField{Name: "email"},
			&core.TextField{Name: "address", Max: 200},
			&core.TextField{Name: "notes", Max: 2000},
			&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		signedIn := types.Pointer(`@request.auth.id != ""`)
		contacts.ListRule, contacts.ViewRule, contacts.CreateRule, contacts.UpdateRule = signedIn, signedIn, signedIn, signedIn
		contacts.DeleteRule = types.Pointer(`created_by = @request.auth.id || @request.auth.role = "owner"`)
		return app.Save(contacts)
	}, func(app core.App) error {
		contacts, err := app.FindCollectionByNameOrId("contacts")
		if err != nil {
			return err
		}
		return app.Delete(contacts)
	})
}
