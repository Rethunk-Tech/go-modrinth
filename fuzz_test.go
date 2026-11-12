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
		"", "../etc", "a/b", "a b", "a?b", strings.Repeat("a", 65),
	} {
		f.Add(id)
	}
	f.Fuzz(func(t *testing.T, id string) {
		want := len(id) >= 1 && len(id) <= 64
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
	f.Fuzz(func(t *testing.T, raw []byte) {
		var p Project
		if err := json.Unmarshal(raw, &p); err != nil {
			return
		}
		enc, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		var p2 Project
		if err := json.Unmarshal(enc, &p2); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(p, p2) {
			t.Fatalf("%+v vs %+v", p, p2)
		}
	})
}
