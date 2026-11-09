// Package shop sells outfits for companions, for berries. The catalog is
// here, with the prices; the web app draws each item by its id.
package shop

import (
	"errors"
	"net/http"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"github.com/selfhostit/calio/play/rewards"
)

// Inventory is the collection, from migration 1763400000.
const Inventory = "inventory"

// Item is something a companion can wear. A companion wears one item per slot.
type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Slot  string `json:"slot"`
	Price int    `json:"price"`
}

// Catalog is everything in the shop, cheapest first.
var Catalog = []Item{
	{"bow_tie", "Bow tie", "neck", 15},
	{"bandana", "Bandana", "neck", 15},
	{"party_hat", "Party hat", "head", 20},
	{"scarf", "Striped scarf", "neck", 20},
	{"flower_crown", "Flower crown", "head", 25},
	{"beanie", "Cozy beanie", "head", 30},
	{"glasses", "Round glasses", "face", 35},
	{"explorer_hat", "Explorer hat", "head", 45},
}

// Find is an item in the catalog by its id.
func Find(id string) (Item, bool) {
	for _, i := range Catalog {
		if i.ID == id {
			return i, true
		}
	}
	return Item{}, false
}

// Register serves the shop under /api/app/play.
func Register(app core.App) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		g := se.Router.Group("/api/app/play").Bind(apis.RequireAuth("users"))
		g.GET("/shop", func(e *core.RequestEvent) error {
			return e.JSON(http.StatusOK, map[string]any{"items": Catalog})
		})
		g.POST("/buy", buy)
		g.POST("/wear", wear)
		return se.Next()
	})
}

// buy spends berries on an item, which the companion then wears.
func buy(e *core.RequestEvent) error {
	var body struct {
		Item string `json:"item"`
	}
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("", err)
	}
	item, ok := Find(body.Item)
	if !ok {
		return e.NotFoundError("That isn't in the shop.", nil)
	}
	user := e.Auth.Id
	err := e.App.RunInTransaction(func(tx core.App) error {
		if owned(tx, user, item.ID) != nil {
			return errOwned
		}
		if err := rewards.Spend(tx, user, item.Price, "shop", "shop:"+user+":"+item.ID); err != nil {
			return err
		}
		c, err := tx.FindCachedCollectionByNameOrId(Inventory)
		if err != nil {
			return err
		}
		r := core.NewRecord(c)
		r.Load(map[string]any{"user": user, "item": item.ID})
		if err := tx.Save(r); err != nil {
			return err
		}
		return put(tx, user, item, true)
	})
	switch {
	case errors.Is(err, errOwned):
		return e.BadRequestError("You have that one already.", nil)
	case errors.Is(err, rewards.ErrNotEnough):
		return e.BadRequestError("Not enough berries for that yet.", nil)
	case err != nil:
		return e.InternalServerError("", err)
	}
	return e.JSON(http.StatusOK, map[string]any{"item": item.ID})
}

// wear puts an item on (taking off whatever was in its slot) or takes it off.
func wear(e *core.RequestEvent) error {
	var body struct {
		Item string `json:"item"`
		On   bool   `json:"on"`
	}
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("", err)
	}
	item, ok := Find(body.Item)
	if !ok || owned(e.App, e.Auth.Id, item.ID) == nil {
		return e.NotFoundError("You don't have that one.", nil)
	}
	err := e.App.RunInTransaction(func(tx core.App) error {
		return put(tx, e.Auth.Id, item, body.On)
	})
	if err != nil {
		return e.InternalServerError("", err)
	}
	return e.NoContent(http.StatusNoContent)
}

var errOwned = errors.New("already owned")

func owned(app core.App, user, item string) *core.Record {
	r, err := app.FindFirstRecordByFilter(Inventory, "user = {:user} && item = {:item}",
		dbx.Params{"user": user, "item": item})
	if err != nil {
		return nil
	}
	return r
}

// put wears or takes off an item; wearing it takes off the others in its slot.
func put(tx core.App, user string, item Item, on bool) error {
	mine, err := tx.FindAllRecords(Inventory, dbx.HashExp{"user": user})
	if err != nil {
		return err
	}
	for _, r := range mine {
		other, _ := Find(r.GetString("item"))
		want := r.GetBool("equipped")
		switch {
		case r.GetString("item") == item.ID:
			want = on
		case on && other.Slot == item.Slot:
			want = false
		}
		if want != r.GetBool("equipped") {
			r.Set("equipped", want)
			if err := tx.Save(r); err != nil {
				return err
			}
		}
	}
	return nil
}
