# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### :sparkles: Release Note <!-- optional -->

### :rocket: Enhancements
- [#911](https://github.com/reviewdog/reviewdog/pull/911) Added bitbucket server (v1 api) support
- Added support for Bitbucket Pipes executed within Pipelines.

---

## [v0.12.0] - 2021-06-26

### :rocket: Enhancements
- [#888](https://github.com/reviewdog/reviewdog/pull/888) Allow GitHub PR reporting for a forked repository iff it's triggered by `pull_request_target`
- [#976](https://github.com/reviewdog/reviewdog/pull/976) Treat `GITHUB_API_URL` environment variable as same as `GITHUB_API`, so users can use reviewdog in GitHub Actions in Enterprise Server without setting `GITHUB_API`

---

## [v0.11.0] - 2020-10-25

### :sparkles: Release Note
reviewdog v0.11 introduced [Reviewdog Diagnostic Format (RDFormat)](./README.md#reviewdog-diagnostic-format-rdformat)
as generic machine-readable diagnostic format and it unlocks new rich features like code suggestions.

### :rocket: Enhancements
- [#629](https://github.com/reviewdog/reviewdog/pull/629) Introduced Reviewdog Diagnostic Format.
 - [#674](https://github.com/reviewdog/reviewdog/pull/674) [#703](https://github.com/reviewdog/reviewdog/pull/703) Support rdjsonl/rdjson as input format
 - [#680](https://github.com/reviewdog/reviewdog/pull/680) github-pr-review: Support multiline comments
 - [#675](https://github.com/reviewdog/reviewdog/pull/675) [#698](https://github.com/reviewdog/reviewdog/pull/698) github-pr-review: Support suggested changes
 - [#699](https://github.com/reviewdog/reviewdog/pull/699) Support diff input format (`-f=diff`). Useful for suggested changes.
 - [#700](https://github.com/reviewdog/reviewdog/pull/700) Support to show code(rule), code URL and severity in GitHub and GitLab reporters.
- [#678](https://github.com/reviewdog/reviewdog/issues/678) github-pr-review: Support Code Suggestions
  - Introduced [reviewdog/action-suggester](https://github.com/reviewdog/action-suggester) action.
- Introduced [reviewdog/action-setup](https://github.com/reviewdog/action-setup) GitHub Action which installs reviewdog easily including nightly release.
- [#769](https://github.com/reviewdog/reviewdog/pull/769) Integration with [Bitbucket Code Insights](https://support.atlassian.com/bitbucket-cloud/docs/code-insights/) and [Bitbucket Pipelines](https://bitbucket.org/product/ru/features/pipelines)

---

## [v0.10.2] - 2020-08-04

### :bug: Fixes
- [#709](https://github.com/reviewdog/reviewdog/pull/709) Check for GITHUB_ACTIONS instead of GITHUB_ACTION

---

## [v0.10.1] - 2020-06-30

### :rocket: Enhancements
- [#563](https://github.com/reviewdog/reviewdog/issues/563) Use `CI_API_V4_URL` environment variable when present.

### :bug: Fixes
- [#609](https://github.com/reviewdog/reviewdog/issues/609) reviewdog command will fail with unexpected tool's error for github-check/github-pr-check reporters as well. ([@haya14busa])
- [#603](https://github.com/reviewdog/reviewdog/issues/603) Fixed detection of Pull Requests from forked repo. ([@haya14busa])

---

## [v0.10.0] - 2020-05-07

### :sparkles: Release Note

With v0.10.0 release, now reviewdog can find issues outside diff by controlling
filtering behavior with `-filter-mode`. Also, you can ensure to check reported
results by exit 1 with `-fail-on-error`.

Example
```shell
$ cd subdir/ && reviewdog -filter-mode=file -fail-on-error -reporter=github-pr-review
```

### :rocket: Enhancements
- [#446](https://github.com/reviewdog/reviewdog/issues/446)
  Added `-fail-on-error` flag
  ([document](https://github.com/reviewdog/reviewdog/tree/e359505275143ec85e9b114fc1ab4a4e91d04fb5#exit-codes))
  and improved exit code handling. ([@DmitryLanda](https://github.com/DmitryLanda), [@haya14busa])
- [#187](https://github.com/reviewdog/reviewdog/issues/187)
  Added `-filter-mode` flag [`added`, `diff_context`, `file`, `nofilter`]
  ([document](https://github.com/reviewdog/reviewdog/tree/e359505275143ec85e9b114fc1ab4a4e91d04fb5#filter-mode))
  which controls how reviewdog filter results. ([@Le6ow5k1](https://github.com/Le6ow5k1), [@haya14busa])
- [#69](https://github.com/reviewdog/reviewdog/issues/69) Support gerrit! ([@staticmukesh](https://github.com/staticmukesh))
- [#548](https://github.com/reviewdog/reviewdog/issues/548) Introduced nightly release ([reviewdog/nightly](https://github.com/reviewdog/nightly)). ([@haya14busa])

### :bug: Fixes
- [#461](https://github.com/reviewdog/reviewdog/issues/461) All reporters now supports sub-directory run. ([@haya14busa])

### :rotating_light: Breaking changes
- `github-check` reporter won't report results outside diff by default now. You
  need to use `-filter-mode=nofilter` to keep the same bahavior.

---

See https://github.com/reviewdog/reviewdog/releases for older release note.

[Unreleased]: https://github.com/reviewdog/reviewdog/compare/v0.12.0...HEAD
[v0.10.0]: https://github.com/reviewdog/reviewdog/compare/v0.9.17...v0.10.0
[v0.10.1]: https://github.com/reviewdog/reviewdog/compare/v0.10.0...v0.10.1
[v0.10.2]: https://github.com/reviewdog/reviewdog/compare/v0.10.1...v0.10.2
[v0.11.0]: https://github.com/reviewdog/reviewdog/compare/v0.10.2...v0.11.0
[v0.12.0]: https://github.com/reviewdog/reviewdog/compare/v0.11.0...v0.12.0
[@haya14busa]: https://github.com/haya14busa
