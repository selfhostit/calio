package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestContacts(t *testing.T) {
	t.Run("a member adds the vet", func(t *testing.T) {
		a := newApp(t)
		u, token := addUser(t, a, "ada@example.com", "member")
		(&tests.ApiScenario{
			Method:          http.MethodPost,
			URL:             "/api/collections/contacts/records",
			Body:            strings.NewReader(`{"name":"Dr. Papadopoulou","role":"vet","phone":"+30 210 123 4567"}`),
			Headers:         map[string]string{"Authorization": token},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"role":"vet"`, `"created_by":"` + u.Id + `"`},
			TestAppFactory:  func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})

	t.Run("signed-out visitors see nothing", func(t *testing.T) {
		a := newApp(t)
		contacts, _ := a.FindCollectionByNameOrId("contacts")
		vet := core.NewRecord(contacts)
		vet.Load(map[string]any{"name": "Dr. Papadopoulou", "role": "vet"})
		if err := a.Save(vet); err != nil {
			t.Fatal(err)
		}
		(&tests.ApiScenario{
			Method:             http.MethodGet,
			URL:                "/api/collections/contacts/records",
			ExpectedStatus:     http.StatusOK,
			ExpectedContent:    []string{`"totalItems":0`},
			NotExpectedContent: []string{"Papadopoulou"},
			TestAppFactory:     func(testing.TB) *tests.TestApp { return a },
		}).Test(t)
	})
}
