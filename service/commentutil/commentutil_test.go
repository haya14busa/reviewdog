package commentutil

import (
	"strings"
	"testing"

	"github.com/reviewdog/reviewdog"
	"github.com/reviewdog/reviewdog/filter"
	"github.com/reviewdog/reviewdog/proto/rdf"
)

func TestCommentBody(t *testing.T) {
	tests := []struct {
		in   *reviewdog.Comment
		want string
	}{
		{
			in: &reviewdog.Comment{
				ToolName: "tool-name",
				Result: &filter.FilteredDiagnostic{
					Diagnostic: &rdf.Diagnostic{
						Message: "test message 1",
					},
				},
			},
			want: `
**[tool-name]** <sub>reported by [reviewdog](https://github.com/reviewdog/reviewdog) :dog:</sub><br>test message 1
`,
		},
		{
			in: &reviewdog.Comment{
				Result: &filter.FilteredDiagnostic{
					Diagnostic: &rdf.Diagnostic{
						Message: "test message 2 (no tool)",
					},
				},
			},
			want: `
<sub>reported by [reviewdog](https://github.com/reviewdog/reviewdog) :dog:</sub><br>test message 2 (no tool)
`,
		},
		{
			in: &reviewdog.Comment{
				ToolName: "global-tool-name",
				Result: &filter.FilteredDiagnostic{
					Diagnostic: &rdf.Diagnostic{
						Message: "test message 3",
						Source:  &rdf.Source{Name: "custom-tool-name"},
					},
				},
			},
			want: `
**[custom-tool-name]** <sub>reported by [reviewdog](https://github.com/reviewdog/reviewdog) :dog:</sub><br>test message 3
`,
		},
		{
			in: &reviewdog.Comment{
				Result: &filter.FilteredDiagnostic{
					Diagnostic: &rdf.Diagnostic{
						Message:  "test message 4",
						Source:   &rdf.Source{Name: "tool-name"},
						Severity: rdf.Severity_WARNING,
					},
				},
			},
			want: `
⚠️ **[tool-name]** <sub>reported by [reviewdog](https://github.com/reviewdog/reviewdog) :dog:</sub><br>test message 4
`,
		},
		{
			in: &reviewdog.Comment{
				Result: &filter.FilteredDiagnostic{
					Diagnostic: &rdf.Diagnostic{
						Message: "test message 5 (code)",
						Source:  &rdf.Source{Name: "tool-name"},
						Code: &rdf.Code{
							Value: "CODE14",
						},
					},
				},
			},
			want: `
**[tool-name]** <CODE14> <sub>reported by [reviewdog](https://github.com/reviewdog/reviewdog) :dog:</sub><br>test message 5 (code)
`,
		},
		{
			in: &reviewdog.Comment{
				Result: &filter.FilteredDiagnostic{
					Diagnostic: &rdf.Diagnostic{
						Message: "test message 6 (code with URL)",
						Source:  &rdf.Source{Name: "tool-name"},
						Code: &rdf.Code{
							Value: "CODE14",
							Url:   "https://example.com/#CODE14",
						},
					},
				},
			},
			want: `
**[tool-name]** <[CODE14](https://example.com/#CODE14)> <sub>reported by [reviewdog](https://github.com/reviewdog/reviewdog) :dog:</sub><br>test message 6 (code with URL)
`,
		},
	}
	for _, tt := range tests {
		want := strings.Trim(tt.want, "\n")
		if got := MarkdownComment(tt.in); got != want {
			t.Errorf("got unexpected comment.\ngot:\n%s\nwant:\n%s", got, want)
		}
	}
}

func TestMarkdownSuggestions(t *testing.T) {
	tests := []struct {
		in   *reviewdog.Comment
		want string
	}{
		{
			in: &reviewdog.Comment{
				ToolName: "tool-name",
				Result: &filter.FilteredDiagnostic{
					Diagnostic: &rdf.Diagnostic{
						Message: "no suggestion",
					},
				},
			},
			want: "",
		},
		{
			in: buildTestComment(
				"one suggestion",
				[]*rdf.Suggestion{
					buildTestsSuggestion("line1-fixed\nline2-fixed", 10, 10),
				},
			),
			want: strings.Join([]string{
				"```suggestion:-0+0",
				"line1-fixed",
				"line2-fixed",
				"```",
			}, "\n"),
		},
		{
			in: buildTestComment(
				"two suggestions",
				[]*rdf.Suggestion{
					buildTestsSuggestion("line1-fixed\nline2-fixed", 10, 11),
					buildTestsSuggestion("line3-fixed\nline4-fixed", 20, 21),
				},
			),
			want: strings.Join([]string{
				"```suggestion:-0+1",
				"line1-fixed",
				"line2-fixed",
				"```",
				"",
				"```suggestion:-0+1",
				"line3-fixed",
				"line4-fixed",
				"```",
			}, "\n"),
		},
	}
	for _, tt := range tests {
		suggestion := MarkdownSuggestions(tt.in)
		if suggestion != tt.want {
			t.Errorf("got unexpected suggestion.\ngot:\n%s\nwant:\n%s", suggestion, tt.want)
		}
	}
}

func TestMarkdownSuggestionsInvalid(t *testing.T) {
	tests := []struct {
		in   *reviewdog.Comment
		want string
	}{
		{
			in: buildTestComment(
				"two suggestions, one without range",
				[]*rdf.Suggestion{
					{
						Text: "line3-fixed\nline4-fixed",
					},
					buildTestsSuggestion("line1-fixed\nline2-fixed", 10, 11),
				},
			),
			want: strings.Join([]string{
				"```suggestion:-0+1",
				"line1-fixed",
				"line2-fixed",
				"```",
			}, "\n"),
		},
		{
			in: buildTestComment(
				"two suggestions, one without range end",
				[]*rdf.Suggestion{
					{
						Text: "line3-fixed\nline4-fixed",
						Range: &rdf.Range{
							Start: &rdf.Position{
								Line: 20,
							},
						},
					},
					buildTestsSuggestion("line1-fixed\nline2-fixed", 10, 11),
				}),
			want: strings.Join([]string{
				"```suggestion:-0+1",
				"line1-fixed",
				"line2-fixed",
				"```",
			}, "\n"),
		},
	}
	for _, tt := range tests {
		suggestion := MarkdownSuggestions(tt.in)
		if suggestion != tt.want {
			t.Errorf("got unexpected suggestion.\ngot:\n%s\nwant:\n%s", suggestion, tt.want)
		}
	}
}

func buildTestsSuggestion(text string, start int32, end int32) *rdf.Suggestion {
	return &rdf.Suggestion{
		Text: text,
		Range: &rdf.Range{
			Start: &rdf.Position{
				Line: start,
			},
			End: &rdf.Position{
				Line: end,
			},
		},
	}
}

func buildTestComment(message string, suggestions []*rdf.Suggestion) *reviewdog.Comment {
	return &reviewdog.Comment{
		ToolName: "tool-name",
		Result: &filter.FilteredDiagnostic{
			Diagnostic: &rdf.Diagnostic{
				Message:     message,
				Suggestions: suggestions,
			},
		},
	}
}
