<div align="center">
  <a href="https://github.com/reviewdog/reviewdog">
    <img alt="reviewdog" src="https://raw.githubusercontent.com/haya14busa/i/d598ed7dc49fefb0018e422e4c43e5ab8f207a6b/reviewdog/reviewdog.logo.png">
  </a>
</div>

<h2 align="center">
  reviewdog - A code review dog who keeps your codebase healthy.
</h2>

<div align="center">
  <a href="./LICENSE">
    <img alt="LICENSE" src="https://img.shields.io/badge/license-MIT-blue.svg?maxAge=43200">
  </a>
  <a href="https://godoc.org/github.com/reviewdog/reviewdog">
    <img alt="GoDoc" src="https://img.shields.io/badge/godoc-reference-4F73B3.svg?label=godoc.org&maxAge=43200&logo=go">
  </a>
  <a href="./CHANGELOG.md">
    <img alt="releases" src="https://img.shields.io/github/release/reviewdog/reviewdog.svg?logo=github">
  </a>
  <a href="https://github.com/reviewdog/nightly">
    <img alt="nightly releases" src="https://img.shields.io/github/v/release/reviewdog/nightly.svg?logo=github">
  </a>
</div>

<div align="center">
  <a href="https://github.com/reviewdog/reviewdog/actions?query=workflow%3AGo+event%3Apush+branch%3Amaster">
    <img alt="GitHub Actions" src="https://github.com/reviewdog/reviewdog/workflows/Go/badge.svg">
  </a>
  <a href="https://github.com/reviewdog/reviewdog/actions?query=workflow%3Areviewdog+event%3Apush+branch%3Amaster">
    <img alt="reviewdog" src="https://github.com/reviewdog/reviewdog/workflows/reviewdog/badge.svg?branch=master&event=push">
  </a>
  <a href="https://github.com/reviewdog/reviewdog/actions?query=workflow%3Arelease">
    <img alt="release" src="https://github.com/reviewdog/reviewdog/workflows/release/badge.svg">
  </a>
  <a href="https://travis-ci.org/reviewdog/reviewdog"><img alt="Travis Status" src="https://img.shields.io/travis/reviewdog/reviewdog/master.svg?label=Travis&logo=travis"></a>
  <a href="https://circleci.com/gh/reviewdog/reviewdog"><img alt="CircleCI Status" src="http://img.shields.io/circleci/build/github/reviewdog/reviewdog/master.svg?label=CircleCI&logo=circleci"></a>
  <a href="https://codecov.io/github/reviewdog/reviewdog"><img alt="Coverage Status" src="https://img.shields.io/codecov/c/github/reviewdog/reviewdog/master.svg?logo=codecov"></a>
</div>

<div align="center">
  <a href="https://gitlab.com/reviewdog/reviewdog/pipelines">
    <img alt="GitLab Supported" src="https://img.shields.io/badge/GitLab%20-Supported-fc6d26?logo=gitlab">
  </a>
  <a href="https://github.com/haya14busa/action-bumpr">
    <img alt="action-bumpr supported" src="https://img.shields.io/badge/bumpr-supported-ff69b4?logo=github&link=https://github.com/haya14busa/action-bumpr">
  </a>
  <a href="https://github.com/reviewdog/.github/blob/master/CODE_OF_CONDUCT.md">
    <img alt="Contributor Covenant" src="https://img.shields.io/badge/Contributor%20Covenant-v2.0%20adopted-ff69b4.svg">
  </a>
  <a href="https://somsubhra.github.io/github-release-stats/?username=reviewdog&repository=reviewdog&per_page=30">
    <img alt="Github Releases Stats" src="https://img.shields.io/github/downloads/reviewdog/reviewdog/total.svg?logo=github">
  </a>
  <a href="https://starcharts.herokuapp.com/reviewdog/reviewdog"><img alt="Stars" src="https://img.shields.io/github/stars/reviewdog/reviewdog.svg?style=social"></a>
</div>
<br />

reviewdog provides a way to post review comments to code hosting service,
such as GitHub, automatically by integrating with any linter tools with ease.
It uses an output of lint tools and posts them as a comment if findings are in
diff of patches to review.

reviewdog also supports run in the local environment to filter an output of lint tools
by diff.

[design doc](https://docs.google.com/document/d/1mGOX19SSqRowWGbXieBfGPtLnM0BdTkIc9JelTiu6wA/edit?usp=sharing)

## Table of Contents

- [Installation](#installation)
- [Input Format](#input-format)
  * ['errorformat'](#errorformat)
  * [Available pre-defined 'errorformat'](#available-pre-defined-errorformat)
  * [Reviewdog Diagnostic Format (RDFormat)](#reviewdog-diagnostic-format-rdformat)
  * [Diff](#diff)
  * [checkstyle format](#checkstyle-format)
- [Code Suggestions](#code-suggestions)
- [reviewdog config file](#reviewdog-config-file)
- [Reporters](#reporters)
  * [Reporter: Local (-reporter=local) [default]](#reporter-local--reporterlocal-default)
  * [Reporter: GitHub Checks (-reporter=github-pr-check)](#reporter-github-checks--reportergithub-pr-check)
  * [Reporter: GitHub Checks (-reporter=github-check)](#reporter-github-checks--reportergithub-check)
  * [Reporter: GitHub PullRequest review comment (-reporter=github-pr-review)](#reporter-github-pullrequest-review-comment--reportergithub-pr-review)
  * [Reporter: GitLab MergeRequest discussions (-reporter=gitlab-mr-discussion)](#reporter-gitlab-mergerequest-discussions--reportergitlab-mr-discussion)
  * [Reporter: GitLab MergeRequest commit (-reporter=gitlab-mr-commit)](#reporter-gitlab-mergerequest-commit--reportergitlab-mr-commit)
  * [Reporter: Bitbucket Code Insights Reports (-reporter=bitbucket-code-report)](#reporter-bitbucket-code-insights-reports--reporterbitbucket-code-report)
  * [Reporter: Bitbucket Server PullRequest review comment  (-reporter=bitbucket-pr-review)](#reporter-bitbucket-server-pullrequest-review-comment--reporterbitbucket-pr-review)
- [Supported CI services](#supported-ci-services)
  * [GitHub Actions](#github-actions)
  * [Travis CI](#travis-ci)
  * [Circle CI](#circle-ci)
  * [GitLab CI](#gitlab-ci)
  * [Bitbucket Pipelines](#bitbucket-pipelines)
  * [Common (Jenkins, local, etc...)](#common-jenkins-local-etc)
    + [Jenkins with Github pull request builder plugin](#jenkins-with-github-pull-request-builder-plugin)
- [Exit codes](#exit-codes)
- [Filter mode](#filter-mode)
- [Articles](#articles)

[![github-pr-check sample](https://user-images.githubusercontent.com/3797062/40884858-6efd82a0-6756-11e8-9f1a-c6af4f920fb0.png)](https://github.com/reviewdog/reviewdog/pull/131/checks)
![comment in pull-request](https://user-images.githubusercontent.com/3797062/40941822-1d775064-6887-11e8-98e9-4775d37d47f8.png)
![commit status](https://user-images.githubusercontent.com/3797062/40941738-d62acb0a-6886-11e8-858d-7b97aded2a42.png)
[![sample-comment.png](https://raw.githubusercontent.com/haya14busa/i/dc0ccb1e110515ea407c146d99b749018db05c45/reviewdog/sample-comment.png)](https://github.com/reviewdog/reviewdog/pull/24#discussion_r84599728)
![reviewdog-local-demo.gif](https://raw.githubusercontent.com/haya14busa/i/dc0ccb1e110515ea407c146d99b749018db05c45/reviewdog/reviewdog-local-demo.gif)

## Installation

```shell
# Install the latest version. (Install it into ./bin/ by default).
$ curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
$ curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
$ wget -O - -q https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh | sh -s [vX.Y.Z]
```

### Nightly releases

You can also use [nightly reviewdog release](https://github.com/reviewdog/nightly)
to try the latest reviewdog improvements every day!

```shell
$ curl -sfL https://raw.githubusercontent.com/reviewdog/nightly/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### GitHub Action: [reviewdog/action-setup](https://github.com/reviewdog/action-setup)

```yaml
steps:
- uses: reviewdog/action-setup@v1
  with:
    reviewdog_version: latest # Optional. [latest,nightly,v.X.Y.Z]
```

### homebrew / linuxbrew
You can also install reviewdog using brew:

```shell
$ brew install reviewdog/tap/reviewdog
$ brew upgrade reviewdog/tap/reviewdog
```

### Build with go install

```shell
$ go install github.com/reviewdog/reviewdog/cmd/reviewdog@latest
```

## Input Format

### 'errorformat'

reviewdog accepts any compiler or linter result from stdin and parses it with
scan-f like [**'errorformat'**](https://github.com/reviewdog/errorformat),
which is the port of Vim's [errorformat](https://vim-jp.org/vimdoc-en/quickfix.html#error-file-format)
feature.

For example, if the result format is `{file}:{line number}:{column number}: {message}`,
errorformat should be `%f:%l:%c: %m` and you can pass it as `-efm` arguments.

```shell
$ golint ./...
comment_iowriter.go:11:6: exported type CommentWriter should have comment or be unexported
$ golint ./... | reviewdog -efm="%f:%l:%c: %m" -diff="git diff FETCH_HEAD"
```

| name | description |
| ---- | ----------- |
| %f | file name |
| %l | line number |
| %c | column number |
| %m | error message |
| %% | the single '%' character |
| ... | ... |

Please see [reviewdog/errorformat](https://github.com/reviewdog/errorformat)
and [:h errorformat](https://vim-jp.org/vimdoc-en/quickfix.html#error-file-format)
if you want to deal with a more complex output. 'errorformat' can handle more
complex output like a multi-line error message.

You can also try errorformat on [the Playground](https://reviewdog.github.io/errorformat-playground/)!

By this 'errorformat' feature, reviewdog can support any tools output with ease.

### Available pre-defined 'errorformat'

But, you don't have to write 'errorformat' in many cases. reviewdog supports
pre-defined errorformat for major tools.

You can find available errorformat name by `reviewdog -list` and you can use it
with `-f={name}`.

```shell
$ reviewdog -list
golint          linter for Go source code                                       - https://github.com/golang/lint
govet           Vet examines Go source code and reports suspicious problems     - https://golang.org/cmd/vet/
sbt             the interactive build tool                                      - http://www.scala-sbt.org/
...
```

```shell
$ golint ./... | reviewdog -f=golint -diff="git diff FETCH_HEAD"
```

You can add supported pre-defined 'errorformat' by contributing to [reviewdog/errorformat](https://github.com/reviewdog/errorformat)

### Reviewdog Diagnostic Format (RDFormat)

reviewdog supports [Reviewdog Diagnostic Format (RDFormat)](./proto/rdf/) as a
generic diagnostic format and it supports both [rdjson](./proto/rdf/#rdjson) and
[rdjsonl](./proto/rdf/#rdjsonl) formats.

This rdformat supports rich feature like multiline ranged comments, severity,
rule code with URL, and [code suggestions](#code-suggestions).

```shell
$ <linter> | <convert-to-rdjson> | reviewdog -f=rdjson -reporter=github-pr-review
# or
$ <linter> | <convert-to-rdjsonl> | reviewdog -f=rdjsonl -reporter=github-pr-review
```

#### Example: ESLint with RDFormat 

![eslint reviewdog rdjson demo](https://user-images.githubusercontent.com/3797062/97085944-87233a80-165b-11eb-94a8-0a47d5e24905.png)

You can use [eslint-formatter-rdjson](https://www.npmjs.com/package/eslint-formatter-rdjson)
to output `rdjson` as eslint output format.

```shell
$ npm install --save-dev eslint-formatter-rdjson
$ eslint -f rdjson . | reviewdog -f=rdjson -reporter=github-pr-review
```

Or you can also use [reviewdog/action-eslint](https://github.com/reviewdog/action-eslint) for GitHub Actions.

### Diff

![reviewdog with gofmt example](https://user-images.githubusercontent.com/3797062/89168305-a3ad5a80-d5b7-11ea-8939-be7ac1976d30.png)

reviewdog supports diff (unified format) as an input format especially useful
for [code suggestions](#code-suggestions).
reviewdog can integrate with any code suggestions tools or formatters to report suggestions.

`-f.diff.strip`: option for `-f=diff`: strip NUM leading components from diff file names (equivalent to 'patch -p') (default is 1 for git diff) (default 1)

```shell
$ <any-code-fixer/formatter> # e.g. eslint --fix, gofmt
$ TMPFILE=$(mktemp)
$ git diff >"${TMPFILE}"
$ git stash -u && git stash drop
$ reviewdog -f=diff -f.diff.strip=1 -reporter=github-pr-review < "${TMPFILE}"
```

Or you can also use [reviewdog/action-suggester](https://github.com/reviewdog/action-suggester) for GitHub Actions.

If diagnostic tools support diff output format, you can pipe the diff directly.

```shell
$ gofmt -s -d . | reviewdog -name="gofmt" -f=diff -f.diff.strip=0 -reporter=github-pr-review
$ shellcheck -f diff $(shfmt -f .) | reviewdog -f=diff
```

### checkstyle format

reviewdog also accepts [checkstyle XML format](http://checkstyle.sourceforge.net/) as well.
If the linter supports checkstyle format as a report format, you can use
-f=checkstyle instead of using 'errorformat'.

```shell
# Local
$ eslint -f checkstyle . | reviewdog -f=checkstyle -diff="git diff"

# CI (overwrite tool name which is shown in review comment by -name arg)
$ eslint -f checkstyle . | reviewdog -f=checkstyle -name="eslint" -reporter=github-check
```

Also, if you want to pass other Json/XML/etc... format to reviewdog, you can write a converter.

```shell
$ <linter> | <convert-to-checkstyle> | reviewdog -f=checkstyle -name="<linter>" -reporter=github-pr-check
```

## Code Suggestions

![eslint reviewdog suggestion demo](https://user-images.githubusercontent.com/3797062/97085944-87233a80-165b-11eb-94a8-0a47d5e24905.png)
![reviewdog with gofmt example](https://user-images.githubusercontent.com/3797062/89168305-a3ad5a80-d5b7-11ea-8939-be7ac1976d30.png)

reviewdog supports *code suggestions* feature with [rdformat](#reviewdog-diagnostic-format-rdformat) or [diff](#diff) input.
You can also use [reviewdog/action-suggester](https://github.com/reviewdog/action-suggester) for GitHub Actions.

reviewdog can suggest code changes along with diagnostic results if a diagnostic tools supports code suggestions data.
You can integrate reviewdog with any code fixing tools and any code formatter with [diff](#diff) input as well.

### Code Suggestions Support Table
Note that not all reporters provide support of code suggestion.

| `-reporter`     | Suggestion support |
| ---------------------------- | ------- |
| **`local`**                  | NO [1]  |
| **`github-check`**           | NO [2]  |
| **`github-pr-check`**        | NO [2]  |
| **`github-pr-review`**       | OK      |
| **`gitlab-mr-discussion`**   | NO [1]  |
| **`gitlab-mr-commit`**       | NO [2]  |
| **`gerrit-change-review`**   | NO [1]  |
| **`bitbucket-code-report`**  | NO [2]  |
| **`bitbucket-pr-review`**    | NO [2]  |

- [1] The reporter service support code suggestion feature, but reviewdog does not support it yet. See [#678](https://github.com/reviewdog/reviewdog/issues/678) for the status.
- [2] The reporter service itself doesn't support code suggestion feature.

## reviewdog config file

reviewdog can also be controlled via the .reviewdog.yml configuration file instead of "-f" or "-efm" arguments.

With .reviewdog.yml, you can run the same commands both CI service and local
environment including editor integration with ease.

#### .reviewdog.yml

```yaml
runner:
  <tool-name>:
    cmd: <command> # (required)
    errorformat: # (optional if you use `format`)
      - <list of errorformat>
    format: <format-name> # (optional if you use `errorformat`. e.g. golint,rdjson,rdjsonl)
    name: <tool-name> # (optional. you can overwrite <tool-name> defined by runner key)
    level: <level> # (optional. same as -level flag. [info,warning,error])

  # examples
  golint:
    cmd: golint ./...
    errorformat:
      - "%f:%l:%c: %m"
    level: warning
  govet:
    cmd: go vet -all .
    format: govet
  your-awesome-linter:
    cmd: awesome-linter run
    format: rdjson
    name: AwesomeLinter
```

```shell
$ reviewdog -diff="git diff FETCH_HEAD"
project/run_test.go:61:28: [golint] error strings should not end with punctuation
project/run.go:57:18: [errcheck]        defer os.Setenv(name, os.Getenv(name))
project/run.go:58:12: [errcheck]        os.Setenv(name, "")
# You can use -runners to run only specified runners.
$ reviewdog -diff="git diff FETCH_HEAD" -runners=golint,govet
project/run_test.go:61:28: [golint] error strings should not end with punctuation
# You can use -conf to specify config file path.
$ reviewdog -conf=./.reviewdog.yml -reporter=github-pr-check
```

Output format for project config based run is one of the following formats.

- `<file>: [<tool name>] <message>`
- `<file>:<lnum>: [<tool name>] <message>`
- `<file>:<lnum>:<col>: [<tool name>] <message>`

## Reporters

reviewdog can report results both in local environment and review services as
continuous integration.

### Reporter: Local (-reporter=local) [default]

reviewdog can find newly introduced findings by filtering linter results
using diff. You can pass diff command as `-diff` arg.

```shell
$ golint ./... | reviewdog -f=golint -diff="git diff FETCH_HEAD"
```

### Reporter: GitHub Checks (-reporter=github-pr-check)

[![github-pr-check sample annotation with option 1](https://user-images.githubusercontent.com/3797062/64875597-65016f80-d688-11e9-843f-4679fb666f0d.png)](https://github.com/reviewdog/reviewdog/pull/275/files#annotation_6177941961779419)
[![github-pr-check sample](https://user-images.githubusercontent.com/3797062/40884858-6efd82a0-6756-11e8-9f1a-c6af4f920fb0.png)](https://github.com/reviewdog/reviewdog/pull/131/checks)

github-pr-check reporter reports results to [GitHub Checks](https://help.github.com/articles/about-status-checks/).

You can change report level for this reporter by `level` field in [config
file](#reviewdog-config-file) or `-level` flag. You can control GitHub status
check result with this feature. (default: error)

| Level     | GitHub Status |
| --------- | ------------- |
| `info`    | neutral       |
| `warning` | neutral       |
| `error`   | failure       |

There are two options to use this reporter.

#### Option 1) Run reviewdog from GitHub Actions w/ secrets.GITHUB_TOKEN

Example: [.github/workflows/reviewdog.yml](.github/workflows/reviewdog.yml)

```yaml
- name: Run reviewdog
  env:
    REVIEWDOG_GITHUB_API_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    golint ./... | reviewdog -f=golint -reporter=github-pr-check
```

See [GitHub Actions](#github-actions) section too. You can also use public
reviewdog GitHub Actions.

#### Option 2) Install reviewdog GitHub Apps
reviewdog CLI send a request to reviewdog GitHub App server and the server post
results as GitHub Checks, because Check API only supported for GitHub App and
GitHub Actions.

1. Install reviewdog Apps. https://github.com/apps/reviewdog
2. Set `REVIEWDOG_TOKEN` or run reviewdog CLI in trusted CI providers.
  - Get token from `https://reviewdog.app/gh/{owner}/{repo-name}`.

```shell
$ export REVIEWDOG_TOKEN="<token>"
$ reviewdog -reporter=github-pr-check
```

Note: Token is not required if you run reviewdog in Travis or AppVeyor.

*Caution*

As described above, github-pr-check reporter with Option 2 depends on
reviewdog GitHub App server.
The server is running with haya14busa's pocket money for now and I may break
things, so I cannot ensure that the server is running 24h and 365 days.

**UPDATE:** Started getting support by [opencollective](https://opencollective.com/reviewdog)
and GitHub sponsor.
See [Supporting reviewdog](#supporting-reviewdog)

You can use github-pr-review reporter or use run reviewdog under GitHub Actions
if you don't want to depend on reviewdog server.

### Reporter: GitHub Checks (-reporter=github-check)

It's basically same as `-reporter=github-pr-check` except it works not only for
Pull Request but also for commit.

[![sample comment outside diff](https://user-images.githubusercontent.com/3797062/69917921-e0680580-14ae-11ea-9a56-de9e3cbac005.png)](https://github.com/reviewdog/reviewdog/pull/364/files)

You can create [reviewdog badge](#reviewdog-badge-) for this reporter.

### Reporter: GitHub PullRequest review comment (-reporter=github-pr-review)

[![sample-comment.png](https://raw.githubusercontent.com/haya14busa/i/dc0ccb1e110515ea407c146d99b749018db05c45/reviewdog/sample-comment.png)](https://github.com/reviewdog/reviewdog/pull/24#discussion_r84599728)

github-pr-review reporter reports results to GitHub PullRequest review comments
using GitHub Personal API Access Token.
[GitHub Enterprise](https://enterprise.github.com/home) is supported too.

- Go to https://github.com/settings/tokens and generate new API token.
- Check `repo` for private repositories or `public_repo` for public repositories.

```shell
$ export REVIEWDOG_GITHUB_API_TOKEN="<token>"
$ reviewdog -reporter=github-pr-review
```

For GitHub Enterprise, set API endpoint by environment variable.

```shell
$ export GITHUB_API="https://example.githubenterprise.com/api/v3/"
$ export REVIEWDOG_INSECURE_SKIP_VERIFY=true # set this as you need to skip verifying SSL
```

See [GitHub Actions](#github-actions) section too if you can use GitHub
Actions. You can also use public reviewdog GitHub Actions.

### Reporter: GitLab MergeRequest discussions (-reporter=gitlab-mr-discussion)

[![gitlab-mr-discussion sample](https://user-images.githubusercontent.com/3797062/41810718-f91bc540-773d-11e8-8598-fbc09ce9b1c7.png)](https://gitlab.com/haya14busa/reviewdog/merge_requests/113#note_83411103)

Required GitLab version: >= v10.8.0

gitlab-mr-discussion reporter reports results to GitLab MergeRequest discussions using
GitLab Personal API Access token.
Get the token with `api` scope from https://gitlab.com/profile/personal_access_tokens.

```shell
$ export REVIEWDOG_GITLAB_API_TOKEN="<token>"
$ reviewdog -reporter=gitlab-mr-discussion
```

The `CI_API_V4_URL` environment variable, defined automatically by Gitlab CI (v11.7 onwards), will be used to find out the Gitlab API URL.

Alternatively, `GITLAB_API` can also be defined, in which case it will take precedence over `CI_API_V4_URL`.

```shell
$ export GITLAB_API="https://example.gitlab.com/api/v4"
$ export REVIEWDOG_INSECURE_SKIP_VERIFY=true # set this as you need to skip verifying SSL
```

### Reporter: GitLab MergeRequest commit (-reporter=gitlab-mr-commit)

gitlab-mr-commit is similar to [gitlab-mr-discussion](#reporter-gitlab-mergerequest-discussions--reportergitlab-mr-discussion) reporter but reports results to each commit in GitLab MergeRequest.

gitlab-mr-discussion is recommended, but you can use gitlab-mr-commit reporter
if your GitLab version is under v10.8.0.

```shell
$ export REVIEWDOG_GITLAB_API_TOKEN="<token>"
$ reviewdog -reporter=gitlab-mr-commit
```

### Reporter: Gerrit Change review (-reporter=gerrit-change-review)

gerrit-change-review reporter reports result to Gerrit Change using Gerrit Rest APIs.

The reporter supports Basic Authentication and Git-cookie based authentication for reporting results.

Set `GERRIT_USERNAME` and `GERRIT_PASSWORD` environment variables for basic authentication, and put `GIT_GITCOOKIE_PATH` for git cookie based authentication.

```shell
$ export GERRIT_CHANGE_ID=changeID
$ export GERRIT_REVISION_ID=revisionID
$ export GERRIT_BRANCH=master
$ export GERRIT_ADDRESS=http://<gerrit-host>:<gerrit-port>
$ reviewdog -reporter=gerrit-change-review
```

### Reporter: Bitbucket Code Insights Reports (-reporter=bitbucket-code-report)

[![bitbucket-code-report](https://user-images.githubusercontent.com/9948629/96770123-c138d600-13e8-11eb-8e46-250b4bb393bd.png)](https://bitbucket.org/Trane9991/reviewdog-example/pull-requests/1)
[![bitbucket-code-annotations](https://user-images.githubusercontent.com/9948629/97054896-5e813f00-158e-11eb-9ad7-f8d75489b8ba.png)](https://bitbucket.org/Trane9991/reviewdog-example/pull-requests/1)

bitbucket-code-report generates the annotated
[Bitbucket Code Insights](https://support.atlassian.com/bitbucket-cloud/docs/code-insights/) report.

For now, only the `no-filter` mode supported, so the whole project is scanned on every run.
Reports are stored per commit and can be viewed per commit from Bitbucket Pipelines UI or
in Pull Request. In the Pull Request UI affected code lines will be annotated in the diff,
as well as you will be able to filter the annotations by **This pull request** or **All**.

If running from [Bitbucket Pipelines](#bitbucket-pipelines), no additional configuration is needed (even credentials).
If running locally or from some other CI system you would need to provide Bitbucket API credentials:

- For Basic Auth you need to set following env variables:
    `BITBUCKET_USER` and `BITBUCKET_PASSWORD`
- For AccessToken Auth you need to set `BITBUCKET_ACCESS_TOKEN`

```shell
$ export BITBUCKET_USER="my_user"
$ export BITBUCKET_PASSWORD="my_password"
$ reviewdog -reporter=bitbucket-code-report
```

### Reporter: Bitbucket Server PullRequest review comment  (-reporter=bitbucket-pr-review)

bitbucket-pr-review reporter reports results to Bitbucket Server PullRequest review comments

- For Basic Auth you need to set following env variables:
    `BITBUCKET_USER` and `BITBUCKET_PASSWORD`
- For AccessToken Auth you need to set `BITBUCKET_ACCESS_TOKEN`

```shell
$ export BITBUCKET_USER="my_user"
$ export BITBUCKET_PASSWORD="my_password"
$ export BITBUCKET_API="https://<bitbucket-host>/rest"
$ reviewdog -reporter=bitbucket-pr-review
```

## Supported CI services

### [GitHub Actions](https://github.com/features/actions)

Example: [.github/workflows/reviewdog.yml](.github/workflows/reviewdog.yml)

```yaml
name: reviewdog
on: [pull_request]
jobs:
  reviewdog:
    name: reviewdog
    runs-on: ubuntu-latest
    steps:
      # ...
      - uses: reviewdog/action-setup@v1
        with:
          reviewdog_version: latest # Optional. [latest,nightly,v.X.Y.Z]
      - name: Run reviewdog
        env:
          REVIEWDOG_GITHUB_API_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          reviewdog -reporter=github-pr-check -runners=golint,govet
          # or
          reviewdog -reporter=github-pr-review -runners=golint,govet
```

<details>
<summary><strong>Example (github-check reporter):</strong></summary>

[.github/workflows/reviewdog](.github/workflows/reviewdog.yml)

Only `github-check` reporter can run on push event too.

```yaml
name: reviewdog (github-check)
on:
  push:
    branches:
      - master
  pull_request:

jobs:
  reviewdog:
    name: reviewdog
    runs-on: ubuntu-latest
    steps:
      # ...
      - name: Run reviewdog
        env:
          REVIEWDOG_GITHUB_API_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          reviewdog -reporter=github-check -runners=golint,govet
```

</details>

#### Public Reviewdog GitHub Actions
You can use public GitHub Actions to start using reviewdog with ease! :tada: :arrow_forward: :tada:

- Common
  - [reviewdog/action-misspell](https://github.com/reviewdog/action-misspell) - Run [misspell](https://github.com/client9/misspell).
- Text (e.g. Markdown)
  - [reviewdog/action-alex](https://github.com/reviewdog/action-alex) - Run [alex](https://github.com/get-alex/alex) which catches insensitive, inconsiderate writing. (e.g. master/slave)
  - [reviewdog/action-languagetool](https://github.com/reviewdog/action-languagetool) - Run [languagetool](https://github.com/languagetool-org/languagetool).
  - [tsuyoshicho/action-textlint](https://github.com/tsuyoshicho/action-textlint) - Run [textlint](https://github.com/textlint/textlint)
  - [tsuyoshicho/action-redpen](https://github.com/tsuyoshicho/action-redpen) - Run [redpen](https://github.com/redpen-cc/redpen)
- Docker
  - [reviewdog/action-hadolint](https://github.com/reviewdog/action-hadolint) - Run [hadolint](https://github.com/hadolint/hadolint) to lint `Dockerfile`.
- Env
  - [dotenv-linter/action-dotenv-linter](https://github.com/dotenv-linter/action-dotenv-linter) - Run [dotenv-linter](https://github.com/dotenv-linter/dotenv-linter) to lint `.env` files.
- Shell script
  - [reviewdog/action-shellcheck](https://github.com/reviewdog/action-shellcheck) - Run [shellcheck](https://github.com/koalaman/shellcheck).
- Go
  - [reviewdog/action-staticcheck](https://github.com/reviewdog/action-staticcheck) - Run [staticcheck](https://staticcheck.io/).
  - [reviewdog/action-golangci-lint](https://github.com/reviewdog/action-golangci-lint) - Run [golangci-lint](https://github.com/golangci/golangci-lint) and supported linters individually by golangci-lint.
- JavaScript
  - [reviewdog/action-eslint](https://github.com/reviewdog/action-eslint) - Run [eslint](https://github.com/eslint/eslint).
- CSS
  - [reviewdog/action-stylelint](https://github.com/reviewdog/action-stylelint) - Run [stylelint](https://github.com/stylelint/stylelint).
- Vim script
  - [reviewdog/action-vint](https://github.com/reviewdog/action-vint) - Run [vint](https://github.com/Kuniwak/vint).
  - [tsuyoshicho/action-vimlint](https://github.com/tsuyoshicho/action-vimlint) - Run [vim-vimlint](https://github.com/syngan/vim-vimlint)
- Terraform
  - [reviewdog/action-tflint](https://github.com/reviewdog/action-tflint) - Run [tflint](https://github.com/wata727/tflint).
- YAML
  - [reviewdog/action-yamllint](https://github.com/reviewdog/action-yamllint) - Run [yamllint](https://github.com/adrienverge/yamllint).
- Ruby
  - [reviewdog/action-brakeman](https://github.com/reviewdog/action-brakeman) - Run [brakeman](https://github.com/presidentbeef/brakeman).
  - [reviewdog/action-reek](https://github.com/reviewdog/action-reek) - Run [reek](https://github.com/troessner/reek).
  - [reviewdog/action-rubocop](https://github.com/reviewdog/action-rubocop) - Run [rubocop](https://github.com/rubocop-hq/rubocop).
  - [vk26/action-fasterer](https://github.com/vk26/action-fasterer) - Run [fasterer](https://github.com/DamirSvrtan/fasterer).
  - [SennaLabs/action-standardrb](https://github.com/SennaLabs/action-standardrb) - Run [standardrb](https://github.com/testdouble/standard).

- Python
  - [wemake-python-styleguide](https://github.com/wemake-services/wemake-python-styleguide) - Run wemake-python-styleguide
  - [tsuyoshicho/action-mypy](https://github.com/tsuyoshicho/action-mypy) - Run [mypy](https://pypi.org/project/mypy/)
  - [jordemort/action-pyright](https://github.com/jordemort/action-pyright) - Run [pyright](https://github.com/Microsoft/pyright)
- Kotlin
  - [ScaCap/action-ktlint](https://github.com/ScaCap/action-ktlint) - Run [ktlint](https://ktlint.github.io/).
- Android Lint
  - [dvdandroid/action-android-lint](https://github.com/DVDAndroid/action-android-lint) - Run [Android Lint](https://developer.android.com/studio/write/lint)
- Ansible
  - [reviewdog/action-ansiblelint](https://github.com/reviewdog/action-ansiblelint) - Run [ansible-lint](https://github.com/ansible/ansible-lint)
  
... and more on [GitHub Marketplace](https://github.com/marketplace?utf8=✓&type=actions&query=reviewdog).

Missing actions? Check out [reviewdog/action-template](https://github.com/reviewdog/action-template) and create a new reviewdog action!

Please open a Pull Request to add your created reviewdog actions here :sparkles:.
I can also put your repositories under reviewdog org and co-maintain the actions.
Example: [action-tflint](https://github.com/reviewdog/reviewdog/issues/322).

#### Graceful Degradation for Pull Requests from forked repositories

![Graceful Degradation example](https://user-images.githubusercontent.com/3797062/71781334-e2266b00-3010-11ea-8a38-dee6e30c8162.png)

`GITHUB_TOKEN` for Pull Requests from forked repository doesn't have write
access to Check API nor Review API due to [GitHub Actions
restriction](https://docs.github.com/en/actions/configuring-and-managing-workflows/authenticating-with-the-github_token#permissions-for-the-github_token).

Instead, reviewdog uses [Logging commands of GitHub
Actions](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/development-tools-for-github-actions#set-an-error-message-error)
to post results as
[annotations](https://developer.github.com/v3/checks/runs/#annotations-object)
similar to `github-pr-check` reporter.

Note that there is a limitation for annotations created by logging commands,
such as [max # of annotations per run](https://github.com/reviewdog/reviewdog/issues/411#issuecomment-570893427).
You can check GitHub Actions log to see full results in such cases.

#### reviewdog badge [![reviewdog](https://github.com/reviewdog/reviewdog/workflows/reviewdog/badge.svg?branch=master&event=push)](https://github.com/reviewdog/reviewdog/actions?query=workflow%3Areviewdog+event%3Apush+branch%3Amaster)

As [`github-check` reporter](#reporter-github-checks--reportergithub-pr-check) support running on commit, we can create reviewdog
[GitHub Action badge](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/configuring-a-workflow#adding-a-workflow-status-badge-to-your-repository)
to check the result against master commit for example. :tada:

Example:
```
<!-- Replace <OWNER> and <REPOSITORY>. It assumes workflow name is "reviewdog" -->
[![reviewdog](https://github.com/<OWNER>/<REPOSITORY>/workflows/reviewdog/badge.svg?branch=master&event=push)](https://github.com/<OWNER>/<REPOSITORY>/actions?query=workflow%3Areviewdog+event%3Apush+branch%3Amaster)
```

### Travis CI

#### Travis CI (-reporter=github-pr-check)

If you use -reporter=github-pr-check in Travis CI, you don't need to set `REVIEWDOG_TOKEN`.

Example:

```yaml
install:
  - mkdir -p ~/bin/ && export PATH="~/bin/:$PATH"
  - curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh| sh -s -- -b ~/bin

script:
  - reviewdog -conf=.reviewdog.yml -reporter=github-pr-check
```

#### Travis CI (-reporter=github-pr-review)

Store GitHub API token by [travis encryption keys](https://docs.travis-ci.com/user/encryption-keys/).

```shell
$ gem install travis
$ travis encrypt REVIEWDOG_GITHUB_API_TOKEN=<token> --add env.global
```
Example:

```yaml
env:
  global:
    - secure: <token>

install:
  - mkdir -p ~/bin/ && export PATH="~/bin/:$PATH"
  - curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh| sh -s -- -b ~/bin

script:
  - >-
    golint ./... | reviewdog -f=golint -reporter=github-pr-review
```

Examples
- https://github.com/azu/textlint-reviewdog-example

### Circle CI

Store `REVIEWDOG_GITHUB_API_TOKEN` (or `REVIEWDOG_TOKEN` for github-pr-check) in
[Environment variables - CircleCI](https://circleci.com/docs/environment-variables/#setting-environment-variables-for-all-commands-without-adding-them-to-git)

#### .circleci/config.yml sample

```yaml
version: 2
jobs:
  build:
    docker:
      - image: golang:latest
    steps:
      - checkout
      - run: curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh| sh -s -- -b ./bin
      - run: go vet ./... 2>&1 | ./bin/reviewdog -f=govet -reporter=github-pr-review

      # Deprecated: prefer GitHub Actions to use github-pr-check reporter.
      - run: go vet ./... 2>&1 | ./bin/reviewdog -f=govet -reporter=github-pr-check
```

### GitLab CI

Store `REVIEWDOG_GITLAB_API_TOKEN` in [GitLab CI variable](https://docs.gitlab.com/ee/ci/variables/#variables).

#### .gitlab-ci.yml sample

```yaml
reviewdog:
  script:
    - reviewdog -reporter=gitlab-mr-discussion
    # Or
    - reviewdog -reporter=gitlab-mr-commit
```

### Bitbucket Pipelines

No additional configuration is needed.

#### bitbucket-pipelines.yml sample

```yaml
pipelines:
  default:
    - step:
        name: Reviewdog
        image: golangci/golangci-lint:v1.31-alpine
        script:
          - wget -O - -q https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh | 
              sh -s -- -b $(go env GOPATH)/bin
          - golangci-lint run --out-format=line-number ./... | reviewdog -f=golangci-lint -reporter=bitbucket-code-report
```

### Common (Jenkins, local, etc...)

You can use reviewdog to post review comments from anywhere with following
environment variables.

| name | description |
| ---- | ----------- |
| `CI_PULL_REQUEST` | Pull Request number (e.g. 14) |
| `CI_COMMIT`       | SHA1 for the current build |
| `CI_REPO_OWNER`   | repository owner (e.g. "haya14busa" for https://github.com/haya14busa/reviewdog) |
| `CI_REPO_NAME`    | repository name (e.g. "reviewdog" for https://github.com/haya14busa/reviewdog) |
| `CI_BRANCH`       | [optional] branch of the commit |

```shell
$ export CI_PULL_REQUEST=14
$ export CI_REPO_OWNER=haya14busa
$ export CI_REPO_NAME=reviewdog
$ export CI_COMMIT=$(git rev-parse HEAD)
```
and set a token if required.

```shell
$ REVIEWDOG_TOKEN="<token>"
$ REVIEWDOG_GITHUB_API_TOKEN="<token>"
$ REVIEWDOG_GITLAB_API_TOKEN="<token>"
```

If a CI service doesn't provide information such as Pull Request ID - reviewdog can guess it by branch name and commit SHA.
Just pass the flag `guess`:

```shell
$ reviewdog -conf=.reviewdog.yml -reporter=github-pr-check -guess
```

#### Jenkins with Github pull request builder plugin
- [GitHub pull request builder plugin - Jenkins - Jenkins Wiki](https://wiki.jenkins-ci.org/display/JENKINS/GitHub+pull+request+builder+plugin)

```shell
$ export CI_PULL_REQUEST=${ghprbPullId}
$ export CI_REPO_OWNER=haya14busa
$ export CI_REPO_NAME=reviewdog
$ export CI_COMMIT=${ghprbActualCommit}
$ export REVIEWDOG_INSECURE_SKIP_VERIFY=true # set this as you need
$ REVIEWDOG_TOKEN="<token>" reviewdog -reporter=github-pr-check
# Or
$ REVIEWDOG_GITHUB_API_TOKEN="<token>" reviewdog -reporter=github-pr-review
```

## Exit codes
By default reviewdog will return `0` as exit code even if it finds errors.
If `-fail-on-error` flag is passed, reviewdog exits with `1` when at least one error was found/reported.
This can be helpful when you are using it as a step in your CI pipeline and want to mark the step failed if any error found by linter.

See also `-level` flag for [github-pr-check/github-check](#reporter-github-checks--reportergithub-pr-check) reporters.
reviewdog will exit with `1` if reported check status is `failure` as well if `-fail-on-error=true`.

## Filter mode
reviewdog filter results by diff and you can control how reviewdog filter results by `-filter-mode` flag.
Available filter modes are as below.

### `added` (default)
Filter results by added/modified lines.
### `diff_context`
Filter results by diff context. i.e. changed lines +-N lines (N=3 for example).
### `file`
Filter results by added/modified file. i.e. reviewdog will report results as long as they are in added/modified file even if the results are not in actual diff.
### `nofilter`
Do not filter any results. Useful for posting results as comments as much as possible and check other results in console at the same time.

`-fail-on-error` also works with any filter-mode and can catch all results from any linters with `nofilter` mode.

Example:
```shell
$ reviewdog -reporter=github-pr-review -filter-mode=nofilter -fail-on-error
```

### Filter Mode Support Table
Note that not all reporters provide full support of filter mode due to API limitation.
e.g. `github-pr-review` reporter uses [GitHub Review
API](https://developer.github.com/v3/pulls/reviews/) but it doesn't support posting comment outside diff (`diff_context`),
so reviewdog will use [Check annotation](https://developer.github.com/v3/checks/runs/) as fallback to post those comments [1]. 

| `-reporter` \ `-filter-mode` | `added` | `diff_context` | `file`                  | `nofilter` |
| ---------------------------- | ------- | -------------- | ----------------------- | ---------- |
| **`local`**                  | OK      | OK             | OK                      | OK |
| **`github-check`**           | OK      | OK             | OK                      | OK |
| **`github-pr-check`**        | OK      | OK             | OK                      | OK |
| **`github-pr-review`**       | OK      | OK             | Partially Supported [1] | Partially Supported [1] |
| **`gitlab-mr-discussion`**   | OK      | OK             | OK                      | Partially Supported [2] |
| **`gitlab-mr-commit`**       | OK      | Partially Supported [2] | Partially Supported [2] | Partially Supported [2] |
| **`gerrit-change-review`**   | OK      | OK? [3]        | OK? [3]                 | Partially Supported? [2][3] |
| **`bitbucket-code-report`**  | NO [4]  | NO [4]         | NO [4]                  | OK |
| **`bitbucket-pr-review`**    | OK      | OK             | OK                      | Partially Supported [2] |

- [1] Report results which is outside diff context with Check annotation as fallback if it's running in GitHub actions instead of Review API (comments). All results will be reported to console as well.
- [2] Report results which is outside diff file to console.
- [3] It should work, but not verified yet.
- [4] Not implemented at the moment

## Debugging

Use the `-tee` flag to show debug info.

```shell
reviewdog -filter-mode=nofilter -tee
```

## Articles
- [reviewdog — A code review dog who keeps your codebase healthy ](https://medium.com/@haya14busa/reviewdog-a-code-review-dog-who-keeps-your-codebase-healthy-d957c471938b)
- [reviewdog ♡ GitHub Check — improved automated review experience](https://medium.com/@haya14busa/reviewdog-github-check-improved-automated-review-experience-58f89e0c95f3)
- [Automated Code Review on GitHub Actions with reviewdog for any languages/tools](https://medium.com/@haya14busa/automated-code-review-on-github-actions-with-reviewdog-for-any-languages-tools-20285e04448e)
- [GitHub Actions to guard your workflow](https://evrone.com/github-actions)

## :bird: Author
haya14busa [![GitHub followers](https://img.shields.io/github/followers/haya14busa.svg?style=social&label=Follow)](https://github.com/haya14busa)

## Contributors

[![Contributors](https://opencollective.com/reviewdog/contributors.svg?width=890)](https://github.com/reviewdog/reviewdog/graphs/contributors)

### Supporting reviewdog

Become GitHub Sponsor for [each contributor](https://github.com/reviewdog/reviewdog/graphs/contributors)
or become a backer or sponsor from [opencollective](https://opencollective.com/reviewdog).

[![Become a backer](https://opencollective.com/reviewdog/tiers/backer.svg?avatarHeight=64)](https://opencollective.com/reviewdog#backers)
