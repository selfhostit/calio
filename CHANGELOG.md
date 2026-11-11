# Changelog

What changed in each release of calio. Installs update to the newest release; see
[docs/deploy.md](docs/deploy.md#updates-and-versions).

## Unreleased

- Moving things on the calendar is easier. In the month view on a computer, drag an entry to
  another day; in the form, change the start and the end moves with it, so a trip keeps its
  nights and a dinner its hours, or use −1 day, +1 day and +1 week. Each move ends with a
  note saying where it went, with Undo. The form shows how long something lasts and says
  when the end is before the start instead of failing to save.
- Moving an event that comes from a subscribed calendar now sticks. Before, the next
  half-hourly refresh put it back; now it follows that calendar again only when the event
  changes there.
- Your red panda notices the calendar: it packs a suitcase the days before a trip, sends
  postcards while you're away, plays with a visiting pet and puts on a party hat for
  birthdays. Below your habits, the household shelf shows everyone else's panda and how
  their day is going, and you can cheer them on once a day (they get a notification). An
  optional evening reminder at 8 pm tells you how many habits are left.
- Berries for your red panda: one for each habit done, three more for finishing the day,
  and a bonus at 7, 30 and 100 days in a row. Skipping a habit on purpose doesn't break
  the streak. Spend them in the new Shop on outfits (hats, a flower crown, scarves, a bow
  tie, glasses), which you can put on and take off. Your panda also grows from a cub to a
  grown-up as you keep up your habits.
- A red panda of your own, on the new Panda tab. Name it, pick a few daily habits (brush
  teeth, vitamin D, eat healthy, or your own, on the days and times of day you choose) and
  tick them off. Its mood follows the last few days: it glows when you keep up, gets a bit
  droopy when you don't, and one good day cheers it up. Tap it to pet it. On phones, the
  tabs you're not on show just their icon.
- The app image installs Alpine's latest security fixes when it is built, which brings in the
  patched zlib (CVE-2026-4738).
- A Birthday category (pink, with a cake), switched on for every household. New entries
  and imported events titled "... birthday" go in it by themselves, and new ones are all
  day and repeat every year. A household that already made a birthday category (by name, like Birthdays or
  Verjaardagen, or with the cake icon) keeps that one instead. Birthdays already in the
  Event category move to it.
- Import calendars: add a Google Calendar, Outlook, iCloud or other calendar by its link,
  and Calio keeps it up to date every half hour; or import an `.ics` file once. Settings
  shows how for each app, and so does [docs/importing.md](docs/importing.md).
- Fixed: the add and edit form didn't open when someone in the household had no name
  set, because the app tried to use their email, which only they can see.

## 0.2.0 (2026-11-05)

- Categories: choose which ones the household uses (Event, Pet sitting, Trip,
  Babysitting) and add your own with a color and an icon, in Settings.
- Comments on entries, which notify everyone else.
- Documents replace contacts: pages with notes, photos and files about the dog, the house
  or the kids, shareable with a sitter through a private link. Existing contacts become
  documents.
- Shopping lists on entries, which can be sent to Shoppa.
- A logo: a house-shaped calendar.
- The installer takes `PORT` and picks the next free port.
- golang.org/x/net 0.62.0, which fixes GO-2026-6860.
- SECURITY.md: how to report a problem privately, and how quickly fixes ship.
- Security scans on every pull request and every Monday: CodeQL, govulncheck, Grype
  (web and worker dependencies, released images), dependency review and gitleaks.
- The containers' filesystems are read-only apart from `pb_data` and `/tmp`.
- The compose file follows the major version (`:0`) instead of `:latest`, and the
  installer fetches the compose files of the newest release instead of `main`.

## 0.1.0 (2026-11-04)

- Calio: a shared household calendar for events, pet sitting and trips, with upcoming and
  month views, notifications for new entries and booked sitters, and a morning summary.
- Contacts for the vet, pet sitters, babysitters and others, with notes.
- A calendar feed (iCalendar) per person, for Apple Calendar, Outlook and Google Calendar.
- Repeating entries, checklists (packing lists on trips) and Google Maps links for places.
- Releases: a `vX.Y.Z` tag publishes `:X.Y.Z`, `:X.Y`, `:X` and `:latest`, and installs follow
  `:latest`, so they only update to a release. `main` publishes `:edge`.
- Every published image has a signed provenance attestation (`gh attestation verify`).
- Security and maintenance section in the README.
