package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

func TestDocuments(t *testing.T) {
	t.Run("a member writes one about the dog", func(t *testing.T) {
		a := newApp(t)
		_, token := addUser(t, a, "ada@example.com", "member")
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/collections/documents/records",
			Body:            strings.NewReader(`{"title":"Wolfy","icon":"paw","body":"Two walks a day","shared":true}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusBadRequest, // shared is the server's to set
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
			ExpectedContent: []string{`"status":400`},
		}).Test(t)
	})

	t.Run("and the server marks it not shared", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/collections/documents/records",
			Body:            strings.NewReader(`{"title":"Wolfy","icon":"paw","body":"Two walks a day"}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"title":"Wolfy"`, `"shared":false`, `"created_by":"` + u.Id + `"`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("signed-out visitors see nothing", func(t *testing.T) {
		a := newApp(t)
		addRecord(t, a, "documents", map[string]any{"title": "Wolfy"})
		(&tests.ApiScenario{
			Method:             http.MethodGet,
			URL:                "/api/collections/documents/records",
			ExpectedStatus:     http.StatusOK,
			ExpectedContent:    []string{`"totalItems":0`},
			NotExpectedContent: []string{"Wolfy"},
			TestAppFactory:     func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})
}

const shareToken = "shareshareshareshareshareshare12"

// sharedApp has one person and a document about the dog with a photo; shared
// says whether it is shared with shareToken.
func sharedApp(t testing.TB, shared bool) (*tests.TestApp, *core.Record, string) {
	a := newApp(t)
	_, token := addUser(t, a, "ada@example.com", "member")
	doc := addRecord(t, a, "documents", map[string]any{
		"title": "Wolfy <3", "body": "Food: 2 cups\nVet: +32 470 12 34 56, https://vet.example/wolfy <script>",
	})
	photo, err := filesystem.NewFileFromBytes([]byte("\xff\xd8\xff\xe0 not really a jpeg"), "wolfy.jpg")
	if err != nil {
		t.Fatal(err)
	}
	doc.Set("files", []*filesystem.File{photo})
	if shared {
		doc.Set("share_token", shareToken)
		doc.Set("shared", true)
	}
	if err := a.Save(doc); err != nil {
		t.Fatal(err)
	}
	return a, doc, token
}

func TestSharing(t *testing.T) {
	t.Setenv("CALIO_FEED_URL", "https://calio-feed.example.ts.net")
	link := "/api/app/share/" + shareToken

	t.Run("making a link", func(t *testing.T) {
		a, doc, token := sharedApp(t, false)
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/app/documents/" + doc.Id + "/share",
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"url":"https://calio-feed.example.ts.net/api/app/share/`, `"public":true`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
			AfterTestFunc: func(t testing.TB, a *tests.TestApp, _ *http.Response) {
				doc, _ := a.FindRecordById("documents", doc.Id)
				if !doc.GetBool("shared") || len(doc.GetString("share_token")) != 32 {
					t.Error("the document isn't shared")
				}
			},
		}).Test(t)
	})

	t.Run("the page, escaped, without signing in", func(t *testing.T) {
		a, doc, _ := sharedApp(t, true)
		(&tests.ApiScenario{
			Method: http.MethodGet, URL: link, ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				"<h1>Wolfy &lt;3</h1>", "Food: 2 cups",
				`<a href="https://vet.example/wolfy" rel="noopener noreferrer">`, "&lt;script&gt;",
				`<a href="tel:+32470123456" rel="noopener noreferrer">+32 470 12 34 56</a>`,
				`<img src="` + link + "/" + doc.GetStringSlice("files")[0] + `"`,
			},
			NotExpectedContent: []string{"<script>"},
			TestAppFactory:     func(testing.TB) *tests.TestApp { return a },
			AfterTestFunc: func(t testing.TB, _ *tests.TestApp, res *http.Response) {
				if csp := res.Header.Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'none'") {
					t.Errorf("Content-Security-Policy = %q", csp)
				}
			},
		}).Test(t)
	})

	t.Run("its photo, and no other files", func(t *testing.T) {
		a, doc, _ := sharedApp(t, true)
		(&tests.ApiScenario{
			Method: http.MethodGet, URL: link + "/" + doc.GetStringSlice("files")[0], ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{"not really a jpeg"},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
		a, _, _ = sharedApp(t, true)
		(&tests.ApiScenario{
			Method: http.MethodGet, URL: link + "/other.jpg", ExpectedStatus: http.StatusNotFound,
			ExpectedContent: []string{`"status":404`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("stopping it", func(t *testing.T) {
		a, doc, token := sharedApp(t, true)
		(&tests.ApiScenario{
			Method:          http.MethodDelete,
			URL:             "/api/app/documents/" + doc.Id + "/share",
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"url":""`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
			AfterTestFunc: func(t testing.TB, a *tests.TestApp, _ *http.Response) {
				if _, err := a.FindFirstRecordByData("documents", "share_token", shareToken); err == nil {
					t.Error("the old link still works")
				}
			},
		}).Test(t)
	})

	t.Run("a link that isn't shared", func(t *testing.T) {
		a, _, _ := sharedApp(t, false)
		(&tests.ApiScenario{
			Method: http.MethodGet, URL: link, ExpectedStatus: http.StatusNotFound,
			ExpectedContent: []string{`"status":404`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("only the household makes links", func(t *testing.T) {
		a, doc, _ := sharedApp(t, false)
		(&tests.ApiScenario{
			Method: http.MethodPost, URL: "/api/app/documents/" + doc.Id + "/share",
			ExpectedStatus: http.StatusUnauthorized, ExpectedContent: []string{`"status":401`},
			TestAppFactory: func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})
}
