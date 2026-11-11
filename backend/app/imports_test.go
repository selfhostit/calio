package app_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"

	"github.com/selfhostit/calio/app"
)

// a calendar with events well in the future, so none is left out as past
const familyICS = "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nX-WR-CALNAME:Family\r\n" +
	"BEGIN:VEVENT\r\nUID:trip-1\r\nSUMMARY:Crete\r\nDTSTART;VALUE=DATE:20991212\r\nDTEND;VALUE=DATE:20991219\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:dinner-1\r\nSUMMARY:Dinner\r\nLOCATION:Maria's\r\nDTSTART:20991205T173000Z\r\nDTEND:20991205T200000Z\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:piano-1\r\nSUMMARY:Piano\r\nDTSTART:20991202T153000Z\r\nDTEND:20991202T163000Z\r\nRRULE:FREQ=WEEKLY;UNTIL=21000630T000000Z\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:swim-1\r\nSUMMARY:Swimming\r\nDTSTART:20991201T160000Z\r\nDTEND:20991201T170000Z\r\nRRULE:FREQ=WEEKLY;BYDAY=TU,TH\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:old-1\r\nSUMMARY:Long ago\r\nDTSTART;VALUE=DATE:20100101\r\nEND:VEVENT\r\n" +
	"END:VCALENDAR\r\n"

func upload(t testing.TB, fields map[string]string, file string) (io.Reader, string) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	if file != "" {
		f, err := w.CreateFormFile("file", "family.ics")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = f.Write([]byte(file))
	}
	_ = w.Close()
	return &body, w.FormDataContentType()
}

// server answers requests with one app, kept between them.
func server(t testing.TB, a core.App) func(method, url string, body io.Reader, headers map[string]string) (int, string) {
	r, err := apis.NewRouter(a)
	if err != nil {
		t.Fatal(err)
	}
	se := &core.ServeEvent{App: a, Router: r}
	if err := a.OnServe().Trigger(se, func(*core.ServeEvent) error { return nil }); err != nil {
		t.Fatal(err)
	}
	mux, err := r.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	return func(method, url string, body io.Reader, headers map[string]string) (int, string) {
		req := httptest.NewRequest(method, url, body)
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}
}

func expect(t testing.TB, status int, body string, wantStatus int, content ...string) {
	t.Helper()
	if status != wantStatus {
		t.Errorf("status %d, want %d: %s", status, wantStatus, body)
	}
	for _, c := range content {
		if !strings.Contains(body, c) {
			t.Errorf("no %s in %s", c, body)
		}
	}
}

func entriesByTitle(t testing.TB, a core.App) map[string]*core.Record {
	all, err := a.FindAllRecords("entries")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]*core.Record{}
	for _, r := range all {
		out[r.GetString("title")] = r
	}
	return out
}

func TestImportFile(t *testing.T) {
	a := newApp(t)
	u, token := addUser(t, a, "ada@example.com", "member")
	trips, err := a.FindFirstRecordByFilter("categories", "kind = 'trip'")
	if err != nil {
		t.Fatal(err)
	}

	call := server(t, a)
	post := func(fields map[string]string, file string, status int, content ...string) {
		t.Helper()
		body, ct := upload(t, fields, file)
		code, out := call(http.MethodPost, "/api/app/import", body, map[string]string{"Authorization": token, "Content-Type": ct})
		expect(t, code, out, status, content...)
	}

	post(nil, "", http.StatusBadRequest, "Choose a calendar file")
	post(nil, "<html>not a calendar</html>", http.StatusBadRequest, "isn't a calendar")

	post(map[string]string{"category": trips.Id}, familyICS, http.StatusOK, `"added":5`)
	got := entriesByTitle(t, a)
	if _, ok := got["Long ago"]; ok {
		t.Error("an event from 2010 was imported")
	}
	trip := got["Crete"]
	if trip == nil || !trip.GetBool("all_day") || trip.GetString("kind") != "trip" ||
		trip.GetString("created_by") != u.Id || trip.GetString("category") != trips.Id {
		t.Fatalf("trip = %v", trip)
	}
	// the end is the last day, not the day after
	if s, e := trip.GetDateTime("start").String(), trip.GetDateTime("end").String(); s != "2099-12-12 00:00:00.000Z" || e != "2099-12-18 00:00:00.000Z" {
		t.Errorf("trip runs %s to %s", s, e)
	}
	if got["Dinner"].GetString("location") != "Maria's" || got["Dinner"].GetBool("all_day") {
		t.Errorf("dinner = %v", got["Dinner"])
	}
	if p := got["Piano"]; p.GetString("repeat") != "weekly" || p.GetDateTime("repeat_until").String() != "2100-06-30 00:00:00.000Z" {
		t.Errorf("piano repeats %q until %s", p.GetString("repeat"), p.GetDateTime("repeat_until"))
	}
	// "every Tuesday and Thursday" becomes a weekly entry for each day
	swims, _ := a.FindRecordsByFilter("entries", "title = 'Swimming'", "start", 0, 0)
	if len(swims) != 2 || swims[1].GetDateTime("start").String() != "2099-12-03 16:00:00.000Z" {
		t.Errorf("swimming = %v", swims)
	}

	// importing it again changes nothing, and keeps what was added in Calio
	trip.Set("sitter", "Nikos")
	if err := a.Save(trip); err != nil {
		t.Fatal(err)
	}
	post(nil, familyICS, http.StatusOK, `"added":0`, `"updated":0`)
	post(nil, strings.Replace(familyICS, "SUMMARY:Dinner", "SUMMARY:Dinner at eight", 1), http.StatusOK, `"added":0`, `"updated":1`)
	got = entriesByTitle(t, a)
	if got["Crete"].GetString("sitter") != "Nikos" || got["Dinner at eight"] == nil {
		t.Error("a second import lost a change")
	}
}

// Moving an imported entry in Calio sticks until the event changes in its
// own calendar; then that calendar wins.
func TestMovedImportStays(t *testing.T) {
	a := newApp(t)
	_, token := addUser(t, a, "ada@example.com", "member")
	call := server(t, a)
	post := func(file string, content ...string) {
		t.Helper()
		body, ct := upload(t, nil, file)
		code, out := call(http.MethodPost, "/api/app/import", body, map[string]string{"Authorization": token, "Content-Type": ct})
		expect(t, code, out, http.StatusOK, content...)
	}
	post(familyICS, `"added":5`)

	dinner := entriesByTitle(t, a)["Dinner"]
	dinner.Set("start", "2099-12-06 17:30:00.000Z")
	dinner.Set("end", "2099-12-06 20:00:00.000Z")
	if err := a.Save(dinner); err != nil {
		t.Fatal(err)
	}
	post(familyICS, `"updated":0`)
	if s := entriesByTitle(t, a)["Dinner"].GetDateTime("start").String(); s != "2099-12-06 17:30:00.000Z" {
		t.Errorf("the moved dinner went back to %s", s)
	}

	post(strings.Replace(familyICS, "DTSTART:20991205T173000Z", "DTSTART:20991205T180000Z", 1), `"updated":1`)
	if s := entriesByTitle(t, a)["Dinner"].GetDateTime("start").String(); s != "2099-12-05 18:00:00.000Z" {
		t.Errorf("the dinner changed in its calendar starts at %s", s)
	}
}

func TestImportNeedsSignIn(t *testing.T) {
	body, ct := upload(t, nil, familyICS)
	(&tests.ApiScenario{
		Method: http.MethodPost, URL: "/api/app/import", Body: body,
		Headers:        map[string]string{"Content-Type": ct},
		ExpectedStatus: http.StatusUnauthorized, ExpectedContent: []string{`"data":{}`},
		TestAppFactory: func(t testing.TB) *tests.TestApp { return newApp(t) },
	}).Test(t)
}

func TestSubscribe(t *testing.T) {
	var feed atomic.Value
	feed.Store(familyICS)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/family.ics" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = io.WriteString(w, feed.Load().(string))
	}))
	defer srv.Close()

	a := newApp(t)
	u, token := addUser(t, a, "ada@example.com", "member")
	call := server(t, a)
	auth := map[string]string{"Authorization": token}
	create := func(link string, status int, content ...string) {
		t.Helper()
		code, out := call(http.MethodPost, "/api/app/calendars",
			strings.NewReader(`{"url":"`+link+`"}`), auth)
		expect(t, code, out, status, content...)
		if strings.Contains(out, srv.URL) || strings.Contains(out, "secret-link") {
			t.Error("the secret link was shown")
		}
	}

	create("ftp://example.com/x.ics", http.StatusBadRequest, "calendar link")
	create("https://127.0.0.1:1/secret-link.ics", http.StatusOK, `"error":"couldn't reach the calendar`)
	create(srv.URL+"/gone.ics", http.StatusOK, `"error":"the calendar isn't there any more`)

	create(srv.URL+"/family.ics", http.StatusOK, `"name":"Family"`, `"events":5`, `"error":""`, `"created_by":"`+u.Id+`"`)
	c, err := a.FindFirstRecordByFilter("calendars", "events = 5")
	if err != nil {
		t.Fatal(err)
	}
	if n := len(entriesByTitle(t, a)); n != 4 { // swimming twice
		t.Fatalf("%d titles after subscribing", n)
	}

	// a refresh follows the calendar: changed events change, gone ones go
	feed.Store(strings.Replace(strings.Replace(familyICS, "SUMMARY:Dinner", "SUMMARY:Dinner moved", 1),
		"UID:trip-1", "UID:trip-2", 1))
	code, out := call(http.MethodPost, "/api/app/calendars/"+c.Id+"/refresh", nil, auth)
	expect(t, code, out, http.StatusOK, `"events":5`, `"error":""`)
	got := entriesByTitle(t, a)
	if got["Dinner moved"] == nil || got["Dinner"] != nil {
		t.Error("the changed dinner wasn't updated")
	}
	trips, _ := a.FindRecordsByFilter("entries", "title = 'Crete'", "", 0, 0)
	if len(trips) != 1 || trips[0].GetString("uid") != "trip-2" {
		t.Errorf("trips = %v", trips)
	}

	// removing the subscription removes its entries
	c, _ = a.FindRecordById("calendars", c.Id)
	if err := a.Delete(c); err != nil {
		t.Fatal(err)
	}
	left, _ := a.FindRecordsByFilter("entries", "calendar = {:c}", "", 0, 0, map[string]any{"c": c.Id})
	if len(left) != 0 {
		t.Errorf("%d entries left after unsubscribing", len(left))
	}
}

func TestBirthdaysGoInTheirCategory(t *testing.T) {
	a := newApp(t)
	_, token := addUser(t, a, "ada@example.com", "member")
	birthday := app.BirthdayCategory(a)
	if birthday == "" {
		t.Fatal("no birthday category to start with")
	}
	call := server(t, a)
	ics := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\n" +
		"BEGIN:VEVENT\r\nUID:b1\r\nSUMMARY:Maria's Birthday\r\nDTSTART;VALUE=DATE:20991128\r\nRRULE:FREQ=YEARLY\r\nEND:VEVENT\r\n" +
		"BEGIN:VEVENT\r\nUID:d1\r\nSUMMARY:Dentist\r\nDTSTART:20991205T090000Z\r\nEND:VEVENT\r\n" +
		"BEGIN:VEVENT\r\nUID:t1\r\nSUMMARY:Birthday trip\r\nDTSTART;VALUE=DATE:20991210\r\nEND:VEVENT\r\n" +
		"END:VCALENDAR\r\n"
	post := func(category, file string) {
		t.Helper()
		body, ct := upload(t, map[string]string{"category": category}, file)
		code, out := call(http.MethodPost, "/api/app/import", body, map[string]string{"Authorization": token, "Content-Type": ct})
		expect(t, code, out, http.StatusOK)
	}

	post("", ics)
	got := entriesByTitle(t, a)
	if got["Maria's Birthday"].GetString("category") != birthday {
		t.Error("the birthday didn't go in the birthday category")
	}
	if got["Dentist"].GetString("category") == birthday {
		t.Error("the dentist went in the birthday category")
	}

	// imported into a trip category, nothing moves
	trips, _ := a.FindFirstRecordByFilter("categories", "kind = 'trip'")
	post(trips.Id, strings.ReplaceAll(ics, "UID:", "UID:again-"))
	trip, _ := a.FindFirstRecordByFilter("entries", "uid = 'again-t1'")
	if trip.GetString("category") != trips.Id {
		t.Error("a birthday trip left the trip category")
	}

	// a calendar named Birthdays goes there whole
	post("", strings.Replace(strings.ReplaceAll(ics, "UID:", "UID:cal-"), "VERSION:2.0", "VERSION:2.0\r\nX-WR-CALNAME:Birthdays", 1))
	dentist, _ := a.FindFirstRecordByFilter("entries", "uid = 'cal-d1'")
	if dentist.GetString("category") != birthday {
		t.Error("an event of the Birthdays calendar went elsewhere")
	}
}

func TestIsBirthday(t *testing.T) {
	for title, want := range map[string]bool{
		"Maria's birthday": true, "BIRTHDAYS": true, "Leo's bday": true, "B-day party": true,
		"Verjaardag oma": true, "Γενέθλια Νίκου": true,
		"Dentist": false, "Bdaynamic meeting": false,
	} {
		if got := app.IsBirthday(title); got != want {
			t.Errorf("IsBirthday(%q) = %v", title, got)
		}
	}
}
