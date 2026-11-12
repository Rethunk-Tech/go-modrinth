package modrinth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Rethunk-Tech/go-httpkit"
)

const (
	testAgent = "go-modrinth/test"

	sodiumProject = `{"id":"AANobbMI","slug":"sodium","title":"Sodium","description":"fast","body":"# Sodium",
"categories":["fabric"],"additional_categories":["Fabric","optimization"],"status":"approved"}`

	sodiumVersion = `{"id":"v2","project_id":"AANobbMI","version_number":"2.0","date_published":"2026-09-20T00:00:00Z",
"dependencies":[{"project_id":"P1","version_id":"V1","dependency_type":"required"},
{"project_id":"P2","dependency_type":"optional"},{"version_id":"V3","dependency_type":"required"}],
"files":[{"hashes":{"sha512":"ab"},"url":"https://cdn/x.jar","filename":"x.jar","primary":false,"size":1},
{"hashes":{"sha512":"cd","sha1":"ef"},"url":"https://cdn/y.jar","filename":"y.jar","primary":true,"size":7}]}`
)

// fake serves one Client whose every request is checked for the user agent.
func fake(t *testing.T, handle func(w http.ResponseWriter, r *http.Request)) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ua := r.UserAgent(); ua != testAgent {
			t.Errorf("user agent %q", ua)
		}
		handle(w, r)
	}))
	t.Cleanup(srv.Close)
	return New(Config{BaseURL: srv.URL + "/", HTTP: srv.Client(), UserAgent: testAgent})
}

func TestSearchSendsFacetsAndDecodesHits(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/search" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("query") != "sod" || q.Get("facets") != `[["categories:fabric"],["categories!=quilt"]]` ||
			q.Get("index") != IndexDownloads || q.Get("limit") != "20" || q.Get("offset") != "40" {
			t.Errorf("search params %v", q)
		}
		_, _ = w.Write([]byte(`{"total_hits":41,"hits":[{"project_id":"AANobbMI","slug":"sodium","title":"Sodium",
"description":"fast","author":"jelly","icon_url":"https://cdn/i.png","downloads":9,"follows":3,
"date_modified":"2026-09-20T00:00:00Z","versions":["v1","v2"],"latest_version":"v2"}]}`))
	})
	got, err := c.Search(t.Context(), SearchQuery{
		Text: "sod", Facets: [][]string{{"categories:fabric"}, {"categories!=quilt"}},
		Index: IndexDownloads, Offset: 40, Limit: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Total != 41 || len(got.Hits) != 1 {
		t.Fatalf("%+v", got)
	}
	hit := got.Hits[0]
	if hit.ProjectID != "AANobbMI" || hit.Slug != "sodium" || hit.Author != "jelly" || hit.IconURL != "https://cdn/i.png" ||
		hit.Downloads != 9 || hit.Follows != 3 || hit.Modified != "2026-09-20T00:00:00Z" || hit.Latest != "v2" ||
		len(hit.Versions) != 2 {
		t.Fatalf("%+v", hit)
	}
}

func TestSearchDefaultsToTheLargestPageAndRelevance(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("index") != IndexRelevance || q.Get("limit") != "100" || q.Get("offset") != "0" {
			t.Errorf("defaults %v", q)
		}
		if q.Get("facets") != "" {
			t.Errorf("facets sent with none: %v", q)
		}
		_, _ = w.Write([]byte(`{"total_hits":0,"hits":[]}`))
	})
	if _, err := c.Search(t.Context(), SearchQuery{Offset: -1, Limit: 500}); err != nil {
		t.Fatal(err)
	}
}

func TestProjectReadsOneByIDOrSlug(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project/sodium" {
			t.Errorf("path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(sodiumProject))
	})
	got, err := c.Project(t.Context(), "sodium")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "AANobbMI" || got.Slug != "sodium" || got.Body != "# Sodium" || got.Status != "approved" ||
		strings.Join(got.AdditionalCategories, ",") != "Fabric,optimization" {
		t.Fatalf("%+v", got)
	}
}

func TestProjectsNamesEveryIDInOneArray(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects" || r.URL.Query().Get("ids") != `["AANobbMI","P1"]` {
			t.Errorf("%s %v", r.URL.Path, r.URL.Query())
		}
		_, _ = w.Write([]byte(`[{"id":"AANobbMI","slug":"sodium","status":"archived"},{"id":"P1"}]`))
	})
	got, err := c.Projects(t.Context(), []string{"AANobbMI", "P1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Status != "archived" || got[1].ID != "P1" {
		t.Fatalf("%+v", got)
	}
}

func TestProjectsSendsEmptyArrayRatherThanNull(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("ids") != "[]" {
			t.Errorf("ids %q", r.URL.Query().Get("ids"))
		}
		_, _ = w.Write([]byte(`[]`))
	})
	if _, err := c.Projects(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestProjectVersionsSendsLoadersAndGameVersions(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/project/AANobbMI/version" || q.Get("loaders") != `["fabric"]` ||
			q.Get("game_versions") != `["1.21","1.20"]` {
			t.Errorf("%s %v", r.URL.Path, q)
		}
		_, _ = w.Write([]byte(`[` + sodiumVersion + `]`))
	})
	got, err := c.ProjectVersions(t.Context(), "AANobbMI", []string{"fabric"}, []string{"1.21", "1.20"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].VersionNumber != "2.0" || len(got[0].Dependencies) != 3 || got[0].Dependencies[0].Type != "required" {
		t.Fatalf("%+v", got)
	}
}

func TestProjectVersionsSendsNoFilterWhenNoneGiven(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query(); len(q) != 0 {
			t.Errorf("unfiltered params %v", q)
		}
		_, _ = w.Write([]byte(`[]`))
	})
	if _, err := c.ProjectVersions(t.Context(), "AANobbMI", nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestVersionsReadsByID(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/versions" || r.URL.Query().Get("ids") != `["v1","v2"]` {
			t.Errorf("%s %v", r.URL.Path, r.URL.Query())
		}
		_, _ = w.Write([]byte(`[` + sodiumVersion + `]`))
	})
	got, err := c.Versions(t.Context(), []string{"v1", "v2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "v2" {
		t.Fatalf("%+v", got)
	}
}

func TestVersionsSendsEmptyArrayRatherThanNull(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("ids") != "[]" {
			t.Errorf("ids %q", r.URL.Query().Get("ids"))
		}
		_, _ = w.Write([]byte(`[]`))
	})
	if _, err := c.Versions(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
}

// The version's own primary file wins, and the first file stands in when none is flagged.
func TestPrimaryFilePrefersTheFlaggedFileThenTheFirst(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[` + sodiumVersion + `]`))
	})
	got, err := c.Versions(t.Context(), []string{"v2"})
	if err != nil {
		t.Fatal(err)
	}
	f, ok := got[0].PrimaryFile()
	if !ok || f.Filename != "y.jar" || f.Size != 7 || f.Hashes["sha512"] != "cd" || f.Hashes["sha1"] != "ef" {
		t.Fatalf("%+v %v", f, ok)
	}
	none := Version{Files: []File{{Filename: "o.jar"}}}
	if f, ok := none.PrimaryFile(); !ok || f.Filename != "o.jar" {
		t.Fatalf("%+v %v", f, ok)
	}
	if _, ok := (Version{}).PrimaryFile(); ok {
		t.Fatal("a version with no file reported one")
	}
}

func TestVersionsByHashPostsTheHashesAndKeysTheAnswer(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/version_files" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content type %q", ct)
		}
		var body struct {
			Hashes    []string `json:"hashes"`
			Algorithm string   `json:"algorithm"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if strings.Join(body.Hashes, ",") != "ab,cd" || body.Algorithm != "sha512" {
			t.Errorf("body %+v", body)
		}
		_, _ = w.Write([]byte(`{"ab":` + sodiumVersion + `}`))
	})
	got, err := c.VersionsByHash(t.Context(), "sha512", []string{"ab", "cd"})
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := got["ab"]; !ok || v.ProjectID != "AANobbMI" {
		t.Fatalf("%+v", got)
	}
}

func TestLatestByHashSendsLoadersAndGameVersions(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/version_files/update" {
			t.Errorf("path %q", r.URL.Path)
		}
		var body struct {
			Hashes       []string `json:"hashes"`
			Algorithm    string   `json:"algorithm"`
			Loaders      []string `json:"loaders"`
			GameVersions []string `json:"game_versions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if strings.Join(body.Hashes, ",") != "ab" || body.Algorithm != "sha1" ||
			strings.Join(body.Loaders, ",") != "fabric,quilt" || strings.Join(body.GameVersions, ",") != "1.21" {
			t.Errorf("body %+v", body)
		}
		_, _ = w.Write([]byte(`{"ab":` + sodiumVersion + `}`))
	})
	got, err := c.LatestByHash(t.Context(), "sha1", []string{"ab"}, []string{"fabric", "quilt"}, []string{"1.21"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := got["ab"]; !ok || v.ID != "v2" {
		t.Fatalf("%+v", got)
	}
}

// The version types narrow an update to the channels a caller accepts, and an empty filter is left out rather than
// sent as an empty array the API could read as admitting no type.
func TestLatestByHashSendsVersionTypes(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		types []string
		want  string
	}{
		{[]string{"release", "beta"}, `,"version_types":["release","beta"]`},
		{nil, ``},
	} {
		c := fake(t, func(w http.ResponseWriter, r *http.Request) {
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			want := `{"hashes":["ab"],"algorithm":"sha512","loaders":["fabric"],"game_versions":["1.21"]` + tc.want + `}`
			if string(raw) != want {
				t.Errorf("body %s, want %s", raw, want)
			}
			_, _ = w.Write([]byte(`{}`))
		})
		if _, err := c.LatestByHash(t.Context(), "sha512", []string{"ab"}, []string{"fabric"}, []string{"1.21"}, tc.types); err != nil {
			t.Fatal(err)
		}
	}
}

// The API reads a missing loader or game version list as a filter, so an unfiltered call sends an empty array.
func TestLatestByHashSendsEmptyArraysRatherThanNull(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "null") {
			t.Errorf("body %s", raw)
		}
		_, _ = w.Write([]byte(`{}`))
	})
	if _, err := c.LatestByHash(t.Context(), "sha512", nil, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestCategoriesListsEveryTag(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tag/category" {
			t.Errorf("path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"name":"fabric","project_type":"mod"},{"name":"adventure","project_type":"modpack"}]`))
	})
	got, err := c.Categories(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "fabric" || got[1].ProjectType != "modpack" {
		t.Fatalf("%+v", got)
	}
}

// A non-2xx answer reaches the caller as the status, so a 404 is distinguishable from a transport failure.
func TestStatusBecomesAStatusError(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	_, err := c.Project(t.Context(), "gone")
	if !errors.Is(err, httpkit.ErrNotFound) {
		t.Fatalf("%v", err)
	}
	var se *httpkit.StatusError
	if !errors.As(err, &se) || se.Code != http.StatusNotFound {
		t.Fatalf("%v", err)
	}
}

// A redirect to another origin is refused, so a hostile answer cannot move a request to a host it names.
func TestCrossOriginRedirectIsRefused(t *testing.T) {
	t.Parallel()
	elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the refused redirect was followed")
	}))
	t.Cleanup(elsewhere.Close)
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		// A raw header, not http.Redirect: the redirect is the point of the test, and its target is a test server.
		w.Header().Set("Location", elsewhere.URL+r.URL.Path)
		w.WriteHeader(http.StatusFound)
	})
	if _, err := c.Categories(t.Context()); err == nil {
		t.Fatal("a cross-origin redirect was followed")
	}
}

// An answer over the byte cap is refused rather than read, so a hostile page cannot exhaust memory.
func TestOversizedAnswerIsRefused(t *testing.T) {
	t.Parallel()
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`["`))
		_, _ = w.Write([]byte(strings.Repeat("a", maxBody)))
		_, _ = w.Write([]byte(`"]`))
	})
	if _, err := c.Categories(t.Context()); err == nil {
		t.Fatal("an answer over the cap was read")
	}
}

func TestNewDefaultsTheBaseURLAndTrimsTheSlash(t *testing.T) {
	t.Parallel()
	if got := New(Config{}).base; got != BaseURL {
		t.Fatalf("base %q", got)
	}
	if got := New(Config{BaseURL: "https://example.test/v2///"}).base; got != "https://example.test/v2" {
		t.Fatalf("base %q", got)
	}
}

// ValidID is the only check between a caller's id and a request path or a site link.
func TestValidIDRefusesPathTricks(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"AANobbMI", "sodium", "my.mod_1-2", strings.Repeat("a", 64)} {
		if !ValidID(id) {
			t.Errorf("%q refused", id)
		}
	}
	for _, id := range []string{"", "../etc", "a/b", "a b", "a?b", strings.Repeat("a", 65)} {
		if ValidID(id) {
			t.Errorf("%q accepted", id)
		}
	}
}

func TestPageURLNamesTheTypeAndSlug(t *testing.T) {
	t.Parallel()
	if got := PageURL("mod", "sodium"); got != "https://modrinth.com/mod/sodium" {
		t.Fatalf("%q", got)
	}
	if got := PageURL("modpack", "a b"); got != "https://modrinth.com/modpack/a%20b" {
		t.Fatalf("%q", got)
	}
}
