// Package play is the companion side of Calio: habits, the berries they earn,
// the shop, cheering each other on and evening reminders. Each part is its own package; habits tells the others when
// a habit is ticked off (habits.Module.OnCheck), so a new part listens
// rather than reaching into the others.
package play

import (
	"github.com/pocketbase/pocketbase/core"

	"github.com/selfhostit/calio/play/habits"
	"github.com/selfhostit/calio/play/nudges"
	"github.com/selfhostit/calio/play/rewards"
	"github.com/selfhostit/calio/play/shop"
	"github.com/selfhostit/calio/play/social"
)

// Register wires up the parts.
func Register(app core.App) {
	h := habits.Register(app)
	rewards.Register(app, h)
	shop.Register(app)
	social.Register(app)
	nudges.Register(app)
}
