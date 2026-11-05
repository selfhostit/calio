# Changelog

What changed in each release of calio. Installs update to the newest release; see
[docs/deploy.md](docs/deploy.md#updates-and-versions).

## Unreleased

- SECURITY.md: how to report a problem privately, and how quickly fixes ship.
- Security scans on every pull request and every Monday: CodeQL, Grype (dependencies and
  released images), dependency review and gitleaks.
- The containers' filesystems are read-only apart from `pb_data` and `/tmp`.
- The compose file follows the major version (`:0`) instead of `:latest`, and the
  installer fetches the compose files of the newest release instead of `main`.
- Categories: choose which ones the household uses (Event, Pet sitting, Trip,
  Babysitting) and add your own with a color and an icon, in Settings.
- Comments on entries, which notify everyone else.
- Documents replace contacts: pages with notes, photos and files about the dog, the house
  or the kids, shareable with a sitter through a private link. Existing contacts become
  documents.
- Shopping lists on entries, which can be sent to Shoppa.
- Contacts for the vet, pet sitters, babysitters and others, with notes.
- A calendar feed (iCalendar) per person, for Apple Calendar, Outlook and Google Calendar.
- Repeating entries, checklists (packing lists on trips) and Google Maps links for places.
- Calio: a shared household calendar for events, pet sitting and trips, with upcoming and
  month views, notifications for new entries and booked sitters, and a morning summary.
- Releases: a `vX.Y.Z` tag publishes `:X.Y.Z`, `:X.Y`, `:X` and `:latest`, and installs follow
  `:latest`, so they only update to a release. `main` publishes `:edge`.
- Every published image has a signed provenance attestation (`gh attestation verify`).
- Security and maintenance section in the README.
