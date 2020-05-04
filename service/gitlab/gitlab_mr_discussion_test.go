package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/reviewdog/reviewdog"
	"github.com/reviewdog/reviewdog/service/commentutil"
)

func TestGitLabMergeRequestDiscussionCommenter_Post_Flush_review_api(t *testing.T) {
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir("../..")

	alreadyCommented1 := &reviewdog.Comment{
		Result: &reviewdog.FilteredCheck{CheckResult: &reviewdog.CheckResult{
			Path: "file.go",
			Lnum: 1,
		}, InDiffFile: true},
		Body: "already commented",
	}
	alreadyCommented2 := &reviewdog.Comment{
		Result: &reviewdog.FilteredCheck{CheckResult: &reviewdog.CheckResult{
			Path: "another/file.go",
			Lnum: 14,
		}, InDiffFile: true},
		Body: "already commented 2",
	}
	newComment1 := &reviewdog.Comment{
		Result: &reviewdog.FilteredCheck{CheckResult: &reviewdog.CheckResult{
			Path: "file.go",
			Lnum: 14,
		}, InDiffFile: true},
		Body: "new comment",
	}
	newComment2 := &reviewdog.Comment{
		Result: &reviewdog.FilteredCheck{CheckResult: &reviewdog.CheckResult{
			Path: "file2.go",
			Lnum: 15,
		}, InDiffFile: true},
		Body: "new comment 2",
	}
	newComment3 := &reviewdog.Comment{
		Result: &reviewdog.FilteredCheck{
			CheckResult: &reviewdog.CheckResult{
				Path: "new_file.go",
				Lnum: 14,
			},
			OldPath:    "old_file.go",
			OldLine:    7,
			InDiffFile: true,
		},
		Body: "new comment 3",
	}
	commentOutsideDiff := &reviewdog.Comment{
		Result: &reviewdog.FilteredCheck{CheckResult: &reviewdog.CheckResult{
			Path: "path.go",
			Lnum: 14,
		}, InDiffFile: false},
		Body: "comment outside diff",
	}
	commentWithoutLnum := &reviewdog.Comment{
		Result: &reviewdog.FilteredCheck{CheckResult: &reviewdog.CheckResult{
			Path: "path.go",
			Lnum: 0,
		}, InDiffFile: true},
		Body: "comment without lnum",
	}

	comments := []*reviewdog.Comment{
		alreadyCommented1,
		alreadyCommented2,
		newComment1,
		newComment2,
		newComment3,
		commentOutsideDiff,
		commentWithoutLnum,
	}
	var postCalled int32
	const wantPostCalled = 3

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/o/r/merge_requests/14/discussions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			switch r.URL.Query().Get("page") {
			default:
				dls := []*gitlab.Discussion{
					{
						Notes: []*gitlab.Note{
							{
								Body: commentutil.CommentBody(alreadyCommented1),
								Position: &gitlab.NotePosition{
									NewPath: alreadyCommented1.Result.Path,
									NewLine: alreadyCommented1.Result.Lnum,
								},
							},
							{
								Body: "unrelated commented",
								Position: &gitlab.NotePosition{
									NewPath: "file.go",
									NewLine: 1,
								},
							},
						},
					},
				}
				w.Header().Add("X-Next-Page", "2")
				if err := json.NewEncoder(w).Encode(dls); err != nil {
					t.Fatal(err)
				}
			case "2":
				dls := []*gitlab.Discussion{
					{
						Notes: []*gitlab.Note{
							{
								Body: commentutil.CommentBody(alreadyCommented2),
								Position: &gitlab.NotePosition{
									NewPath: alreadyCommented2.Result.Path,
									NewLine: alreadyCommented2.Result.Lnum,
								},
							},
						},
					},
				}
				if err := json.NewEncoder(w).Encode(dls); err != nil {
					t.Fatal(err)
				}
			}

		case http.MethodPost:
			atomic.AddInt32(&postCalled, 1)
			got := new(gitlab.CreateMergeRequestDiscussionOptions)
			if err := json.NewDecoder(r.Body).Decode(got); err != nil {
				t.Error(err)
			}
			switch got.Position.NewPath {
			case "file.go":
				want := &gitlab.CreateMergeRequestDiscussionOptions{
					Body: gitlab.String(commentutil.CommentBody(newComment1)),
					Position: &gitlab.NotePosition{
						BaseSHA: "xxx", StartSHA: "xxx", HeadSHA: "sha", PositionType: "text", NewPath: "file.go", NewLine: 14},
				}
				if diff := cmp.Diff(got, want); diff != "" {
					t.Error(diff)
				}
			case "file2.go":
				want := &gitlab.CreateMergeRequestDiscussionOptions{
					Body: gitlab.String(commentutil.CommentBody(newComment2)),
					Position: &gitlab.NotePosition{
						BaseSHA: "xxx", StartSHA: "xxx", HeadSHA: "sha", PositionType: "text", NewPath: "file2.go", NewLine: 15},
				}
				if diff := cmp.Diff(got, want); diff != "" {
					t.Error(diff)
				}
			case "new_file.go":
				want := &gitlab.CreateMergeRequestDiscussionOptions{
					Body: gitlab.String(commentutil.CommentBody(newComment3)),
					Position: &gitlab.NotePosition{
						BaseSHA: "xxx", StartSHA: "xxx", HeadSHA: "sha", PositionType: "text",
						NewPath: "new_file.go", NewLine: 14,
						OldPath: "old_file.go", OldLine: 7,
					},
				}
				if diff := cmp.Diff(got, want); diff != "" {
					t.Error(diff)
				}
			default:
				t.Errorf("got unexpected discussion: %#v", got)
			}
			if err := json.NewEncoder(w).Encode(gitlab.Discussion{}); err != nil {
				t.Fatal(err)
			}
		default:
			t.Errorf("unexpected access: %v %v", r.Method, r.URL)
		}
	})
	mux.HandleFunc("/api/v4/projects/o/r/merge_requests/14", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected access: %v %v", r.Method, r.URL)
		}
		w.Write([]byte(`{"target_project_id": 14, "target_branch": "test-branch"}`))
	})
	mux.HandleFunc("/api/v4/projects/14/repository/branches/test-branch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected access: %v %v", r.Method, r.URL)
		}
		w.Write([]byte(`{"commit": {"id": "xxx"}}`))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	cli, err := gitlab.NewClient("", gitlab.WithBaseURL(ts.URL+"/api/v4"))
	if err != nil {
		t.Fatal(err)
	}

	g, err := NewGitLabMergeRequestDiscussionCommenter(cli, "o", "r", 14, "sha")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range comments {
		if err := g.Post(context.Background(), c); err != nil {
			t.Error(err)
		}
	}
	if err := g.Flush(context.Background()); err != nil {
		t.Errorf("%v", err)
	}
	if postCalled != wantPostCalled {
		t.Errorf("%d discussions posted, but want %d", postCalled, wantPostCalled)
	}
}
