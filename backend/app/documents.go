package app

import (
	"html/template"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

// Documents: notes and files about the dog, the house or the kids. A document
// can be shared with someone outside the household, like a sitter, through a
// private link that needs no sign-in:
//
//	GET /api/app/share/{token}         the document as a page
//	GET /api/app/share/{token}/{file}  one of its files
//
// The optional Tailscale Funnel publishes /api/app/share/ next to the feed, so
// the link works for a sitter who isn't on the tailnet.
const sharePrefix = "/api/app/share/"

func registerDocuments(app core.App) {
	app.OnRecordCreateRequest("documents").BindFunc(func(e *core.RecordRequestEvent) error {
		if e.Auth != nil && e.Auth.Collection().Name == "users" {
			e.Record.Set("created_by", e.Auth.Id)
		}
		e.Record.Set("shared", false)
		e.Record.Set("share_token", "")
		return e.Next()
	})
}

func registerDocumentRoutes(se *core.ServeEvent) {
	share := se.Router.Group("/api/app/documents/{id}/share").Bind(apis.RequireAuth("users"))
	share.GET("", func(e *core.RequestEvent) error { return shareLink(e, "get") })
	share.POST("", func(e *core.RequestEvent) error { return shareLink(e, "make") })
	share.DELETE("", func(e *core.RequestEvent) error { return shareLink(e, "stop") })

	se.Router.GET(sharePrefix+"{token}", serveShared)
	se.Router.GET(sharePrefix+"{token}/{file}", serveSharedFile)
}

// shareLink tells someone a document's private link ("url" is empty when it
// isn't shared), makes one, or stops it working.
func shareLink(e *core.RequestEvent, action string) error {
	doc, err := e.App.FindRecordById("documents", e.Request.PathValue("id"))
	if err != nil {
		return e.NotFoundError("That document is gone.", err)
	}
	switch action {
	case "make":
		if doc.GetString("share_token") == "" {
			doc.Set("share_token", security.RandomString(32))
			doc.Set("shared", true)
			if err := e.App.Save(doc); err != nil {
				return e.InternalServerError("", err)
			}
		}
	case "stop":
		doc.Set("share_token", "")
		doc.Set("shared", false)
		if err := e.App.Save(doc); err != nil {
			return e.InternalServerError("", err)
		}
	}
	base, public := publicBase(e)
	url := ""
	if token := doc.GetString("share_token"); token != "" {
		url = base + sharePrefix + token
	}
	return e.JSON(http.StatusOK, map[string]any{"url": url, "public": public})
}

func sharedDocument(e *core.RequestEvent) (*core.Record, bool) {
	token := e.Request.PathValue("token")
	if len(token) < 32 {
		return nil, false
	}
	doc, err := e.App.FindFirstRecordByData("documents", "share_token", token)
	return doc, err == nil
}

// The shared page lives on its own: no scripts, and nothing loaded from
// anywhere but its own files.
const sharedPolicy = "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; " +
	"base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

func serveShared(e *core.RequestEvent) error {
	doc, ok := sharedDocument(e)
	if !ok {
		return e.NotFoundError("", nil)
	}
	page := sharedPage{Title: doc.GetString("title"), Body: linkify(doc.GetString("body"))}
	for _, name := range doc.GetStringSlice("files") {
		page.Files = append(page.Files, sharedFile{
			URL:   sharePrefix + doc.GetString("share_token") + "/" + name,
			Name:  displayFileName(name),
			Image: isImage(name),
		})
	}
	var b strings.Builder
	if err := sharedTemplate.Execute(&b, page); err != nil {
		return e.InternalServerError("", err)
	}
	h := e.Response.Header()
	h.Set("Content-Security-Policy", sharedPolicy)
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("X-Robots-Tag", "noindex, nofollow")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Cache-Control", "no-store")
	return e.HTML(http.StatusOK, b.String())
}

func serveSharedFile(e *core.RequestEvent) error {
	doc, ok := sharedDocument(e)
	name := e.Request.PathValue("file")
	if !ok || !slices.Contains(doc.GetStringSlice("files"), name) {
		return e.NotFoundError("", nil)
	}
	fsys, err := e.App.NewFilesystem()
	if err != nil {
		return e.InternalServerError("", err)
	}
	defer fsys.Close()
	h := e.Response.Header()
	// as PocketBase serves files: nothing in a file runs
	h.Set("Content-Security-Policy", "default-src 'none'; media-src 'self'; style-src 'unsafe-inline'; sandbox")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Cache-Control", "no-store")
	if err := fsys.Serve(e.Response, e.Request, doc.BaseFilesPath()+"/"+name, name); err != nil {
		return e.NotFoundError("", err)
	}
	return nil
}

type sharedFile struct {
	URL, Name string
	Image     bool
}

type sharedPage struct {
	Title string
	Body  template.HTML
	Files []sharedFile
}

var imageExt = []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif"}

func isImage(name string) bool {
	return slices.Contains(imageExt, strings.ToLower(path.Ext(name)))
}

// PocketBase adds "_" and ten random characters to each file name.
var randomSuffix = regexp.MustCompile(`_[a-zA-Z0-9]{10}(\.[^.]*)?$`)

func displayFileName(name string) string {
	return randomSuffix.ReplaceAllString(name, "$1")
}

var (
	webLink   = regexp.MustCompile(`https?://[^\s<>"]+[^\s<>".,;:!?)]|\+?\d[\d ().-]{6,}\d`)
	notDialed = regexp.MustCompile(`[^\d+]`)
)

// linkify escapes text for a page and turns web addresses and phone numbers
// in it into links.
func linkify(s string) template.HTML {
	var b strings.Builder
	last := 0
	for _, m := range webLink.FindAllStringIndex(s, -1) {
		b.WriteString(template.HTMLEscapeString(s[last:m[0]]))
		text := s[m[0]:m[1]]
		href := text
		if !strings.HasPrefix(text, "http") {
			href = "tel:" + notDialed.ReplaceAllString(text, "")
		}
		b.WriteString(`<a href="` + template.HTMLEscapeString(href) + `" rel="noopener noreferrer">` +
			template.HTMLEscapeString(text) + `</a>`)
		last = m[1]
	}
	b.WriteString(template.HTMLEscapeString(s[last:]))
	return template.HTML(b.String())
}

var sharedTemplate = template.Must(template.New("shared").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex, nofollow">
<title>{{.Title}}</title>
<style>
  :root { color-scheme: light dark; --bg: #fafaf9; --card: #fff; --text: #18181b; --muted: #71717a; --line: #e4e4e7; --link: #0f766e; }
  @media (prefers-color-scheme: dark) { :root { --bg: #09090b; --card: #18181b; --text: #f4f4f5; --muted: #a1a1aa; --line: #27272a; --link: #5eead4; } }
  * { box-sizing: border-box; }
  body { margin: 0; background: var(--bg); color: var(--text); font: 16px/1.6 system-ui, -apple-system, "Segoe UI", sans-serif; }
  main { max-width: 42rem; margin: 0 auto; padding: 2rem 1rem 3rem; }
  h1 { font-size: 1.9rem; line-height: 1.2; margin: 0 0 1.25rem; letter-spacing: -0.02em; }
  .body { background: var(--card); border: 1px solid var(--line); border-radius: 1rem; padding: 1.1rem 1.25rem; white-space: pre-wrap; overflow-wrap: anywhere; }
  a { color: var(--link); }
  .images { display: grid; grid-template-columns: repeat(auto-fill, minmax(10rem, 1fr)); gap: .6rem; margin-top: 1rem; }
  .images img { width: 100%; height: 100%; aspect-ratio: 1; object-fit: cover; border-radius: .8rem; display: block; }
  ul { list-style: none; padding: 0; margin: 1rem 0 0; display: flex; flex-direction: column; gap: .4rem; }
  ul a { display: block; background: var(--card); border: 1px solid var(--line); border-radius: .8rem; padding: .6rem .9rem; text-decoration: none; }
  footer { margin-top: 2rem; color: var(--muted); font-size: .8rem; text-align: center; }
</style>
</head>
<body>
<main>
  <h1>{{.Title}}</h1>
  {{if .Body}}<div class="body">{{.Body}}</div>{{end}}
  {{if .Files}}
  <div class="images">{{range .Files}}{{if .Image}}<a href="{{.URL}}"><img src="{{.URL}}" alt="{{.Name}}" loading="lazy"></a>{{end}}{{end}}</div>
  <ul>{{range .Files}}{{if not .Image}}<li><a href="{{.URL}}">📄 {{.Name}}</a></li>{{end}}{{end}}</ul>
  {{end}}
  <footer>Shared from Calio</footer>
</main>
</body>
</html>
`))
