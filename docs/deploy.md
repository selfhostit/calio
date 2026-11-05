# Deploy

A calio app runs as one container on any Linux machine with Docker: a NAS, a
Raspberry Pi, a small server. Workers run next to it or on other machines.
Tailscale gives your devices an `https://` address for it without opening
anything to the internet.

```
 phone / laptop ──Tailscale──▶ tailscale sidecar ──▶ app (PocketBase + web app) ── pb_data/
                                                       ▲
                     worker on the same machine ───────┤  (records through the API)
                     worker on a Pi ──Tailscale────────┘
```

Two images are published for amd64, arm64 and armv7: `ghcr.io/selfhostit/calio` (the
app) and `ghcr.io/selfhostit/calio-worker` (the Python workers). A release tag like
`v1.2.0` publishes `:1.2.0`, `:1.2`, `:1` and `:latest` of both; installs follow `:latest`,
so they only update to a release. Every push to `main` publishes `:edge` and the commit
hash, for trying out what's next. Each image comes with a signed
[provenance attestation](#check-where-an-image-came-from).

## Install

On the machine, as a user that can run `docker`:

```sh
curl -fsSL https://raw.githubusercontent.com/selfhostit/calio/main/deploy/install.sh | sh
```

This creates `~/calio` with the compose files, a `.env` and a `pb_data` folder
for the database, asks for a Tailscale auth key (press Enter to skip, or set an
empty `TS_AUTHKEY=` to skip the question), and starts the app. Pass a folder to
install elsewhere: `... | sh -s /srv/calio`.

Then open the app and create your account with the setup code the installer shows
(it is also in `docker compose logs app`); you become the owner. Only someone who can
see the machine has the code, so finish setup before or after making the app public,
whichever suits you.
Without Tailscale it is only reachable on the machine itself, at
http://localhost:8090 (the port is bound to `127.0.0.1` on purpose).

Settings live in `.env`: time zone, port, Tailscale, the worker account and the
image version. Change them and run `docker compose up -d` in the folder (or the
install script again).

## Remote access with Tailscale

1. In the [Tailscale admin console](https://login.tailscale.com/admin/dns), turn on
   **MagicDNS** and **HTTPS Certificates**.
2. Create an [auth key](https://login.tailscale.com/admin/settings/keys) and paste it
   when the install script asks. It is only used for the first sign-in; the
   sidecar keeps its login in `tailscale/`.
3. The app is now at `https://calio.<your-tailnet>.ts.net` (the script prints it).
   The first visit can take a few seconds while the certificate is issued.

On a phone, install the Tailscale app, open that address and add it to the home
screen. Push notifications need this HTTPS address.

To add Tailscale later, set `TS_AUTHKEY` and uncomment the `COMPOSE_FILE` line in
`.env`, then run `docker compose up -d`. `TS_HOSTNAME` sets the name.

Nothing is open to the internet: only devices signed in to your tailnet can reach
the app, and the port on the machine is bound to `127.0.0.1`. To let someone else
in the household use it, [share the machine](https://tailscale.com/kb/1084/sharing)
or invite them to the tailnet.

### Calendar feed and shared documents

The app has a calendar feed (an `.ics` link the app shows you) that calendar apps
subscribe to. Apple Calendar, Outlook, Thunderbird and DAVx⁵ fetch it from your
own device, so they work over Tailscale with the normal address.

Google Calendar is different: Google's servers fetch the feed, and they are not on
your tailnet. The same goes for a document you share with a sitter (a page about the
dog or the house): their phone isn't on your tailnet either. For both, Calio can
publish **only the feed and shared documents** to the internet with
[Tailscale Funnel](https://tailscale.com/kb/1223/funnel). It runs a second, small
Tailscale node, `https://calio-feed.<your-tailnet>.ts.net`, whose only route is
`/api/app/feed/` and `/api/app/share/`; every other path there answers 404, so
sign-in, the app and the admin UI stay tailnet-only. Feed and document links have a
long random token in them, which is the only thing that protects them, so treat them
like passwords: reset the feed link, or stop sharing the document, if one leaks.
Skip this if you use neither Google Calendar nor shared documents.

1. In the [access controls](https://login.tailscale.com/admin/acls/file), allow
   Funnel. The admin console offers to add it the first time; by hand, add this to
   the policy file:
   ```json
   "nodeAttrs": [{ "target": ["autogroup:member"], "attr": ["funnel"] }]
   ```
2. Create a second [auth key](https://login.tailscale.com/admin/settings/keys) (or
   make the first one reusable) and set it in `.env`:
   ```sh
   COMPOSE_FILE=docker-compose.yml:docker-compose.tailscale.yml:docker-compose.funnel.yml
   TS_FEED_AUTHKEY=tskey-auth-...
   ```
3. Run `docker compose up -d`. The feed node's address is the app's address with
   `calio-feed` in place of `calio` (it is also listed under Machines in the admin
   console). Set it so the app shows public links:
   ```sh
   CALIO_FEED_URL=https://calio-feed.<your-tailnet>.ts.net
   ```
   and run `docker compose up -d` again.
4. In Google Calendar, go to Other calendars > **+** > From URL and paste the feed
   link from the app. Google refreshes it every few hours, not instantly.

Check it from a device that is *not* on your tailnet (a phone on mobile data with
Tailscale off): the feed link downloads a calendar, and
`https://calio-feed.<your-tailnet>.ts.net/` shows a 404. Public DNS for a new Funnel
name can take up to ten minutes. To turn it off, remove `:docker-compose.funnel.yml`
from `COMPOSE_FILE`, run `docker compose up -d --remove-orphans`, and remove the
`calio-feed` machine in the admin console.

### Shoppa

[Shoppa](https://github.com/selfhostit/shoppa), the household shopping list, can take
the shopping lists on Calio entries ("Send to Shoppa"). On a machine that runs both
with Docker Compose:

1. In Shoppa's admin UI (`/_/`), open the **workers** collection and add one, named
   `calio`, with a password.
2. In Calio's `.env`, append `:docker-compose.shoppa.yml` to `COMPOSE_FILE` and set:
   ```sh
   SHOPPA_WORKER=calio
   SHOPPA_PASSWORD=...
   SHOPPA_LIST=Groceries   # optional; the first list otherwise
   ```
3. Run `docker compose up -d`.

The app joins Shoppa's Docker network, `shoppa_default`, and reaches it at
`http://shoppa-app-1:8090`; if Shoppa's folder has another name, set
`SHOPPA_NETWORK` and `SHOPPA_URL` to match (`docker network ls` lists the networks).
Shoppa elsewhere works too: leave out the compose file and set `SHOPPA_URL` to any
address the app can reach. Shoppa keeps one item per thing, so sending a list twice
adds nothing new.

## Updates and versions

```sh
cd ~/calio && docker compose pull && docker compose up -d
```

Running the install script again does the same and also refreshes the compose
files. Updates bring the newest release, listed with what changed in
[CHANGELOG.md](../CHANGELOG.md). To stay on a release line, or go back to a release, set
`CALIO_IMAGE` (and `CALIO_WORKER_IMAGE`) in `.env` to a tag, like
`ghcr.io/selfhostit/calio:1` (every 1.x) or `ghcr.io/selfhostit/calio:1.2.0`. Use `:edge`
to follow `main`. Database migrations run on start, so take a backup before going back
across one.

### Check where an image came from

Every published image has a signed attestation that says which workflow run, in which
repository and at which commit, built it. With the [GitHub CLI](https://cli.github.com):

```sh
gh attestation verify oci://ghcr.io/selfhostit/calio:latest --repo selfhostit/calio
```

It fails if the image was not built by this repository's CI, for example if someone pushed one by hand.

### Releasing

Add what changed under a new version in `CHANGELOG.md`, merge, then publish a GitHub
release with a tag like `v1.2.0` (Releases > Draft a new release) and paste the changelog
entry as its notes. CI tests it, publishes the images and attests them.

## Backups and restore

The app backs up the database and uploads every night at 03:00 and keeps the
last 7, in `pb_data/backups`. A backup on the same disk doesn't survive the disk,
so copy `pb_data` somewhere else too (Hyper Backup on a Synology, `rsync`, or S3
storage set in the admin UI under Settings > Backups). Encrypt any copy that leaves
the machine, and leave out the `tailscale` and `tailscale-feed` folders: they hold the machine's Tailscale
keys, which let whoever has them join your tailnet as those machines.

To restore, open the admin UI (`/_/`), go to Settings > Backups and restore one.
To move to another machine, stop the app, copy the whole folder, and start it
there.

## Workers

### Next to the app

1. In the admin UI, open the **workers** collection and add one with a name and a
   password.
2. In `.env`, set `CALIO_WORKER` and `CALIO_WORKER_PASSWORD` and uncomment
   `COMPOSE_PROFILES=worker`.
3. Run `docker compose up -d`.

It runs `workers/worker.py`, the starting point for your own worker. To run another
one, set `CALIO_WORKER_SCRIPT` in `.env`, for example to `camera.py`. Every `.py`
file in `workers/` and every dependency you add with `uv add` is in the worker image.

### On another machine

Join the machine to your tailnet, then point the worker at the app's tailnet
address:

```sh
docker run -d --name worker --restart unless-stopped \
  -e CALIO_URL=https://calio.<your-tailnet>.ts.net \
  -e CALIO_WORKER=camera -e CALIO_WORKER_PASSWORD=... -e TZ=Europe/Athens \
  ghcr.io/selfhostit/calio-worker python camera.py
```

## Synology NAS

Needs DSM 7.2 or later with **Container Manager** installed.

1. In Control Panel > Terminal & SNMP, enable SSH, then sign in with
   `ssh you@nas.local`.
2. Install into a shared folder, with `sudo` because Docker on DSM needs it:
   ```sh
   curl -fsSL https://raw.githubusercontent.com/selfhostit/calio/main/deploy/install.sh | sudo sh -s /volume1/docker/calio
   ```
   The data stays owned by your DSM user (`PUID` and `PGID` in `.env`), so you can
   see and back it up in File Station.
3. The containers show up in Container Manager, where you can see their logs and
   restart them. Update from SSH as above.

The Tailscale sidecar runs in its own container, so it works whether or not the
Tailscale package from Package Center is installed on the NAS; the NAS and the app
show up as two machines in your tailnet. If you'd rather use the NAS's own Tailscale
package and no sidecar, leave `COMPOSE_FILE` commented out and run, over SSH:

```sh
sudo tailscale serve --bg http://127.0.0.1:8090
```

The app is then at `https://<nas-name>.<your-tailnet>.ts.net`. Funnel for the feed
still works with the feed sidecar alone: set
`COMPOSE_FILE=docker-compose.yml:docker-compose.funnel.yml`. Don't turn on Funnel for
the NAS itself, because that would publish the whole app and DSM.

To serve the app on the internet instead of only your tailnet, add a rule under
Control Panel > Login Portal > Advanced > Reverse Proxy from your HTTPS host name
to `localhost:8090`. Never do this for apps that show cameras.

## Raspberry Pi

Use a Pi 4 or 5 with the 64-bit Raspberry Pi OS (Lite is enough).

```sh
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER   # then sign out and back in
```

Then run the install script as above to host the app on the Pi, or follow
[On another machine](#on-another-machine) to use the Pi only for workers, with
Tailscale installed on the Pi itself
(`curl -fsSL https://tailscale.com/install.sh | sh && sudo tailscale up`).

Hardware a worker needs, like a Hailo accelerator or a camera, is passed to its
container with `devices:` in compose (or `--device /dev/hailo0` with `docker run`).
Large services such as Frigate run as their own containers next to the worker.

## Check a deploy

`make smoke` builds both images, deploys them with `docker-compose.yml` in a
scratch folder, signs up an owner, runs one round of the worker and checks that
it reported in. CI runs the same script on every pull request.
