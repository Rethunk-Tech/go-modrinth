package modrinth

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func FuzzValidID(f *testing.F) {
	for _, id := range []string{
		"AANobbMI", "sodium", "my.mod_1-2", strings.Repeat("a", 64),
		"", ".", "..", "../etc", "a/b", "a b", "a?b", strings.Repeat("a", 65),
	} {
		f.Add(id)
	}
	f.Fuzz(func(t *testing.T, id string) {
		want := id != "." && id != ".." && len(id) >= 1 && len(id) <= 64
		for i := 0; i < len(id) && want; i++ {
			c := id[i]
			want = c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-'
		}
		if ValidID(id) != want {
			t.Fatalf("ValidID(%q) = %v, want %v", id, ValidID(id), want)
		}
	})
}

func FuzzPageURL(f *testing.F) {
	f.Add("mod", "sodium")
	f.Add("modpack", "a b")
	f.Fuzz(func(t *testing.T, projectType, slug string) {
		got := PageURL(projectType, slug)
		u, err := url.Parse(got)
		if err != nil {
			t.Fatalf("PageURL(%q, %q) = %q: %v", projectType, slug, got, err)
		}
		if u.Scheme != "https" || u.Host != "modrinth.com" {
			t.Fatalf("PageURL(%q, %q) left https://modrinth.com: %q", projectType, slug, got)
		}
		parts := strings.SplitN(strings.TrimPrefix(u.EscapedPath(), "/"), "/", 2)
		if len(parts) != 2 {
			t.Fatalf("PageURL(%q, %q) path %q", projectType, slug, u.EscapedPath())
		}
		gotType, err1 := url.PathUnescape(parts[0])
		gotSlug, err2 := url.PathUnescape(parts[1])
		if err1 != nil || err2 != nil || gotType != projectType || gotSlug != slug {
			t.Fatalf("PageURL(%q, %q) = %q", projectType, slug, got)
		}
	})
}

func FuzzProject(f *testing.F) {
	f.Add([]byte(sodiumProject))
	f.Add([]byte(`{"id":"AANobbMI","slug":"sodium","status":"archived"}`))
	f.Add([]byte(`{}`))
	fuzzJSON[Project](f)
}

func FuzzVersion(f *testing.F) {
	f.Add([]byte(sodiumVersion))
	f.Add([]byte(`{"files":[],"dependencies":[]}`))
	f.Add([]byte(`{}`))
	fuzzJSON[Version](f)
}

func FuzzSearchResult(f *testing.F) {
	f.Add([]byte(`{"total_hits":41,"hits":[{"project_id":"AANobbMI","slug":"sodium","title":"Sodium",
"description":"fast","author":"jelly","icon_url":"https://cdn/i.png","downloads":9,"follows":3,
"date_modified":"2026-09-20T00:00:00Z","versions":["v1","v2"],"latest_version":"v2"}]}`))
	f.Add([]byte(`{"total_hits":0,"hits":[]}`))
	f.Add([]byte(`{}`))
	fuzzJSON[SearchResult](f)
}

func FuzzCategory(f *testing.F) {
	f.Add([]byte(`{"name":"fabric","project_type":"mod"}`))
	f.Add([]byte(`{"name":"adventure","project_type":"modpack"}`))
	f.Add([]byte(`{}`))
	fuzzJSON[Category](f)
}

func FuzzProjects(f *testing.F) {
	f.Add([]byte(`[` + sodiumProject + `]`))
	f.Add([]byte(`[{"id":"AANobbMI","slug":"sodium","status":"archived"},{"id":"P1"}]`))
	f.Add([]byte(`[]`))
	fuzzJSON[[]Project](f)
}

func FuzzVersions(f *testing.F) {
	f.Add([]byte(`[` + sodiumVersion + `]`))
	f.Add([]byte(`[]`))
	fuzzJSON[[]Version](f)
}

func FuzzVersionsByHash(f *testing.F) {
	f.Add([]byte(`{"ab":` + sodiumVersion + `}`))
	f.Add([]byte(`{}`))
	fuzzJSON[map[string]Version](f)
}

func fuzzJSON[T any](f *testing.F) {
	f.Fuzz(func(t *testing.T, raw []byte) {
		var v T
		if err := json.Unmarshal(raw, &v); err != nil {
			return
		}
		enc, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		var v2 T
		if err := json.Unmarshal(enc, &v2); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(v, v2) {
			t.Fatalf("%+v vs %+v", v, v2)
		}
	})
}
