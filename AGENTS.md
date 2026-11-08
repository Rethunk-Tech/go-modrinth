# AGENTS.md

`modrinth` is a client for the public Modrinth v2 API: search, projects, versions, files and categories. It holds the
protocol half of a source driver, so a program that searches or installs Modrinth projects imports it; it imports
go-httpkit and the standard library only, never a consumer package.

## Gate

`gate` is the command. Each override in `.gate.toml` is there because a
runner cannot resolve that check from PATH.

## Shape

- `Client.do` is the single request path for every method: it caps the answer at `maxBody` (16 MiB), refuses a
  redirect off the API's origin through `httpkit.SameOrigin`, redacts a transport error's URL, and turns a non-2xx
  answer into a `*httpkit.StatusError` so a caller can tell a 404 from a network failure. A new endpoint goes through
  it rather than building its own request.
- `maxBody` is a policy bound, not the API's: it is far above the largest real answer and far below a size that
  would threaten memory. `MaxPageSize` and `ProjectsBatch` are the API's own bounds, so they are exported for the
  caller to page and to batch with; the client clamps `SearchQuery.Limit` and nothing else, and never splits a
  batch itself.
- Every exported type field exists because a caller reads it today. A field is exported only once something reads it,
  because an unread field is a promise to keep it working.
- `PageURL` escapes both segments and `ValidID` refuses anything outside `[A-Za-z0-9._-]{1,64}`, so an id a user typed
  cannot escape a request path or a site link.
- Nothing here is stateful: no cache, no token, no retry. A caller that must batch, retry or cache does so on top of
  a `*Client`, which is why `New` can be called again at any time.

## Files

| File | Holds |
| --- | --- |
| `client.go` | Config, Client, New, the request path, Search, Project, Projects, ProjectVersions, Versions, VersionsByHash, LatestByHash, Categories |
| `types.go` | Project, Version, File, Dependency, SearchHit, SearchResult, Category, ValidID, PageURL |
| `doc.go` | The package clause |

## Tests

`client_test.go` runs one `httptest.Server` per method and checks the request it receives — method, path, query or
body, User-Agent — as well as what the method decodes. Behaviour that exists for a hostile answer (the byte cap, the
cross-origin redirect refusal) is pinned there too.
