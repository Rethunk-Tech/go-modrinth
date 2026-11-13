package modrinth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

type answer struct {
	code int
	body []byte
}

// FuzzHostileAnswer holds that no status or body makes a call panic. One server serves the whole run: a server per
// input runs the machine out of ports.
func FuzzHostileAnswer(f *testing.F) {
	f.Add(200, []byte(sodiumProject))
	f.Add(200, []byte(strings.Repeat(`{"a":[`, 20000)))
	f.Add(200, []byte(`{"h":{"files":[{"hashes":{"sha1":1}}]},"x":null}`))
	f.Add(500, []byte(`{"error":"x"}`))
	var a atomic.Pointer[answer]
	a.Store(&answer{code: 200})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cur := a.Load()
		w.WriteHeader(cur.code)
		_, _ = w.Write(cur.body)
	}))
	f.Cleanup(srv.Close)
	c := New(Config{BaseURL: srv.URL})
	f.Fuzz(func(t *testing.T, code int, body []byte) {
		if code < 200 || code > 599 || code/100 == 3 {
			code = 200
		}
		a.Store(&answer{code: code, body: body})
		ctx := t.Context()
		_, _ = c.Search(ctx, SearchQuery{})
		_, _ = c.Project(ctx, "sodium")
		_, _ = c.Projects(ctx, []string{"a"})
		_, _ = c.ProjectVersions(ctx, "sodium", nil, nil)
		_, _ = c.VersionsByHash(ctx, "sha1", []string{"h"})
		_, _ = c.LatestByHash(ctx, "sha1", []string{"h"}, nil, nil, nil)
		_, _ = c.Categories(ctx)
	})
}

func TestALyingContentLengthFails(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		conn, rw, err := http.NewResponseController(w).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = conn.Close() }()
		_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 1099511627776\r\n\r\n{}")
		_ = rw.Flush()
	}))
	t.Cleanup(srv.Close)
	if _, err := New(Config{BaseURL: srv.URL}).Project(t.Context(), "sodium"); err == nil {
		t.Fatal("a body cut short of its Content-Length was accepted")
	}
}
