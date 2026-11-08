# HUMANS.md

Run and use `go-modrinth`. Internals: [AGENTS.md](AGENTS.md).

## Quick start

```bash
go get github.com/Rethunk-Tech/go-modrinth
```

```go
func main() {
 ctx := context.Background()
 c := modrinth.New(modrinth.Config{UserAgent: "my-launcher/1.0 (+https://example.com)"})
 found, err := c.Search(ctx, modrinth.SearchQuery{
  Text:   "sodium",
  Facets: [][]string{{"categories:fabric"}, {"versions:1.21.1"}},
  Index:  modrinth.IndexDownloads,
 })
 if err != nil {
  log.Fatal(err)
 }
 for _, hit := range found.Hits {
  fmt.Println(hit.Title, hit.Downloads, modrinth.PageURL("mod", hit.Slug))
 }
 versions, err := c.ProjectVersions(ctx, "AANobbMI", []string{"fabric"}, []string{"1.21.1"})
 if err != nil {
  log.Fatal(err)
 }
 if len(versions) == 0 {
  log.Fatal("no version matches that loader and game version")
 }
 file, ok := versions[0].PrimaryFile()
 if !ok {
  log.Fatal("the newest version has no file")
 }
 fmt.Println(file.URL, file.Filename, file.Size, file.Hashes["sha512"])
}
```

## Contracts

| Entry point | Contract |
| --- | --- |
| `New` | Never fails; `BaseURL` empty means the real API and a trailing slash is trimmed. |
| `Client` calls | A non-2xx answer is a `*httpkit.StatusError` (404 matches `httpkit.ErrNotFound`, 429 matches `httpkit.ErrRateLimited`); a transport error has its URL redacted. |
| Redirects | A redirect to another scheme or host is refused, so the request stays on the origin the caller picked. |
| Answers | Every body is read at most 16 MiB; a larger one is an error, never a partial decode. |
| `Search` | Sends no facet filter when `Facets` is empty; clamps `Limit` to `MaxPageSize`; never splits a page itself. |
| `Projects`, `Versions` | One request; name at most `ProjectsBatch` ids, since the client does not batch for you. |
| `ProjectVersions` | Sends no filter at all when `loaders` and `gameVersions` are both empty, rather than an empty array. |
| `LatestByHash` | A file already on its newest version is absent from the map; the client never downloads anything. |
| `Version.PrimaryFile` | The file flagged primary, else the first; `ok` is false only when the version has no files. |
| `ValidID` | Accepts `[A-Za-z0-9._-]{1,64}`; check a user-supplied id with it before `Project`. |
| `PageURL` | Escapes both segments; it names the site, not the API. |

## Requirements

Go 1.27.3 or newer, plus `lefthook` and `gitleaks` on PATH if you want the git hooks.

```bash
lefthook install
```

That is per clone; git does not commit hooks. Pre-commit scans staged changes for secrets, and
pre-push runs `gate`.

## Verify

```bash
gate
```
