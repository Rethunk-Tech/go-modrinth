# Changelog

Notable changes to `go-modrinth`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-11-08

### Added

- `LatestByHash` takes a `versionTypes` filter, sent as `version_types`, so an update lookup can be kept to release
  channels; an empty filter sends no field.

## [0.1.0] - 2026-11-08

### Added

- `Client` over the Modrinth v2 API: `Search` with facets, index and offset paging; `Project`, `Projects` and
  `ProjectVersions`; `Versions` and `Categories`; and the hash lookups `VersionsByHash` and `LatestByHash`.
- Exported types `Project`, `Version` with `PrimaryFile`, `File`, `Dependency`, `SearchHit`, `SearchResult` and
  `Category`.
- `ValidID` and `PageURL` for a project id or slug, `MaxPageSize` and `ProjectsBatch` for paging and batching, and
  the `Index` sort constants.

[Unreleased]: https://github.com/Rethunk-Tech/go-modrinth/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/Rethunk-Tech/go-modrinth/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Rethunk-Tech/go-modrinth/releases/tag/v0.1.0
