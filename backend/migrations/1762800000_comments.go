package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Comments on an entry: "I'll take Luna to the vet first", "Can you pick up
// the keys?". Everyone else in the household is told about a new one. Only
// the person who wrote a comment changes it; they or the owner delete it.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		entries, err := app.FindCollectionByNameOrId("entries")
		if err != nil {
			return err
		}
		comments := core.NewBaseCollection("comments")
		comments.Fields.Add(
			&core.RelationField{Name: "entry", CollectionId: entries.Id, MaxSelect: 1, Required: true, CascadeDelete: true},
			&core.TextField{Name: "text", Required: true, Max: 2000},
			// set by the server to whoever wrote it
			&core.RelationField{Name: "author", CollectionId: users.Id, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		comments.AddIndex("idx_comments_entry", false, "entry, created", "")
		comments.ListRule = types.Pointer(`@request.auth.id != ""`)
		comments.ViewRule = comments.ListRule
		comments.CreateRule = types.Pointer(`@request.auth.collectionName = "users"`)
		comments.UpdateRule = types.Pointer(`author = @request.auth.id && @request.body.author:isset = false && @request.body.entry:isset = false`)
		comments.DeleteRule = types.Pointer(`author = @request.auth.id || @request.auth.role = "owner"`)
		return app.Save(comments)
	}, func(app core.App) error {
		comments, err := app.FindCollectionByNameOrId("comments")
		if err != nil {
			return err
		}
		return app.Delete(comments)
	})
}
