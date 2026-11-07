package app

import (
	"regexp"

	"github.com/pocketbase/pocketbase/core"
)

// Birthdays go in the household's birthday category by themselves: an
// imported event whose title says "birthday" (in a few languages), or every
// event of a calendar named that way, unless it was imported into a trip or
// sitting category. The web app does the same while you type a new entry's
// title (web/src/lib/categories.ts).
var birthdayWords = regexp.MustCompile(`(?i)birth ?days?|\bb-?days?\b|verjaardag|geburtstag|cumpleaños|anniversaire|γενέθλια`)

// IsBirthday tells whether a title (or a calendar's name) is about a birthday.
func IsBirthday(title string) bool {
	return birthdayWords.MatchString(title)
}

// BirthdayCategory is the switched-on category birthdays go in: one named
// Birthday (or Birthdays), else one with the cake icon. It is "" when the
// household has none, or switched it off.
func BirthdayCategory(app core.App) string {
	categories, err := app.FindRecordsByFilter("categories", "enabled = true && kind = 'event'", "sort,created", 0, 0)
	if err != nil {
		return ""
	}
	for _, c := range categories {
		if IsBirthday(c.GetString("name")) {
			return c.Id
		}
	}
	for _, c := range categories {
		if c.GetString("icon") == "cake" {
			return c.Id
		}
	}
	return ""
}

// plainCategory tells whether a category is a plain one (not a trip or
// sitting), which a birthday may be moved out of.
func plainCategory(app core.App, id string) bool {
	if id == "" {
		return true
	}
	c, err := app.FindRecordById("categories", id)
	return err == nil && c.GetString("kind") == "event"
}
