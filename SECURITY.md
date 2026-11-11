# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting on this repository: <https://github.com/Rethunk-Tech/go-modrinth/security/advisories/new>, or the **Security** tab, then **Report a vulnerability**. The report stays private between you and us until a fix ships.

If private reporting is unavailable, email [oss@rethunk.tech](mailto:oss@rethunk.tech) with **[SECURITY]** in the subject.

Please do not open a public issue, pull request or discussion for a suspected vulnerability.

A useful report has the affected repository and version (or commit), steps to reproduce, what an attacker gets and under which conditions, and whether anyone else knows.

## What counts as a vulnerability

A vulnerability is a flaw that lets someone cross a boundary they should not: run code, read or change data they have no right to, escape a sandbox or path restriction, or leak a credential. A bug is anything else that misbehaves, such as a crash on your own input, wrong output or a slow path; open an ordinary issue for those. If you are unsure which one you have, report it privately and we will say.

## Coordinated disclosure

We handle the report inside the private advisory, so you see each status change and can comment on the fix. The advisory goes public when the fix ships, or 90 days after your report, whichever comes first. If a fix is close, either of us can ask for more time in the thread. You are credited in the advisory unless you would rather not be.

## Response timelines

- We acknowledge your report within 3 business days.
- We give you an assessment (accepted, needs more information or declined, with reasons) within 10 business days.
- We aim to ship a fix within 90 days of the report.

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

## Safe harbour

We will not pursue legal action against good-faith research that avoids privacy violations, data destruction and disruption of services, stops at a proof of concept, and keeps the details private until the advisory is published. Rethunk.Tech, LLC maintains this repository.
