package bitbucket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	bbv1api "github.com/gfleury/go-bitbucket-v1"
	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/reviewdog/reviewdog"
	"github.com/reviewdog/reviewdog/filter"
	"github.com/reviewdog/reviewdog/proto/rdf"
	"github.com/reviewdog/reviewdog/service/commentutil"
)

func TestBitBucketPullRequestCommenter(t *testing.T) {
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir("../..")

	alreadyCommented1 := &reviewdog.Comment{
		Result: &filter.FilteredDiagnostic{
			Diagnostic: &rdf.Diagnostic{
				Location: &rdf.Location{
					Path: "file.go",
					Range: &rdf.Range{Start: &rdf.Position{
						Line: 1,
					}},
				},
				Message: "already commented",
			},
			InDiffFile: true,
		},
	}
	alreadyCommented2 := &reviewdog.Comment{
		Result: &filter.FilteredDiagnostic{
			Diagnostic: &rdf.Diagnostic{
				Location: &rdf.Location{
					Path: "another/file.go",
					Range: &rdf.Range{Start: &rdf.Position{
						Line: 14,
					}},
				},
				Message: "already commented 2",
			},
			InDiffFile: true,
		},
	}
	newComment1 := &reviewdog.Comment{
		Result: &filter.FilteredDiagnostic{
			Diagnostic: &rdf.Diagnostic{
				Location: &rdf.Location{
					Path: "file.go",
					Range: &rdf.Range{Start: &rdf.Position{
						Line: 14,
					}},
				},
				Message: "new comment",
			},
			InDiffFile: true,
		},
	}
	newComment2 := &reviewdog.Comment{
		Result: &filter.FilteredDiagnostic{
			Diagnostic: &rdf.Diagnostic{
				Location: &rdf.Location{
					Path: "file2.go",
					Range: &rdf.Range{Start: &rdf.Position{
						Line: 15,
					}},
				},
				Message: "new comment 2",
			},
			InDiffFile: true,
		},
	}
	newComment3 := &reviewdog.Comment{
		Result: &filter.FilteredDiagnostic{
			Diagnostic: &rdf.Diagnostic{
				Location: &rdf.Location{
					Path: "new_file.go",
					Range: &rdf.Range{Start: &rdf.Position{
						Line: 14,
					}},
				},
				Message: "new comment 3",
			},
			OldPath:    "old_file.go",
			OldLine:    7,
			InDiffFile: true,
		},
	}
	commentOutsideDiff := &reviewdog.Comment{
		Result: &filter.FilteredDiagnostic{
			Diagnostic: &rdf.Diagnostic{
				Location: &rdf.Location{
					Path: "path.go",
					Range: &rdf.Range{Start: &rdf.Position{
						Line: 14,
					}},
				},
				Message: "comment outside diff",
			},
			InDiffFile: false,
		},
	}
	commentWithoutLnum := &reviewdog.Comment{
		Result: &filter.FilteredDiagnostic{
			Diagnostic: &rdf.Diagnostic{
				Location: &rdf.Location{
					Path: "path.go",
				},
				Message: "comment without lnum",
			},
			InDiffFile: true,
		},
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
	mux.HandleFunc("/rest/api/1.0/projects/o/repos/r/pull-requests/14/activities", func(w http.ResponseWriter,
		r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected access: %v %v", r.Method, r.URL)
		}
		switch r.URL.Query().Get("start") {
		default:
			resp := map[string]interface{}{
				"nextPageStart": 2,
				"isLastPage":    false,
				"values": []bbv1api.Activity{
					{
						Action: bbv1api.ActionCommented,
						Comment: bbv1api.ActivityComment{
							Text: commentutil.BitBucketMarkdownComment(alreadyCommented1),
						},
						CommentAnchor: bbv1api.Anchor{
							Path: alreadyCommented1.Result.Diagnostic.GetLocation().GetPath(),
							Line: int(alreadyCommented1.Result.Diagnostic.GetLocation().GetRange().GetStart().GetLine()),
						},
					},
					{
						Action: bbv1api.ActionCommented,
						Comment: bbv1api.ActivityComment{
							Text: "unrelated commented",
						},
						CommentAnchor: bbv1api.Anchor{
							Path: "file.go",
							Line: 1,
						},
					},
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Fatal(err)
			}
		case "2":
			resp := map[string]interface{}{
				"isLastPage": true,
				"values": []bbv1api.Activity{
					{
						Action: bbv1api.ActionCommented,
						Comment: bbv1api.ActivityComment{
							Text: commentutil.BitBucketMarkdownComment(alreadyCommented2),
						},
						CommentAnchor: bbv1api.Anchor{
							Path: alreadyCommented2.Result.Diagnostic.GetLocation().GetPath(),
							Line: int(alreadyCommented2.Result.Diagnostic.GetLocation().GetRange().GetStart().GetLine()),
						},
					},
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Fatal(err)
			}
		}
	})
	mux.HandleFunc("/rest/api/1.0/projects/o/repos/r/pull-requests/14/comments", func(w http.ResponseWriter,
		r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected access: %v %v", r.Method, r.URL)
		}
		atomic.AddInt32(&postCalled, 1)
		got := new(bbv1api.Comment)
		if err := json.NewDecoder(r.Body).Decode(got); err != nil {
			t.Error(err)
		}
		switch got.Anchor.Path {
		case "file.go":
			want := &bbv1api.Comment{
				Text: commentutil.BitBucketMarkdownComment(newComment1),
				Anchor: &bbv1api.Anchor{
					DiffType: bbv1api.DiffTypeEffective,
					LineType: bbv1api.LineTypeAdded,
					FileType: bbv1api.FileTypeTo,
					Path:     "file.go",
					Line:     14,
				},
			}
			if diff := cmp.Diff(got, want); diff != "" {
				t.Error(diff)
			}
		case "file2.go":
			want := &bbv1api.Comment{
				Text: commentutil.BitBucketMarkdownComment(newComment2),
				Anchor: &bbv1api.Anchor{
					DiffType: bbv1api.DiffTypeEffective,
					LineType: bbv1api.LineTypeAdded,
					FileType: bbv1api.FileTypeTo,
					Path:     "file2.go",
					Line:     15,
				},
			}
			if diff := cmp.Diff(got, want); diff != "" {
				t.Error(diff)
			}
		case "new_file.go":
			want := &bbv1api.Comment{
				Text: commentutil.BitBucketMarkdownComment(newComment3),
				Anchor: &bbv1api.Anchor{
					DiffType: bbv1api.DiffTypeEffective,
					LineType: bbv1api.LineTypeAdded,
					FileType: bbv1api.FileTypeTo,
					Path:     "new_file.go",
					Line:     14,
					SrcPath:  "old_file.go",
				},
			}
			if diff := cmp.Diff(got, want); diff != "" {
				t.Error(diff)
			}
		default:
			t.Errorf("got unexpected comment: %#v", got)
		}
		if err := json.NewEncoder(w).Encode(gitlab.Discussion{}); err != nil {
			t.Fatal(err)
		}
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	cli := bbv1api.NewAPIClient(context.Background(), bbv1api.NewConfiguration(ts.URL+"/rest"))

	g, err := NewPullRequestCommenter(cli, "o", "r", 14, "sha")
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
		t.Errorf("%d comments posted, but want %d", postCalled, wantPostCalled)
	}
}
