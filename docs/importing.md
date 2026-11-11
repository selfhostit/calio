# Importing calendars

Calio can bring in events from the calendars you already use. Open **Settings > Import
calendars**. When the calendar is empty, the "or import from Google, Outlook or Apple
Calendar" link under **Add something** opens the same place.

There are two ways in:

- **Add a calendar link** keeps it up to date. Calio fetches the link right away and then
  every half hour. When an event changes there, its title, times, place and notes change in
  Calio too, and events deleted there go away. If you move an event (or retitle it) in Calio,
  that stays until the event changes in its own calendar. Anything you add in Calio stays: the sitter,
  checklists, shopping lists, comments, the category and who's going. Removing the calendar
  removes its events.
- **Import a file** copies the events in an `.ics` file once. Importing the same file again
  updates those events and doesn't add them twice.

Either way, you pick the category the events go in. Birthdays go in the **Birthday**
category by themselves: every event of a calendar named "Birthdays", and any event whose
title says birthday (or bday, verjaardag, Geburtstag, cumpleaños, anniversaire, γενέθλια).
That doesn't happen when you import into a trip or sitting category, or when Birthday is
switched off in Settings. Events that ended more than a year ago are left out.

| App                                      | Link (stays up to date)              | File (copied once)       |
| ---------------------------------------- | ------------------------------------ | ------------------------ |
| Google Calendar                          | Yes, the secret iCal address         | Yes, Import & export     |
| Outlook.com and Microsoft 365            | Yes, publish the calendar            | Outlook for Windows only |
| Apple Calendar (iCloud)                  | Yes, a public calendar link          | Calendar on a Mac        |
| Proton Calendar                          | Yes, a share link                    | Yes                      |
| Fastmail, Nextcloud, Thunderbird, others | Usually: an iCal, ICS or webcal link | Usually                  |

## Google Calendar

**Link.** This works on a computer only; the phone app can't do it.

1. Open [calendar.google.com](https://calendar.google.com), click the gear, then **Settings**.
2. Under **Settings for my calendars** on the left, click the calendar.
3. Scroll to **Integrate calendar** and copy **Secret address in iCal format**.
4. In Calio, choose **Add a calendar link** and paste it.

Anyone with the secret address can see that calendar. If it leaks, **Reset** next to it
makes a new one; add the new one to Calio and remove the old one.

**File.** In Settings, click **Import & export**, then **Export**. Google downloads a
`.zip` file with one `.ics` file for each calendar. Unzip it and import the one you want.

## Outlook.com and Microsoft 365

**Link.**

1. Open Outlook on the web ([outlook.com](https://outlook.live.com/calendar/) or
   [outlook.office.com](https://outlook.office.com/calendar/)) and click the gear.
2. Go to **Calendar > Shared calendars**.
3. Under **Publish a calendar**, pick the calendar and **Can view all details**, then
   **Publish**.
4. Copy the **ICS** link (not the HTML one) and add it in Calio.

At work, your organization may have turned publishing off. In that case, ask whoever runs
your Microsoft 365, or save a file from Outlook for Windows.

**File.** In Outlook for Windows, open the calendar, choose **File > Save Calendar**, and
under **More options** pick the dates and **Full details**. Outlook on the web and the new
Outlook can't save a file.

## Apple Calendar (iCloud)

**Link.** On an iPhone or iPad:

1. Open **Calendar** and tap **Calendars** at the bottom.
2. Tap the ⓘ next to the calendar and turn on **Public Calendar**.
3. Tap **Share Link**, then **Copy**, and add it in Calio.

On a Mac, Control-click the calendar in Calendar, choose **Sharing Settings**, tick
**Public Calendar** and copy the link. It starts with `webcal://`, which Calio accepts.

A public iCloud calendar can be read by anyone who has its link.

**File.** On a Mac, click the calendar in Calendar, then **File > Export > Export…**.

## Proton Calendar

**Link.** In Proton Calendar on the web, open **Settings > Calendars**, pick the calendar,
choose **Share** and create a link with **full details**. Copy it and add it in Calio.

**File.** In **Settings > Calendars**, pick the calendar and choose **Export**.

## Other apps

Most calendar apps and services can give an iCal, ICS or webcal link to "subscribe" to a
calendar (Fastmail, Nextcloud, Zoho, school and sports club calendars, public holiday
calendars), or export an `.ics` file. Both work in Calio.

## Notes

- Calio fetches calendar links from the server, so the server needs to reach the internet.
  A link that only works on your own computer won't work.
- Links are secrets. Calio never shows a link again after it's added, not even to you.
- Why links and not "Sign in with Google" or Microsoft? Signing in would need Calio
  registered as an app with Google and Microsoft, with keys every household would have to
  create. A calendar link works with every calendar app and needs no keys.
- Repeats Calio can show (every day, week, month or year) stay one entry. "Every Monday and
  Wednesday" becomes one weekly entry per day. Other repeats ("every other week", "the
  second Tuesday"), or a repeat with a skipped or moved time, become one entry per time,
  for the next two years.
- Time zones, including the Windows ones Outlook uses, are kept. Times without a time zone
  are read in the server's (`TZ`).
- Going the other way, from Calio into those apps: see **Settings > Calendar feed**.
