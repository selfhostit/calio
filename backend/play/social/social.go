// Package social is the household side of companions: cheering someone on.
package social

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"github.com/selfhostit/calio/play/habits"
	"github.com/selfhostit/calio/push"
)

// Cheers is the collection, from migration 1763500000.
const Cheers = "cheers"

// Register serves POST /api/app/play/cheer/{user}.
func Register(app core.App) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/api/app/play/cheer/{user}", cheer).Bind(apis.RequireAuth("users"))
		return se.Next()
	})
}

// cheer cheers someone in the household on, once a day, and tells them.
func cheer(e *core.RequestEvent) error {
	to := e.Request.PathValue("user")
	if to == e.Auth.Id {
		return e.BadRequestError("Cheering yourself on is lovely, but try someone else.", nil)
	}
	pal, err := e.App.FindFirstRecordByData("companions", "user", to)
	if err != nil {
		return e.NotFoundError("They don't have a companion yet.", nil)
	}
	c, err := e.App.FindCachedCollectionByNameOrId(Cheers)
	if err != nil {
		return e.InternalServerError("", err)
	}
	r := core.NewRecord(c)
	r.Load(map[string]any{"from": e.Auth.Id, "to": to, "day": habits.Today()})
	if err := e.App.Save(r); err != nil {
		return e.BadRequestError("You cheered them on today already.", nil)
	}
	push.ToUser(e.App, to, push.Message{
		Title: firstName(e.Auth) + " cheered you on",
		Body:  pal.GetString("name") + " is very proud of you.",
		Tag:   "cheer",
	})
	return e.JSON(http.StatusOK, r)
}

func firstName(u *core.Record) string {
	if f := strings.Fields(u.GetString("name")); len(f) > 0 {
		return f[0]
	}
	return "Someone"
}
