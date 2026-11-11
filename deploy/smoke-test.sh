#!/bin/sh
# End-to-end check of a deploy: starts the app and a worker with
# docker-compose.yml in a scratch folder, signs up the owner, creates a worker
# account, runs one worker round and checks that the worker reported in. Then
# it adds an entry, moves it to another day, restarts everything and checks
# the move is still there: data lives in ./pb_data, not in the container.
#
#   CALIO_IMAGE=calio:dev CALIO_WORKER_IMAGE=calio-worker:dev deploy/smoke-test.sh
#
# Set EXPECT_VERSION to also check the version the app reports.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
dir=$(mktemp -d)
port=${PORT:-18090}
url=http://127.0.0.1:$port

cleanup() {
	status=$?
	if [ "$status" -ne 0 ]; then (cd "$dir" && docker compose logs) || true; fi
	(cd "$dir" && docker compose --profile worker down -v) >/dev/null 2>&1 || true
	rm -rf "$dir"
	exit "$status"
}
trap cleanup EXIT

fail() { echo "FAIL: $*" >&2; exit 1; }
step() { echo "--- $*"; }

cp "$here/docker-compose.yml" "$dir/"
mkdir "$dir/pb_data"
cat >"$dir/.env" <<ENV
CALIO_IMAGE=${CALIO_IMAGE:-ghcr.io/selfhostit/calio:latest}
CALIO_WORKER_IMAGE=${CALIO_WORKER_IMAGE:-ghcr.io/selfhostit/calio-worker:latest}
PUID=$(id -u)
PGID=$(id -g)
PORT=$port
CALIO_WORKER=smoke
CALIO_WORKER_PASSWORD=smoke-worker-pw
ENV
cd "$dir"

step "Starting the app"
docker compose up -d app
for _ in $(seq 1 30); do curl -sf "$url/api/health" >/dev/null && break; sleep 1; done
curl -sf "$url/api/health" >/dev/null || fail "the app did not become healthy"

step "Checking what it serves"
if [ -n "${EXPECT_VERSION:-}" ]; then
	curl -sf "$url/api/app/info" | grep -q "\"version\":\"$EXPECT_VERSION\"" || fail "wrong version"
fi
curl -sf "$url/" | grep -q '<div id="app">' || fail "no web app"
curl -sf "$url/manifest.webmanifest" | grep -q '"display": "standalone"' || fail "no manifest"
curl -sf "$url/api/app/setup" | grep -q '"needed":true' || fail "setup not offered"

step "Signing up the owner"
code=$(cat pb_data/setup_code 2>/dev/null)
[ -n "$code" ] || fail "no setup code in pb_data"
curl -s -H 'Content-Type: application/json' "$url/api/app/setup" \
	-d '{"name":"Eve","email":"eve@example.com","password":"eve-test-pw","code":"WRONG-CODE"}' |
	grep -q 'setup code is wrong' || fail "setup took a wrong code"
curl -sf -H 'Content-Type: application/json' "$url/api/app/setup" \
	-d "{\"name\":\"Smoke\",\"email\":\"smoke@example.com\",\"password\":\"smoke-test-pw\",\"code\":\"$code\"}" >/dev/null ||
	fail "setup failed"
token=$(curl -sf -H 'Content-Type: application/json' \
	"$url/api/collections/_superusers/auth-with-password" \
	-d '{"identity":"smoke@example.com","password":"smoke-test-pw"}' |
	sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
[ -n "$token" ] || fail "the owner cannot sign in to the admin UI"

step "Creating a worker account and running one round of the worker"
curl -sf -H "Authorization: $token" -H 'Content-Type: application/json' \
	"$url/api/collections/workers/records" \
	-d '{"name":"smoke","password":"smoke-worker-pw","passwordConfirm":"smoke-worker-pw"}' >/dev/null ||
	fail "could not create the worker account"
docker compose run --rm worker python worker.py --once

curl -sf -H "Authorization: $token" "$url/api/collections/workers/records" |
	grep -q '"last_seen":"2' || fail "the worker did not report in"

step "Moving an entry and restarting"
user=$(curl -sf -H 'Content-Type: application/json' \
	"$url/api/collections/users/auth-with-password" \
	-d '{"identity":"smoke@example.com","password":"smoke-test-pw"}' |
	sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
[ -n "$user" ] || fail "the owner cannot sign in to the app"
id=$(curl -sf -H "Authorization: $user" -H 'Content-Type: application/json' \
	"$url/api/collections/entries/records" \
	-d '{"title":"Dinner","kind":"event","all_day":false,"start":"2099-12-05 17:30:00.000Z","end":"2099-12-05 20:00:00.000Z"}' |
	sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
[ -n "$id" ] || fail "could not add an entry"
curl -sf -X PATCH -H "Authorization: $user" -H 'Content-Type: application/json' \
	"$url/api/collections/entries/records/$id" \
	-d '{"start":"2099-12-06 17:30:00.000Z","end":"2099-12-06 20:00:00.000Z"}' >/dev/null ||
	fail "could not move the entry"
docker compose down
docker compose up -d app
for _ in $(seq 1 30); do curl -sf "$url/api/health" >/dev/null && break; sleep 1; done
curl -sf -H "Authorization: $user" "$url/api/collections/entries/records/$id" |
	grep -q '"start":"2099-12-06 17:30:00.000Z"' || fail "the move was lost in the restart"

echo "OK"
