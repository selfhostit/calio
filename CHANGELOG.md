# Changelog

What changed in each release of calio. Installs update to the newest release; see
[docs/deploy.md](docs/deploy.md#updates-and-versions).

## Unreleased

- Calio: a shared household calendar for events, pet sitting and trips, with upcoming and
  month views, notifications for new entries and booked sitters, and a morning summary.
- Releases: a `vX.Y.Z` tag publishes `:X.Y.Z`, `:X.Y`, `:X` and `:latest`, and installs follow
  `:latest`, so they only update to a release. `main` publishes `:edge`.
- Every published image has a signed provenance attestation (`gh attestation verify`).
- Security and maintenance section in the README.
