package app

import (
	"strings"
	"unicode/utf8"

	"github.com/pocketbase/pocketbase/core"

	"github.com/selfhostit/calio/push"
)

// Comments on entries: whoever writes one is its author, and everyone else
// in the household hears about it.
func registerComments(app core.App) {
	app.OnRecordCreateRequest("comments").BindFunc(func(e *core.RecordRequestEvent) error {
		byID, byName := "", "Someone"
		if e.Auth != nil && e.Auth.Collection().Name == "users" {
			byID, byName = e.Auth.Id, displayName(e.Auth)
			e.Record.Set("author", byID)
		}
		if err := e.Next(); err != nil {
			return err
		}
		about := "an entry"
		if entry, err := e.App.FindRecordById("entries", e.Record.GetString("entry")); err == nil {
			about = entry.GetString("title")
		}
		push.ToEveryone(e.App, byID, push.Message{
			Title: byName + " on " + about,
			Body:  shorten(e.Record.GetString("text"), 160),
			Tag:   "comment-" + e.Record.GetString("entry"),
		})
		return nil
	})
}

// shorten cuts text to at most n characters, ending with "…" when it was longer.
func shorten(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:n-1])) + "…"
}
