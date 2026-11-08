<h1 align="center">go-modrinth</h1>

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.27.3+-00ADD8?logo=go&logoColor=white)](go.mod)

</div>

---

A client for the Modrinth API: search, project and version metadata, and file downloads. Package `modrinth`, module `github.com/Rethunk-Tech/go-modrinth`.

The package is the protocol half of a Modrinth client and holds no policy: it sends the request the API documents, decodes the answer into exported types, and stops. It guarantees three things on every call — the body is read at most once and at most 16 MiB, a redirect cannot leave the API's origin, and a non-2xx answer arrives as a `*httpkit.StatusError` the caller can match. It exists so that paging, batching, retrying and deciding which file to install stay with the program that knows the game.

## Quick start

```bash
go get github.com/Rethunk-Tech/go-modrinth
```

A worked example and the contracts: [HUMANS.md](HUMANS.md).

## Features

- Search projects by text and facets, ordered by relevance, downloads, follows, newest or updated, paged by offset.
- Project metadata: title, summary, Markdown body, categories and moderation status.
- One request for many projects, and one for many versions, batched by the caller at `ProjectsBatch`.
- A project's versions filtered by loader and game version, and each version's primary file with its size, URL and SHA-1/SHA-512 digests.
- Hash lookups: the versions that published a set of files, and the newest version each file should be updated to.
- The category list, plus `ValidID` and `PageURL` for turning a user-supplied project id into a safe request or link.

## Documentation

| Doc | Audience |
| --- | --- |
| [HUMANS.md](HUMANS.md) | Running and using the package |
| [AGENTS.md](AGENTS.md) | Internals, invariants, and why |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Process — commits, tests, docs |
| [SECURITY.md](SECURITY.md) | Vulnerability disclosure |
| [CHANGELOG.md](CHANGELOG.md) | Release notes |

## License

MIT — see [LICENSE](LICENSE). © 2026 Rethunk.Tech, LLC.
