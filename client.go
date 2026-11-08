package modrinth

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Rethunk-Tech/go-httpkit"
)

const (
	// BaseURL is the Modrinth API root.
	BaseURL = "https://api.modrinth.com/v2"
	// MaxPageSize is the search API's largest limit.
	MaxPageSize = 100
	// ProjectsBatch is how many ids one Projects, Versions or VersionsByHash call should name.
	ProjectsBatch = 100

	// maxBody caps one API answer at 16 MiB, so a hostile or broken answer cannot be read without bound.
	maxBody = 16 << 20
)

// Index orders a search result; SearchQuery.Index takes one.
const (
	// IndexRelevance ranks by Modrinth's own relevance score, the API's default.
	IndexRelevance = "relevance"
	// IndexDownloads ranks by total downloads.
	IndexDownloads = "downloads"
	// IndexFollows ranks by follower count.
	IndexFollows = "follows"
	// IndexNewest ranks by first publication.
	IndexNewest = "newest"
	// IndexUpdated ranks by last modification.
	IndexUpdated = "updated"
)

// Config is what a Client is built from.
type Config struct {
	// BaseURL replaces the API root; empty means BaseURL.
	BaseURL string
	// HTTP sends every request; nil means a client with httpkit's default timeout. A redirect to another origin is
	// refused.
	HTTP *http.Client
	// UserAgent is sent with every request when set. Modrinth asks every client to identify itself.
	UserAgent string
}

// Client calls the Modrinth API. It holds no state beyond its configuration, so it is safe for concurrent use.
type Client struct {
	base, agent string
	hc          *http.Client
}

// New builds a Client.
func New(cfg Config) *Client {
	return &Client{
		base:  strings.TrimRight(cmp.Or(cfg.BaseURL, BaseURL), "/"),
		agent: cfg.UserAgent,
		hc:    httpkit.WithRedirect(cfg.HTTP, httpkit.SameOrigin),
	}
}

// SearchQuery is one search. Zero fields keep the API's defaults: no text, no facet filter, relevance order and the
// first page.
type SearchQuery struct {
	// Text matches project names, summaries and descriptions.
	Text string
	// Facets keeps only projects matching every group, and the strings inside one group are alternatives, as in
	// [][]string{{"categories:fabric", "categories:quilt"}}.
	Facets [][]string
	// Index is one of the Index constants; empty means IndexRelevance.
	Index string
	// Offset is the first hit's position and Limit the page length, at most MaxPageSize; zero Limit is MaxPageSize.
	Offset, Limit int
}

// Search lists projects matching q.
func (c *Client) Search(ctx context.Context, q SearchQuery) (SearchResult, error) {
	params := url.Values{}
	params.Set("query", q.Text)
	if len(q.Facets) > 0 {
		raw, err := json.Marshal(q.Facets)
		if err != nil {
			return SearchResult{}, fmt.Errorf("modrinth: %w", err)
		}
		params.Set("facets", string(raw))
	}
	params.Set("index", cmp.Or(q.Index, IndexRelevance))
	params.Set("limit", strconv.Itoa(min(cmp.Or(q.Limit, MaxPageSize), MaxPageSize)))
	params.Set("offset", strconv.Itoa(max(q.Offset, 0)))
	var out SearchResult
	err := c.get(ctx, "/search", params, &out)
	return out, err
}

// Project reads one project by id or slug.
func (c *Client) Project(ctx context.Context, idOrSlug string) (Project, error) {
	var out Project
	err := c.get(ctx, "/project/"+url.PathEscape(idOrSlug), nil, &out)
	return out, err
}

// Projects reads many projects in one request; name at most ProjectsBatch ids.
func (c *Client) Projects(ctx context.Context, ids []string) ([]Project, error) {
	raw, err := json.Marshal(ids)
	if err != nil {
		return nil, fmt.Errorf("modrinth: %w", err)
	}
	var out []Project
	err = c.get(ctx, "/projects", url.Values{"ids": {string(raw)}}, &out)
	return out, err
}

// ProjectVersions lists a project's versions, newest first, kept to loaders and game versions when either is named.
func (c *Client) ProjectVersions(ctx context.Context, project string, loaders, gameVersions []string) ([]Version, error) {
	var params url.Values
	for _, f := range []struct {
		name  string
		names []string
	}{{"loaders", loaders}, {"game_versions", gameVersions}} {
		raw, err := json.Marshal(f.names)
		if err != nil {
			return nil, fmt.Errorf("modrinth: %w", err)
		}
		if len(f.names) > 0 {
			if params == nil {
				params = url.Values{}
			}
			params.Set(f.name, string(raw))
		}
	}
	var out []Version
	err := c.get(ctx, "/project/"+url.PathEscape(project)+"/version", params, &out)
	return out, err
}

// Versions reads many versions by id in one request.
func (c *Client) Versions(ctx context.Context, ids []string) ([]Version, error) {
	raw, err := json.Marshal(ids)
	if err != nil {
		return nil, fmt.Errorf("modrinth: %w", err)
	}
	var out []Version
	err = c.get(ctx, "/versions", url.Values{"ids": {string(raw)}}, &out)
	return out, err
}

// VersionsByHash reads the versions that published the given files, keyed by the hash given.
func (c *Client) VersionsByHash(ctx context.Context, algorithm string, hashes []string) (map[string]Version, error) {
	var out map[string]Version
	err := c.post(ctx, "/version_files", struct {
		Hashes    []string `json:"hashes"`
		Algorithm string   `json:"algorithm"`
	}{Hashes: orEmpty(hashes), Algorithm: algorithm}, &out)
	return out, err
}

// LatestByHash finds the newest version of each file that matches the loaders and game versions, keyed by the hash
// given. A file with no newer version is absent from the map.
func (c *Client) LatestByHash(ctx context.Context, algorithm string, hashes, loaders, gameVersions []string) (map[string]Version, error) {
	var out map[string]Version
	err := c.post(ctx, "/version_files/update", struct {
		Hashes       []string `json:"hashes"`
		Algorithm    string   `json:"algorithm"`
		Loaders      []string `json:"loaders"`
		GameVersions []string `json:"game_versions"`
	}{Hashes: orEmpty(hashes), Algorithm: algorithm, Loaders: orEmpty(loaders), GameVersions: orEmpty(gameVersions)}, &out)
	return out, err
}

// Categories lists the site's project categories across project types.
func (c *Client) Categories(ctx context.Context) ([]Category, error) {
	var out []Category
	err := c.get(ctx, "/tag/category", nil, &out)
	return out, err
}

func (c *Client) get(ctx context.Context, path string, params url.Values, out any) error {
	return c.do(ctx, http.MethodGet, path, params, nil, out)
}

func (c *Client) post(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, nil, body, out)
}

// do sends one request and decodes its 200 answer into out. A redirect off the API's origin is refused, a transport
// error has its URL redacted, a non-2xx answer is a *httpkit.StatusError, and an answer over maxBody is an error.
func (c *Client) do(ctx context.Context, method, path string, params url.Values, body, out any) error {
	u := c.base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("modrinth: %w", err)
		}
		payload = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, payload)
	if err != nil {
		return fmt.Errorf("modrinth: %w", err)
	}
	if c.agent != "" {
		req.Header.Set("User-Agent", c.agent)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	raw, err := httpkit.DoCapped(c.hc, req, maxBody, nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("modrinth: %w", err)
	}
	return nil
}

// orEmpty is a filter the API requires as an array, which a nil slice would send as null.
func orEmpty(names []string) []string {
	if len(names) == 0 {
		return []string{}
	}
	return names
}
