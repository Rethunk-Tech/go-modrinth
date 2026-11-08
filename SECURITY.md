# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's **private vulnerability reporting** on this repository (Security → Report a
vulnerability). It stays private until a fix ships.

Please do not open a public issue for a suspected vulnerability first.

## What is in scope

The untrusted input is every JSON answer Modrinth returns, plus wherever its status line and headers
can send the client next.

- **Answer size.** `Search`, `Project`, `Projects`, `ProjectVersions`, `Versions`, `VersionsByHash`,
  `LatestByHash` and `Categories` read at most 16 MiB. An answer past that is an error, not a
  truncated decode, so a hostile or broken server cannot make the client allocate without bound.
- **Redirects.** A redirect to another scheme or host is refused, so no answer can move a request to
  a host it names. The client sends no credential, so a same-origin redirect grants nothing.
- **Malformed bodies.** A body that is not valid JSON for the expected shape is an error from the
  call. Unknown fields are ignored, and no field of an answer is used to build a path the client then
  requests.
- **Error text.** A transport error has its query string redacted before it reaches the caller or a
  log, since a query can carry a signature.

## What is out of scope

- **Id validation is the caller's.** `ValidID` and `PageURL` exist because an id taken from a user or
  a file must be checked before it enters a request path or a link; the client escapes a path segment
  but does not decide which ids are acceptable.
- **The files themselves.** The client reads metadata only: it downloads nothing, and it does not
  verify a file's `sha1` or `sha512` digest. Checking a downloaded file against
  `File.Hashes["sha512"]` is the caller's job.
- **What Modrinth says.** A project's `Body`, a file's `URL` and a dependency's `ProjectID` are
  published content, not instructions. Trusting them is a caller policy.
- **`Config.UserAgent`.** It is sent as given; a caller that puts a secret in it has put that secret
  in every request.
