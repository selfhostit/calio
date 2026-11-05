package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Documents replace contacts: a page of notes and files about something the
// household looks after, like "Wolfy" for the dog (food, walks, the vet's
// number) or "The house" for home sitting (alarm, plants, bins). A document
// can be shared with a sitter through a private link that works without
// signing in (share_token, managed through /api/app/documents/{id}/share).
//
// Each contact becomes a document with its details in the text.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		documents := core.NewBaseCollection("documents")
		documents.Fields.Add(
			&core.TextField{Name: "title", Required: true, Max: 120},
			&core.TextField{Name: "icon", Max: 20, Pattern: `^[a-z]*$`},
			&core.TextField{Name: "body", Max: 20000},
			&core.FileField{Name: "files", MaxSelect: 20, MaxSize: 20 << 20, Thumbs: []string{"400x400"}},
			&core.TextField{Name: "share_token", Max: 64, Hidden: true},
			// set by the server: whether share_token is set
			&core.BoolField{Name: "shared"},
			&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		documents.AddIndex("idx_documents_share_token", false, "share_token", "")
		signedIn := `@request.auth.id != ""`
		notShared := ` && @request.body.shared:isset = false`
		documents.ListRule, documents.ViewRule = types.Pointer(signedIn), types.Pointer(signedIn)
		documents.CreateRule = types.Pointer(`@request.auth.collectionName = "users"` + notShared)
		documents.UpdateRule = types.Pointer(signedIn + notShared)
		documents.DeleteRule = types.Pointer(`created_by = @request.auth.id || @request.auth.role = "owner"`)
		if err := app.Save(documents); err != nil {
			return err
		}

		contacts, err := app.FindCollectionByNameOrId("contacts")
		if err != nil {
			return nil // nothing to move
		}
		records, err := app.FindAllRecords(contacts)
		if err != nil {
			return err
		}
		roles := map[string]string{
			"vet": "Vet", "petsitter": "Pet sitter", "babysitter": "Babysitter",
			"doctor": "Doctor", "neighbour": "Neighbour",
		}
		for _, c := range records {
			var lines []string
			if role := roles[c.GetString("role")]; role != "" {
				lines = append(lines, role)
			}
			for _, f := range []struct{ label, name string }{
				{"Phone", "phone"}, {"Email", "email"}, {"Address", "address"},
			} {
				if v := strings.TrimSpace(c.GetString(f.name)); v != "" {
					lines = append(lines, f.label+": "+v)
				}
			}
			body := strings.Join(lines, "\n")
			if n := strings.TrimSpace(c.GetString("notes")); n != "" {
				body = strings.TrimSpace(body + "\n\n" + n)
			}
			icon := "user"
			if r := c.GetString("role"); r == "vet" || r == "petsitter" {
				icon = "paw"
			}
			d := core.NewRecord(documents)
			d.Load(map[string]any{
				"title": c.GetString("name"), "icon": icon, "body": body,
				"created_by": c.GetString("created_by"),
			})
			if err := app.Save(d); err != nil {
				return err
			}
		}
		return app.Delete(contacts)
	}, func(app core.App) error {
		documents, err := app.FindCollectionByNameOrId("documents")
		if err != nil {
			return err
		}
		return app.Delete(documents)
	})
}
