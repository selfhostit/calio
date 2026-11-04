package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Each person gets a secret for their calendar feed: calendar apps can't sign
// in, so the feed's URL carries it. It's hidden from the API; the app hands it
// out through /api/app/feed-link.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.Fields.Add(&core.TextField{Name: "feed_token", Hidden: true, Max: 64})
		users.AddIndex("idx_users_feed_token", false, "feed_token", "feed_token != ''")
		return app.Save(users)
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.Fields.RemoveByName("feed_token")
		users.RemoveIndex("idx_users_feed_token")
		return app.Save(users)
	})
}
