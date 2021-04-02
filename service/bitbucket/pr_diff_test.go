package bitbucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	bbv1api "github.com/gfleury/go-bitbucket-v1"
)

func TestBitBucketPullRequestDiff_Diff(t *testing.T) {
	getPRAPICall := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/1.0/projects/o/repos/r/pull-requests/14", func(w http.ResponseWriter, r *http.Request) {
		getPRAPICall++
		if r.Method != http.MethodGet {
			t.Errorf("unexpected access: %v %v", r.Method, r.URL)
		}
		_, err := w.Write([]byte(`{"toRef": {"latestCommit": "HEAD~"}}`))
		if err != nil {
			t.Fatal(err)
		}
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	cli := bbv1api.NewAPIClient(context.Background(), bbv1api.NewConfiguration(ts.URL+"/rest"))

	g, err := NewPullRequestDiff(cli, "o", "r", 14, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Diff(context.Background()); err != nil {
		t.Fatal(err)
	}
	if getPRAPICall != 1 {
		t.Errorf("Get GitLab MergeRequest API called %v times, want once", getPRAPICall)
	}
}
