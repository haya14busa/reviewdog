package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/tabwriter"

	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
	"golang.org/x/build/gerrit"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v37/github"
	"github.com/mattn/go-shellwords"
	"github.com/reviewdog/errorformat/fmts"
	"github.com/xanzy/go-gitlab"

	"github.com/reviewdog/go-bitbucket"
	"github.com/reviewdog/reviewdog"
	"github.com/reviewdog/reviewdog/cienv"
	"github.com/reviewdog/reviewdog/commands"
	"github.com/reviewdog/reviewdog/filter"
	"github.com/reviewdog/reviewdog/parser"
	"github.com/reviewdog/reviewdog/project"
	bbservice "github.com/reviewdog/reviewdog/service/bitbucket"
	gerritservice "github.com/reviewdog/reviewdog/service/gerrit"
	githubservice "github.com/reviewdog/reviewdog/service/github"
	"github.com/reviewdog/reviewdog/service/github/githubutils"
	gitlabservice "github.com/reviewdog/reviewdog/service/gitlab"
)

const usageMessage = "" +
	`Usage:	reviewdog [flags]
	reviewdog accepts any compiler or linter results from stdin and filters
	them by diff for review. reviewdog also can posts the results as a comment to
	GitHub if you use reviewdog in CI service.`

type option struct {
	version          bool
	diffCmd          string
	diffStrip        int
	efms             strslice
	f                string // format name
	fDiffStrip       int
	list             bool   // list supported errorformat name
	name             string // tool name which is used in comment
	conf             string
	runners          string
	reporter         string
	level            string
	guessPullRequest bool
	tee              bool
	filterMode       filter.Mode
	failOnError      bool
}

const (
	diffCmdDoc    = `diff command (e.g. "git diff") for local reporter. Do not use --relative flag for git command.`
	diffStripDoc  = "strip NUM leading components from diff file names (equivalent to 'patch -p') (default is 1 for git diff)"
	efmsDoc       = `list of supported machine-readable format and errorformat (https://github.com/reviewdog/errorformat)`
	fDoc          = `format name (run -list to see supported format name) for input. It's also used as tool name in review comment if -name is empty`
	fDiffStripDoc = `option for -f=diff: strip NUM leading components from diff file names (equivalent to 'patch -p') (default is 1 for git diff)`
	listDoc       = `list supported pre-defined format names which can be used as -f arg`
	nameDoc       = `tool name in review comment. -f is used as tool name if -name is empty`

	confDoc             = `config file path`
	runnersDoc          = `comma separated runners name to run in config file. default: run all runners`
	levelDoc            = `report level currently used for github-pr-check reporter ("info","warning","error").`
	guessPullRequestDoc = `guess Pull Request ID by branch name and commit SHA`
	teeDoc              = `enable "tee"-like mode which outputs tools's output as is while reporting results to -reporter. Useful for debugging as well.`
	filterModeDoc       = `how to filter checks results. [added, diff_context, file, nofilter].
		"added" (default)
			Filter by added/modified diff lines.
		"diff_context"
			Filter by diff context, which can include unchanged lines.
			i.e. changed lines +-N lines (e.g. N=3 for default git diff).
		"file"
			Filter by added/modified file.
		"nofilter"
			Do not filter any results.
`
	reporterDoc = `reporter of reviewdog results. (local, github-check, github-pr-check, github-pr-review, gitlab-mr-discussion, gitlab-mr-commit)
	"local" (default)
		Report results to stdout.

	"github-check"
		Report results to GitHub Check. It works both for Pull Requests and commits.
		For Pull Request, you can see report results in GitHub PullRequest Check
		tab and can control filtering mode by -filter-mode flag.

		There are two options to use this reporter.

		Option 1) Run reviewdog from GitHub Actions w/ secrets.GITHUB_TOKEN
			Note that it reports result to GitHub Actions log console for Pull
			Requests from fork repository due to GitHub Actions restriction.
			https://help.github.com/en/articles/virtual-environments-for-github-actions#github_token-secret

			Set REVIEWDOG_GITHUB_API_TOKEN with secrets.GITHUB_TOKEN. e.g.
					REVIEWDOG_GITHUB_API_TOKEN: ${{ secrets.GITHUB_TOKEN }}

		Option 2) Install reviewdog GitHub Apps
			1. Install reviewdog Apps. https://github.com/apps/reviewdog
			2. Set REVIEWDOG_TOKEN or run reviewdog CLI in trusted CI providers.
			You can get token from https://reviewdog.app/gh/<owner>/<repo-name>.
			$ export REVIEWDOG_TOKEN="xxxxx"

			Note: Token is not required if you run reviewdog in Travis CI.

	"github-pr-check"
		Same as github-check reporter but it only supports Pull Requests.

	"github-pr-review"
		Report results to GitHub review comments.

		1. Set REVIEWDOG_GITHUB_API_TOKEN environment variable.
		Go to https://github.com/settings/tokens and create new Personal access token with repo scope.

		For GitHub Enterprise:
			$ export GITHUB_API="https://example.githubenterprise.com/api/v3"

	"gitlab-mr-discussion"
		Report results to GitLab MergeRequest discussion.

		1. Set REVIEWDOG_GITLAB_API_TOKEN environment variable.
		Go to https://gitlab.com/profile/personal_access_tokens

		CI_API_V4_URL (defined by Gitlab CI) as the base URL for the Gitlab API automatically.
		Alternatively, GITLAB_API can also be defined, and it will take precedence over the former:
			$ export GITLAB_API="https://example.gitlab.com/api/v4"

	"gitlab-mr-commit"
		Same as gitlab-mr-discussion, but report results to GitLab comments for
		each commits in Merge Requests.

	"gerrit-change-review"
		Report results to Gerrit Change comments.

		1. Set GERRIT_USERNAME and GERRIT_PASSWORD for basic authentication or
		GIT_GITCOOKIE_PATH for git cookie based authentication.
		2. Set GERRIT_CHANGE_ID, GERRIT_REVISION_ID GERRIT_BRANCH abd GERRIT_ADDRESS

		For example:
			$ export GERRIT_CHANGE_ID=myproject~master~I1293efab014de2
			$ export GERRIT_REVISION_ID=ed318bf9a3c
			$ export GERRIT_BRANCH=master
			$ export GERRIT_ADDRESS=http://localhost:8080
	
	"bitbucket-code-report"
		Create Bitbucket Code Report via Code Insights
		(https://confluence.atlassian.com/display/BITBUCKET/Code+insights).
		You can set custom report name with:

		If running as part of Bitbucket Pipelines no additional configurations is needed.
		If running outside of Bitbucket Pipelines you need to provide git repo data
		(see documentation below for local reporters) and BitBucket credentials:
		- For Basic Auth you need to set following env variables:
			  BITBUCKET_USER and BITBUCKET_PASSWORD
		- For AccessToken Auth you need to set BITBUCKET_ACCESS_TOKEN 
		Running on Bitbucket Server is not tested/supported yet.

    "bitbucket-pr-review"
		Report results to Bitbucket Server review comments.

		1. Set BITBUCKET_USER and BITBUCKET_PASSWORD for basic authentication or
		BITBUCKET_ACCESS_TOKEN for AccessToken Auth authentication.
		2. Set BITBUCKET_API url

	For GitHub Enterprise and self hosted GitLab, set
	REVIEWDOG_INSECURE_SKIP_VERIFY to skip verifying SSL (please use this at your own risk)
		$ export REVIEWDOG_INSECURE_SKIP_VERIFY=true

	For non-local reporters, reviewdog automatically get necessary data from
	environment variable in CI service (GitHub Actions, Travis CI, Circle CI, drone.io, GitLab CI, Bitbucket Pipelines).
	You can set necessary data with following environment variable manually if
	you want (e.g. run reviewdog in Jenkins).

		$ export CI_PULL_REQUEST=14 # Pull Request number (e.g. 14)
		$ export CI_COMMIT="$(git rev-parse @)" # SHA1 for the current build
		$ export CI_REPO_OWNER="haya14busa" # repository owner
		$ export CI_REPO_NAME="reviewdog" # repository name
`
	failOnErrorDoc = `Returns 1 as exit code if any errors/warnings found in input`
)

var opt = &option{}

func init() {
	flag.BoolVar(&opt.version, "version", false, "print version")
	flag.StringVar(&opt.diffCmd, "diff", "", diffCmdDoc)
	flag.IntVar(&opt.diffStrip, "strip", 1, diffStripDoc)
	flag.Var(&opt.efms, "efm", efmsDoc)
	flag.StringVar(&opt.f, "f", "", fDoc)
	flag.IntVar(&opt.fDiffStrip, "f.diff.strip", 1, fDiffStripDoc)
	flag.BoolVar(&opt.list, "list", false, listDoc)
	flag.StringVar(&opt.name, "name", "", nameDoc)
	flag.StringVar(&opt.conf, "conf", "", confDoc)
	flag.StringVar(&opt.runners, "runners", "", runnersDoc)
	flag.StringVar(&opt.reporter, "reporter", "local", reporterDoc)
	flag.StringVar(&opt.level, "level", "error", levelDoc)
	flag.BoolVar(&opt.guessPullRequest, "guess", false, guessPullRequestDoc)
	flag.BoolVar(&opt.tee, "tee", false, teeDoc)
	flag.Var(&opt.filterMode, "filter-mode", filterModeDoc)
	flag.BoolVar(&opt.failOnError, "fail-on-error", false, failOnErrorDoc)
}

func usage() {
	fmt.Fprintln(os.Stderr, usageMessage)
	fmt.Fprintln(os.Stderr, "Flags:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "See https://github.com/reviewdog/reviewdog for more detail.")
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if err := run(os.Stdin, os.Stdout, opt); err != nil {
		fmt.Fprintf(os.Stderr, "reviewdog: %v\n", err)
		os.Exit(1)
	}
}

func run(r io.Reader, w io.Writer, opt *option) error {
	ctx := context.Background()

	if opt.version {
		fmt.Fprintln(w, commands.Version)
		return nil
	}

	if opt.list {
		return runList(w)
	}

	if opt.tee {
		r = io.TeeReader(r, w)
	}

	// assume it's project based run when both -efm and -f are not specified
	isProject := len(opt.efms) == 0 && opt.f == ""
	var projectConf *project.Config

	var cs reviewdog.CommentService
	var ds reviewdog.DiffService

	if isProject {
		var err error
		projectConf, err = projectConfig(opt.conf)
		if err != nil {
			return err
		}

		cs = reviewdog.NewUnifiedCommentWriter(w)
	} else {
		cs = reviewdog.NewRawCommentWriter(w)
	}

	switch opt.reporter {
	default:
		return fmt.Errorf("unknown -reporter: %s", opt.reporter)
	case "github-check":
		return runDoghouse(ctx, r, w, opt, isProject, false)
	case "github-pr-check":
		return runDoghouse(ctx, r, w, opt, isProject, true)
	case "github-pr-review":
		gs, isPR, err := githubService(ctx, opt)
		if err != nil {
			return err
		}
		if !isPR {
			fmt.Fprintln(os.Stderr, "reviewdog: this is not PullRequest build.")
			return nil
		}
		// If it's running in GitHub Actions and it's PR from forked repository,
		// replace comment writer to GitHubActionLogWriter to create annotations
		// instead of review comment because if it's PR from forked repository,
		// GitHub token doesn't have write permission due to security concern and
		// cannot post results via Review API.
		if cienv.IsInGitHubAction() && cienv.HasReadOnlyPermissionGitHubToken() {
			fmt.Fprintln(w, `reviewdog: This GitHub token doesn't have write permission of Review API [1], 
so reviewdog will report results via logging command [2] and create annotations similar to
github-pr-check reporter as a fallback.
[1]: https://docs.github.com/en/actions/reference/events-that-trigger-workflows#pull_request_target, 
[2]: https://help.github.com/en/actions/automating-your-workflow-with-github-actions/development-tools-for-github-actions#logging-commands`)
			cs = githubutils.NewGitHubActionLogWriter(opt.level)
		} else {
			cs = reviewdog.MultiCommentService(gs, cs)
		}
		ds = gs
	case "gitlab-mr-discussion":
		build, cli, err := gitlabBuildWithClient()
		if err != nil {
			return err
		}
		if build.PullRequest == 0 {
			fmt.Fprintln(os.Stderr, "this is not MergeRequest build.")
			return nil
		}

		gc, err := gitlabservice.NewGitLabMergeRequestDiscussionCommenter(cli, build.Owner, build.Repo, build.PullRequest, build.SHA)
		if err != nil {
			return err
		}

		cs = reviewdog.MultiCommentService(gc, cs)
		ds, err = gitlabservice.NewGitLabMergeRequestDiff(cli, build.Owner, build.Repo, build.PullRequest, build.SHA)
		if err != nil {
			return err
		}
	case "gitlab-mr-commit":
		build, cli, err := gitlabBuildWithClient()
		if err != nil {
			return err
		}
		if build.PullRequest == 0 {
			fmt.Fprintln(os.Stderr, "this is not MergeRequest build.")
			return nil
		}

		gc, err := gitlabservice.NewGitLabMergeRequestCommitCommenter(cli, build.Owner, build.Repo, build.PullRequest, build.SHA)
		if err != nil {
			return err
		}

		cs = reviewdog.MultiCommentService(gc, cs)
		ds, err = gitlabservice.NewGitLabMergeRequestDiff(cli, build.Owner, build.Repo, build.PullRequest, build.SHA)
		if err != nil {
			return err
		}
	case "gerrit-change-review":
		b, cli, err := gerritBuildWithClient()
		if err != nil {
			return err
		}
		gc, err := gerritservice.NewChangeReviewCommenter(cli, b.GerritChangeID, b.GerritRevisionID)
		if err != nil {
			return err
		}
		cs = gc

		d, err := gerritservice.NewChangeDiff(cli, b.Branch, b.GerritChangeID)
		if err != nil {
			return err
		}
		ds = d
	case "bitbucket-code-report":
		build, client, ct, err := bitbucketBuildWithClient(ctx)
		if err != nil {
			return err
		}
		ctx = ct

		cs = bbservice.NewReportAnnotator(client,
			build.Owner, build.Repo, build.SHA, getRunnersList(opt, projectConf))

		if !(opt.filterMode == filter.ModeDefault || opt.filterMode == filter.ModeNoFilter) {
			// by default scan whole project with out diff (filter.ModeNoFilter)
			// Bitbucket pipelines doesn't give an easy way to know
			// which commit run pipeline before so we can compare between them
			// however once PR is opened, Bitbucket Reports UI will do automatic
			// filtering of annotations dividing them in two groups:
			// - This pull request (10)
			// - All (50)
			log.Printf("reviewdog: [bitbucket-code-report] supports only with filter.ModeNoFilter for now")
		}
		opt.filterMode = filter.ModeNoFilter
		ds = &reviewdog.EmptyDiff{}
	case "bitbucket-pr-review":
		build, cli, err := bitbucketv1BuildWithClient(ctx)
		if err != nil {
			return err
		}
		if build.PullRequest == 0 {
			fmt.Fprintln(os.Stderr, "this is not PullRequest build.")
			return nil
		}

		gc, err := bbservice.NewPullRequestCommenter(cli, build.Owner, build.Repo, build.PullRequest, build.SHA)
		if err != nil {
			return err
		}

		cs = reviewdog.MultiCommentService(gc, cs)
		ds, err = bbservice.NewPullRequestDiff(cli, build.Owner, build.Repo, build.PullRequest, build.SHA)
		if err != nil {
			return err
		}
	case "local":
		if opt.diffCmd == "" && opt.filterMode == filter.ModeNoFilter {
			ds = &reviewdog.EmptyDiff{}
		} else {
			d, err := diffService(opt.diffCmd, opt.diffStrip)
			if err != nil {
				return err
			}
			ds = d
		}
	}

	if isProject {
		return project.Run(ctx, projectConf, buildRunnersMap(opt.runners), cs, ds, opt.tee, opt.filterMode, opt.failOnError)
	}

	p, err := newParserFromOpt(opt)
	if err != nil {
		return err
	}

	app := reviewdog.NewReviewdog(toolName(opt), p, cs, ds, opt.filterMode, opt.failOnError)
	return app.Run(ctx, r)
}

func runList(w io.Writer) error {
	tabw := tabwriter.NewWriter(w, 0, 8, 0, '\t', 0)
	fmt.Fprintf(tabw, "%s\t%s\t- %s\n", "rdjson", "Reviewdog Diagnostic JSON Format (JSON of DiagnosticResult message)", "https://github.com/reviewdog/reviewdog")
	fmt.Fprintf(tabw, "%s\t%s\t- %s\n", "rdjsonl", "Reviewdog Diagnostic JSONL Format (JSONL of Diagnostic message)", "https://github.com/reviewdog/reviewdog")
	fmt.Fprintf(tabw, "%s\t%s\t- %s\n", "diff", "Unified Diff Format", "https://en.wikipedia.org/wiki/Diff#Unified_format")
	fmt.Fprintf(tabw, "%s\t%s\t- %s\n", "checkstyle", "checkstyle XML format", "http://checkstyle.sourceforge.net/")
	for _, f := range sortedFmts(fmts.DefinedFmts()) {
		fmt.Fprintf(tabw, "%s\t%s\t- %s\n", f.Name, f.Description, f.URL)
	}
	return tabw.Flush()
}

type byFmtName []*fmts.Fmt

func (p byFmtName) Len() int           { return len(p) }
func (p byFmtName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p byFmtName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func sortedFmts(fs fmts.Fmts) []*fmts.Fmt {
	r := make([]*fmts.Fmt, 0, len(fs))
	for _, f := range fs {
		r = append(r, f)
	}
	sort.Sort(byFmtName(r))
	return r
}

func diffService(s string, strip int) (reviewdog.DiffService, error) {
	cmds, err := shellwords.Parse(s)
	if err != nil {
		return nil, err
	}
	if len(cmds) < 1 {
		return nil, errors.New("diff command is empty")
	}
	cmd := exec.Command(cmds[0], cmds[1:]...)
	d := reviewdog.NewDiffCmd(cmd, strip)
	return d, nil
}

func newHTTPClient() *http.Client {
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify()},
	}
	return &http.Client{Transport: tr}
}

func insecureSkipVerify() bool {
	return os.Getenv("REVIEWDOG_INSECURE_SKIP_VERIFY") == "true"
}

func githubService(ctx context.Context, opt *option) (gs *githubservice.PullRequest, isPR bool, err error) {
	token, err := nonEmptyEnv("REVIEWDOG_GITHUB_API_TOKEN")
	if err != nil {
		return nil, isPR, err
	}
	g, isPR, err := cienv.GetBuildInfo()
	if err != nil {
		return nil, isPR, err
	}

	client, err := githubClient(ctx, token)
	if err != nil {
		return nil, isPR, err
	}

	if !isPR {
		if !opt.guessPullRequest {
			return nil, false, nil
		}

		if g.Branch == "" && g.SHA == "" {
			return nil, false, nil
		}

		prID, err := getPullRequestIDByBranchOrCommit(ctx, client, g)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, false, nil
		}
		g.PullRequest = prID
	}

	gs, err = githubservice.NewGitHubPullRequest(client, g.Owner, g.Repo, g.PullRequest, g.SHA)
	if err != nil {
		return nil, false, err
	}
	return gs, true, nil
}

func getPullRequestIDByBranchOrCommit(ctx context.Context, client *github.Client, info *cienv.BuildInfo) (int, error) {
	options := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
	}

	query := []string{
		"type:pr",
		"state:open",
		fmt.Sprintf("repo:%s/%s", info.Owner, info.Repo),
	}
	if info.Branch != "" {
		query = append(query, fmt.Sprintf("head:%s", info.Branch))
	}
	if info.SHA != "" {
		query = append(query, info.SHA)
	}

	preparedQuery := strings.Join(query, " ")
	pullRequests, _, err := client.Search.Issues(ctx, preparedQuery, options)
	if err != nil {
		return 0, err
	}

	if *pullRequests.Total == 0 {
		return 0, fmt.Errorf("reviewdog: PullRequest not found, query: %s", preparedQuery)
	}

	return *pullRequests.Issues[0].Number, nil
}

func githubClient(ctx context.Context, token string) (*github.Client, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, newHTTPClient())
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	var err error
	client.BaseURL, err = githubBaseURL()
	return client, err
}

const defaultGitHubAPI = "https://api.github.com/"

func githubBaseURL() (*url.URL, error) {
	if baseURL := os.Getenv("GITHUB_API"); baseURL != "" {
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("GitHub base URL from GITHUB_API is invalid: %v, %w", baseURL, err)
		}
		return u, nil
	}
	// get GitHub base URL from GitHub Actions' default environment variable GITHUB_API_URL
	// ref: https://docs.github.com/en/actions/reference/environment-variables#default-environment-variables
	if baseURL := os.Getenv("GITHUB_API_URL"); baseURL != "" {
		u, err := url.Parse(baseURL + "/")
		if err != nil {
			return nil, fmt.Errorf("GitHub base URL from GITHUB_API_URL is invalid: %v, %w", baseURL, err)
		}
		return u, nil
	}
	u, err := url.Parse(defaultGitHubAPI)
	if err != nil {
		return nil, fmt.Errorf("GitHub base URL from reviewdog default is invalid: %v, %w", defaultGitHubAPI, err)
	}
	return u, nil
}

func gitlabBuildWithClient() (*cienv.BuildInfo, *gitlab.Client, error) {
	token, err := nonEmptyEnv("REVIEWDOG_GITLAB_API_TOKEN")
	if err != nil {
		return nil, nil, err
	}

	g, _, err := cienv.GetBuildInfo()
	if err != nil {
		return nil, nil, err
	}

	client, err := gitlabClient(token)
	if err != nil {
		return nil, nil, err
	}

	if g.PullRequest == 0 {
		prNr, err := fetchMergeRequestIDFromCommit(client, g.Owner+"/"+g.Repo, g.SHA)
		if err != nil {
			return nil, nil, err
		}
		if prNr != 0 {
			g.PullRequest = prNr
		}
	}

	return g, client, err
}

func gerritBuildWithClient() (*cienv.BuildInfo, *gerrit.Client, error) {
	buildInfo, err := cienv.GetGerritBuildInfo()
	if err != nil {
		return nil, nil, err
	}

	gerritAddr := os.Getenv("GERRIT_ADDRESS")
	if gerritAddr == "" {
		return nil, nil, errors.New("cannot get gerrit host address from environment variable. Set GERRIT_ADDRESS ?")
	}

	username := os.Getenv("GERRIT_USERNAME")
	password := os.Getenv("GERRIT_PASSWORD")
	if username != "" && password != "" {
		client := gerrit.NewClient(gerritAddr, gerrit.BasicAuth(username, password))
		return buildInfo, client, nil
	}

	if useGitCookiePath := os.Getenv("GERRIT_GIT_COOKIE_PATH"); useGitCookiePath != "" {
		client := gerrit.NewClient(gerritAddr, gerrit.GitCookieFileAuth(useGitCookiePath))
		return buildInfo, client, nil
	}

	client := gerrit.NewClient(gerritAddr, gerrit.NoAuth)
	return buildInfo, client, nil
}

func bitbucketBuildWithClient(ctx context.Context) (*cienv.BuildInfo, *bitbucket.APIClient, context.Context, error) {
	build, _, err := cienv.GetBuildInfo()
	if err != nil {
		return nil, nil, ctx, err
	}

	bbUser := os.Getenv("BITBUCKET_USER")
	bbPass := os.Getenv("BITBUCKET_PASSWORD")
	bbAccessToken := os.Getenv("BITBUCKET_ACCESS_TOKEN")

	if bbUser != "" && bbPass != "" {
		ctx = bbservice.WithBasicAuth(ctx, bbUser, bbPass)
	}

	if bbAccessToken != "" {
		ctx = bbservice.WithAccessToken(ctx, bbAccessToken)
	}

	client := bbservice.NewAPIClient(cienv.IsInBitbucketPipeline(), cienv.IsInBitbucketPipe())
	return build, client, ctx, nil
}

func bitbucketv1BuildWithClient(ctx context.Context) (*cienv.BuildInfo, *bitbucketv1.APIClient, error) {
	build, _, err := cienv.GetBuildInfo()
	if err != nil {
		return nil, nil, err
	}

	bbBasePath := os.Getenv("BITBUCKET_API")
	if bbBasePath == "" {
		return nil, nil, errors.New("cannot get bitbucket v1 api address from environment variable. Set BITBUCKET_API ?")
	}

	bbUser := os.Getenv("BITBUCKET_USER")
	bbPass := os.Getenv("BITBUCKET_PASSWORD")
	bbAccessToken := os.Getenv("BITBUCKET_ACCESS_TOKEN")

	if bbUser != "" && bbPass != "" {
		basicAuth := bitbucketv1.BasicAuth{UserName: bbUser, Password: bbPass}
		ctx = context.WithValue(ctx, bitbucketv1.ContextBasicAuth, basicAuth)
	}

	if bbAccessToken != "" {
		ctx = context.WithValue(ctx, bitbucketv1.ContextAccessToken, bbAccessToken)
	}

	client := bitbucketv1.NewAPIClient(ctx, bitbucketv1.NewConfiguration(bbBasePath))
	return build, client, nil
}

func fetchMergeRequestIDFromCommit(cli *gitlab.Client, projectID, sha string) (id int, err error) {
	// https://docs.gitlab.com/ce/api/merge_requests.html#list-project-merge-requests
	opt := &gitlab.ListProjectMergeRequestsOptions{
		State:   gitlab.String("opened"),
		OrderBy: gitlab.String("updated_at"),
	}
	mrs, _, err := cli.MergeRequests.ListProjectMergeRequests(projectID, opt)
	if err != nil {
		return 0, err
	}
	for _, mr := range mrs {
		if mr.SHA == sha {
			return mr.IID, nil
		}
	}
	return 0, nil
}

func gitlabClient(token string) (*gitlab.Client, error) {
	baseURL, err := gitlabBaseURL()
	if err != nil {
		return nil, err
	}
	client, err := gitlab.NewClient(token, gitlab.WithHTTPClient(newHTTPClient()), gitlab.WithBaseURL(baseURL.String()))
	if err != nil {
		return nil, err
	}
	return client, nil
}

const defaultGitLabAPI = "https://gitlab.com/api/v4"

func gitlabBaseURL() (*url.URL, error) {
	gitlabAPI := os.Getenv("GITLAB_API")
	gitlabV4URL := os.Getenv("CI_API_V4_URL")

	var baseURL string
	if gitlabAPI != "" {
		baseURL = gitlabAPI
	} else if gitlabV4URL != "" {
		baseURL = gitlabV4URL
	} else {
		baseURL = defaultGitLabAPI
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("GitLab base URL is invalid: %v, %w", baseURL, err)
	}
	return u, nil
}

func nonEmptyEnv(env string) (string, error) {
	v := os.Getenv(env)
	if v == "" {
		return "", fmt.Errorf("environment variable $%v is not set", env)
	}
	return v, nil
}

type strslice []string

func (ss *strslice) String() string {
	return fmt.Sprintf("%v", *ss)
}

func (ss *strslice) Set(value string) error {
	*ss = append(*ss, value)
	return nil
}

func projectConfig(path string) (*project.Config, error) {
	b, err := readConf(path)
	if err != nil {
		return nil, fmt.Errorf("fail to open config: %w", err)
	}
	conf, err := project.Parse(b)
	if err != nil {
		return nil, fmt.Errorf("config is invalid: %w", err)
	}
	return conf, nil
}

func readConf(conf string) ([]byte, error) {
	var conffiles []string
	if conf != "" {
		conffiles = []string{conf}
	} else {
		conffiles = []string{
			".reviewdog.yaml",
			".reviewdog.yml",
			"reviewdog.yaml",
			"reviewdog.yml",
		}
	}
	for _, f := range conffiles {
		bytes, err := ioutil.ReadFile(f)
		if err == nil {
			return bytes, nil
		}
	}
	return nil, errors.New(".reviewdog.yml not found")
}

func newParserFromOpt(opt *option) (parser.Parser, error) {
	p, err := parser.New(&parser.Option{
		FormatName:  opt.f,
		DiffStrip:   opt.fDiffStrip,
		Errorformat: opt.efms,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to create parser. use either -f or -efm: %w", err)
	}
	return p, err
}

func toolName(opt *option) string {
	name := opt.name
	if name == "" && opt.f != "" {
		name = opt.f
	}
	return name
}

func buildRunnersMap(runners string) map[string]bool {
	m := make(map[string]bool)
	for _, r := range strings.Split(runners, ",") {
		if name := strings.TrimSpace(r); name != "" {
			m[name] = true
		}
	}
	return m
}

func getRunnersList(opt *option, conf *project.Config) []string {
	if len(opt.runners) > 0 { // if runners explicitly defined, use them
		return strings.Split(opt.runners, ",")
	}

	if conf != nil { // if this is a Project run, and no explicitly provided runners
		// if no runners explicitly provided
		// get all runners from config
		list := make([]string, 0, len(conf.Runner))
		for runner := range conf.Runner {
			list = append(list, runner)
		}
		return list
	}

	// if this is simple run, get the single tool name
	if name := toolName(opt); name != "" {
		return []string{name}
	}

	return []string{}
}
