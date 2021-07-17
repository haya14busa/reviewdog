package diff

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

//go:generate go run testdata/gen.go

func TestParseMultiFile(t *testing.T) {
	files, err := filepath.Glob("testdata/*.diff")
	if err != nil {
		t.Fatal(err)
	}
	for _, fname := range files {
		t.Log(fname)
		f, err := os.Open(fname)
		if err != nil {
			t.Fatal(err)
		}
		difffiles, err := ParseMultiFile(f)
		if err != nil {
			t.Errorf("%v: %v", fname, err)
		}

		wantfile, err := os.Open(fname + ".json")
		if err != nil {
			t.Fatal(err)
		}
		dec := json.NewDecoder(wantfile)
		var want []*FileDiff
		if err := dec.Decode(&want); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(difffiles, want) {
			l := len(difffiles)
			if len(want) > l {
				l = len(want)
			}
			for i := 0; i < l; i++ {
				var a, b *FileDiff
				if i < len(want) {
					a = want[i]
				}
				if i < len(difffiles) {
					b = difffiles[i]
				}
				t.Errorf("want %#v, got %#v", a, b)
			}
		}

		wantfile.Close()
		f.Close()
	}
}

func TestParseMultiFile_sample(t *testing.T) {
	content := `--- sample.old.txt	2016-10-13 05:09:35.820791185 +0900
+++ sample.new.txt	2016-10-13 05:15:26.839245048 +0900
@@ -1,3 +1,4 @@
 unchanged, contextual line
-deleted line
+added line
+added line
 unchanged, contextual line
--- nonewline.old.txt	2016-10-13 15:34:14.931778318 +0900
+++ nonewline.new.txt	2016-10-13 15:34:14.868444672 +0900
@@ -1,4 +1,4 @@
 " vim: nofixeol noendofline
 No newline at end of both the old and new file
-a
-a
\ No newline at end of file
+b
+b
\ No newline at end of file
`

	got, err := ParseMultiFile(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	want := []*FileDiff{
		{
			PathOld: "sample.old.txt",
			PathNew: "sample.new.txt",
			TimeOld: "2016-10-13 05:09:35.820791185 +0900",
			TimeNew: "2016-10-13 05:15:26.839245048 +0900",
			Hunks: []*Hunk{
				{
					StartLineOld: 1, LineLengthOld: 3, StartLineNew: 1, LineLengthNew: 4,
					Lines: []*Line{
						{Type: 0, Content: "unchanged, contextual line", LnumDiff: 1, LnumOld: 1, LnumNew: 1},
						{Type: 2, Content: "deleted line", LnumDiff: 2, LnumOld: 2, LnumNew: 0},
						{Type: 1, Content: "added line", LnumDiff: 3, LnumOld: 0, LnumNew: 2},
						{Type: 1, Content: "added line", LnumDiff: 4, LnumOld: 0, LnumNew: 3},
						{Type: 0, Content: "unchanged, contextual line", LnumDiff: 5, LnumOld: 3, LnumNew: 4},
					},
				},
			},
		},
		{
			PathOld: "nonewline.old.txt",
			PathNew: "nonewline.new.txt",
			TimeOld: "2016-10-13 15:34:14.931778318 +0900",
			TimeNew: "2016-10-13 15:34:14.868444672 +0900",
			Hunks: []*Hunk{
				{
					StartLineOld: 1, LineLengthOld: 4, StartLineNew: 1, LineLengthNew: 4,
					Lines: []*Line{
						{Type: 0, Content: "\" vim: nofixeol noendofline", LnumDiff: 1, LnumOld: 1, LnumNew: 1},
						{Type: 0, Content: "No newline at end of both the old and new file", LnumDiff: 2, LnumOld: 2, LnumNew: 2},
						{Type: 2, Content: "a", LnumDiff: 3, LnumOld: 3, LnumNew: 0},
						{Type: 2, Content: "a", LnumDiff: 4, LnumOld: 4, LnumNew: 0},
						{Type: 1, Content: "b", LnumDiff: 5, LnumOld: 0, LnumNew: 3},
						{Type: 1, Content: "b", LnumDiff: 6, LnumOld: 0, LnumNew: 4},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("error. in:\n%v", content)
		for _, fd := range got {
			t.Logf("FileDiff: %#v\n", fd)
			for _, h := range fd.Hunks {
				t.Logf("  Hunk%#v\n", h)
				for _, l := range h.Lines {
					t.Logf("    Line%#v\n", l)
				}
			}
		}
	}
}

func TestFileParser_Parse(t *testing.T) {
	tests := []struct {
		in   string
		want *FileDiff
	}{
		{
			in:   "",
			want: nil,
		},
		{
			in: `diff --git a/empty.txt b/empty.txtq
deleted file mode 100644
index e69de29..0000000
`,
			want: &FileDiff{
				Extended: []string{
					"diff --git a/empty.txt b/empty.txtq",
					"deleted file mode 100644",
					"index e69de29..0000000",
				},
			},
		},
		{
			in: `--- sample.old.txt	2016-10-13 05:09:35.820791185 +0900
+++ sample.new.txt	2016-10-13 05:15:26.839245048 +0900
@@ -1,3 +1,4 @@
 unchanged, contextual line
-deleted line
+added line
+added line
 unchanged, contextual line
`,
			want: &FileDiff{
				PathOld: "sample.old.txt",
				PathNew: "sample.new.txt",
				TimeOld: "2016-10-13 05:09:35.820791185 +0900",
				TimeNew: "2016-10-13 05:15:26.839245048 +0900",
				Hunks: []*Hunk{
					{
						StartLineOld: 1, LineLengthOld: 3, StartLineNew: 1, LineLengthNew: 4,
						Lines: []*Line{
							{Type: 0, Content: "unchanged, contextual line", LnumDiff: 1, LnumOld: 1, LnumNew: 1},
							{Type: 2, Content: "deleted line", LnumDiff: 2, LnumOld: 2, LnumNew: 0},
							{Type: 1, Content: "added line", LnumDiff: 3, LnumOld: 0, LnumNew: 2},
							{Type: 1, Content: "added line", LnumDiff: 4, LnumOld: 0, LnumNew: 3},
							{Type: 0, Content: "unchanged, contextual line", LnumDiff: 5, LnumOld: 3, LnumNew: 4},
						},
					},
				},
			},
		},
		{
			in: `--- sample.old.txt	2016-10-13 05:09:35.820791185 +0900
+++ sample.new.txt	2016-10-13 05:15:26.839245048 +0900
@@ -1,1 +1,1 @@
 unchanged, contextual line
@@ -2,1 +2,1 @@
 unchanged, contextual line
`,
			want: &FileDiff{
				PathOld: "sample.old.txt",
				PathNew: "sample.new.txt",
				TimeOld: "2016-10-13 05:09:35.820791185 +0900",
				TimeNew: "2016-10-13 05:15:26.839245048 +0900",
				Hunks: []*Hunk{
					{
						StartLineOld: 1, LineLengthOld: 1, StartLineNew: 1, LineLengthNew: 1,
						Lines: []*Line{
							{Type: 0, Content: "unchanged, contextual line", LnumDiff: 1, LnumOld: 1, LnumNew: 1},
						},
					},
					{
						StartLineOld: 2, LineLengthOld: 1, StartLineNew: 2, LineLengthNew: 1,
						Lines: []*Line{
							{Type: 0, Content: "unchanged, contextual line", LnumDiff: 3, LnumOld: 2, LnumNew: 2},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		p := &fileParser{r: bufio.NewReader(strings.NewReader(tt.in))}
		got, err := p.Parse()
		if err != nil {
			t.Errorf("got error %v for in:\n %v", err, tt.in)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("fileParser.Parse() = %#v, want %#v\nin: %v", got, tt.want, tt.in)
			t.Log("got:")
			for _, h := range got.Hunks {
				for _, l := range h.Lines {
					t.Logf("%#v", l)
				}
			}
			t.Log("want:")
			for _, h := range tt.want.Hunks {
				for _, l := range h.Lines {
					t.Logf("%#v", l)
				}
			}
		}
	}
}

func TestParseFileHeader(t *testing.T) {
	tests := []struct {
		in        string
		filename  string
		timestamp string
	}{
		{
			in: "--- sample.old.txt	2016-10-13 05:09:35.820791185 +0900",
			filename:  "sample.old.txt",
			timestamp: "2016-10-13 05:09:35.820791185 +0900",
		},
		{
			in:        "+++ sample.old.txt",
			filename:  "sample.old.txt",
			timestamp: "",
		},
	}
	for _, tt := range tests {
		gotf, gott := parseFileHeader(tt.in)
		if gotf != tt.filename || gott != tt.timestamp {
			t.Errorf("parseFileHeader(%v) = (%v, %v), want (%v, %v)", tt.in, gotf, gott, tt.filename, tt.timestamp)
		}
	}
}

func TestParseExtendedHeader(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{
			in: `diff --git a/sample.txt b/sample.txt
index a949a96..769bdae 100644
--- a/sample.old.txt
+++ b/sample.new.txt
@@ -1,3 +1,4 @@
`,
			want: []string{"diff --git a/sample.txt b/sample.txt", "index a949a96..769bdae 100644"},
		},
		{
			in: `diff --git a/sample.txt b/sample.txt
deleted file mode 100644
index e69de29..0000000
`,
			want: []string{"diff --git a/sample.txt b/sample.txt", "deleted file mode 100644", "index e69de29..0000000"},
		},
		{
			in: `diff --git a/sample.txt b/sample.txt
new file mode 100644
index 0000000..e69de29
diff --git a/sample2.txt b/sample2.txt
new file mode 100644
index 0000000..ee946eb
`,
			want: []string{"diff --git a/sample.txt b/sample.txt", "new file mode 100644", "index 0000000..e69de29"},
		},
		{
			in: `--- a/sample.old.txt
+++ b/sample.new.txt
@@ -1,3 +1,4 @@
`,
			want: nil,
		},
	}
	for _, tt := range tests {
		got := parseExtendedHeader(bufio.NewReader(strings.NewReader(tt.in)))
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("in:\n%v\ngot:\n%v\nwant:\n%v", tt.in, strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
		}
	}
}

func TestHunkParser_Parse(t *testing.T) {
	tests := []struct {
		in       string
		lnumdiff int
		want     *Hunk
	}{
		{
			in: `@@ -1,3 +1,4 @@ optional section heading
 unchanged, contextual line
-deleted line
+added line
+added line
 unchanged, contextual line
`,
			want: &Hunk{
				StartLineOld: 1, LineLengthOld: 3, StartLineNew: 1, LineLengthNew: 4,
				Section: "optional section heading",
				Lines: []*Line{
					{Type: 0, Content: "unchanged, contextual line", LnumDiff: 1, LnumOld: 1, LnumNew: 1},
					{Type: 2, Content: "deleted line", LnumDiff: 2, LnumOld: 2, LnumNew: 0},
					{Type: 1, Content: "added line", LnumDiff: 3, LnumOld: 0, LnumNew: 2},
					{Type: 1, Content: "added line", LnumDiff: 4, LnumOld: 0, LnumNew: 3},
					{Type: 0, Content: "unchanged, contextual line", LnumDiff: 5, LnumOld: 3, LnumNew: 4},
				},
			},
		},
		{
			in: `@@ -1,3 +1,4 @@
 unchanged, contextual line
-deleted line
+added line
+added line
 unchanged, contextual line
@@ -1,3 +1,4 @@
`,
			lnumdiff: 14,
			want: &Hunk{
				StartLineOld: 1, LineLengthOld: 3, StartLineNew: 1, LineLengthNew: 4,
				Section: "",
				Lines: []*Line{
					{Type: 0, Content: "unchanged, contextual line", LnumDiff: 15, LnumOld: 1, LnumNew: 1},
					{Type: 2, Content: "deleted line", LnumDiff: 16, LnumOld: 2, LnumNew: 0},
					{Type: 1, Content: "added line", LnumDiff: 17, LnumOld: 0, LnumNew: 2},
					{Type: 1, Content: "added line", LnumDiff: 18, LnumOld: 0, LnumNew: 3},
					{Type: 0, Content: "unchanged, contextual line", LnumDiff: 19, LnumOld: 3, LnumNew: 4},
				},
			},
		},
	}
	for _, tt := range tests {
		got, err := (&hunkParser{r: bufio.NewReader(strings.NewReader(tt.in)), lnumdiff: tt.lnumdiff}).Parse()
		if err != nil {
			t.Errorf("hunkParser.Parse(%v) got an unexpected err %v", tt.in, err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("hunkParser.Parse(%v) = \n%#v\n, want \n%#v", tt.in, got, tt.want)
			t.Logf("got lines:")
			for _, l := range got.Lines {
				t.Logf("%#v", l)
			}
			t.Logf("want lines:")
			for _, l := range tt.want.Lines {
				t.Logf("%#v", l)
			}
		}
	}
}

func TestParseHunkRange(t *testing.T) {
	tests := []struct {
		in   string
		want *hunkrange
	}{
		{
			in:   "@@ -1,3 +1,4 @@",
			want: &hunkrange{lold: 1, sold: 3, lnew: 1, snew: 4},
		},
		{
			in:   "@@ -1 +1 @@",
			want: &hunkrange{lold: 1, sold: 1, lnew: 1, snew: 1},
		},
		{
			in:   "@@ -1,3 +1,4 @@ optional section",
			want: &hunkrange{lold: 1, sold: 3, lnew: 1, snew: 4, section: "optional section"},
		},
	}
	for _, tt := range tests {
		got, err := parseHunkRange(tt.in)
		if err != nil {
			t.Errorf("parseHunkRange(%v) got an unexpected err %v", tt.in, err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseHunkRange(%v) = %#v, want %#v", tt.in, got, tt.want)
		}
	}
}

func TestParseLS(t *testing.T) {
	tests := []struct {
		in string
		l  int
		s  int
	}{
		{in: "1,3", l: 1, s: 3},
		{in: "14", l: 14, s: 1},
	}
	for _, tt := range tests {
		gotl, gots, err := parseLS(tt.in)
		if err != nil {
			t.Errorf("parseLS(%v) got an unexpected err %v", tt.in, err)
		}
		if gotl != tt.l || gots != tt.s {
			t.Errorf("parseLS(%v) = (%v, %v, _), want (%v, %v, _)", tt.in, gotl, gots, tt.l, tt.s)
		}
	}
}

func TestReadline(t *testing.T) {
	tests := []struct {
		in  string
		out []string
	}{
		{
			in: `line1
line2
line3`,
			out: []string{"line1", "line2", "line3"},
		},
		{
			in: `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Convallis tellus id interdum velit laoreet. Porttitor rhoncus dolor purus non enim praesent. Phasellus egestas tellus rutrum tellus pellentesque. Fringilla urna porttitor rhoncus dolor purus non enim praesent. Proin fermentum leo vel orci porta non pulvinar neque laoreet. Imperdiet nulla malesuada pellentesque elit eget gravida cum. Sit amet nulla facilisi morbi tempus iaculis urna id volutpat. Et molestie ac feugiat sed lectus vestibulum mattis. Tortor at auctor urna nunc id cursus metus. Nisl tincidunt eget nullam non nisi est. Tortor at auctor urna nunc id cursus metus.  Ornare massa eget egestas purus viverra accumsan. Venenatis urna cursus eget nunc scelerisque viverra mauris in. Platea dictumst vestibulum rhoncus est pellentesque elit. Auctor neque vitae tempus quam pellentesque nec. Vivamus arcu felis bibendum ut tristique et egestas quis. At in tellus integer feugiat scelerisque varius morbi. Imperdiet nulla malesuada pellentesque elit eget gravida cum sociis natoque. Ac turpis egestas integer eget aliquet nibh praesent tristique. Senectus et netus et malesuada fames ac turpis egestas. Odio eu feugiat pretium nibh ipsum consequat nisl vel. Elit eget gravida cum sociis natoque penatibus et.  Curabitur gravida arcu ac tortor dignissim convallis aenean et tortor. Tincidunt arcu non sodales neque sodales ut etiam sit. Fringilla phasellus faucibus scelerisque eleifend donec pretium vulputate. Urna cursus eget nunc scelerisque viverra mauris in aliquam. Fames ac turpis egestas sed. Vivamus arcu felis bibendum ut tristique et egestas quis ipsum. Turpis tincidunt id aliquet risus feugiat in ante metus dictum. Leo a diam sollicitudin tempor id eu. Consectetur adipiscing elit ut aliquam. Facilisi nullam vehicula ipsum a arcu. Dignissim suspendisse in est ante in nibh mauris cursus. Faucibus interdum posuere lorem ipsum dolor sit amet consectetur. Neque convallis a cras semper. Pellentesque habitant morbi tristique senectus. Arcu non sodales neque sodales ut etiam. Pretium quam vulputate dignissim suspendisse in. Id velit ut tortor pretium viverra suspendisse potenti. Venenatis urna cursus eget nunc scelerisque. Vestibulum lorem sed risus ultricies tristique. Faucibus purus in massa tempor.  Tellus elementum sagittis vitae et. Ut tortor pretium viverra suspendisse potenti nullam ac. Non arcu risus quis varius. Magna ac placerat vestibulum lectus mauris ultrices eros in. Venenatis lectus magna fringilla urna porttitor. Ultrices gravida dictum fusce ut. Hac habitasse platea dictumst quisque sagittis purus sit. Dictum at tempor commodo ullamcorper a lacus. Mi bibendum neque egestas congue quisque. Lobortis elementum nibh tellus molestie nunc non. Dolor sit amet consectetur adipiscing elit pellentesque habitant morbi tristique.  Pulvinar sapien et ligula ullamcorper malesuada proin libero nunc. Vel elit scelerisque mauris pellentesque pulvinar pellentesque. Eget nulla facilisi etiam dignissim diam quis enim. Convallis posuere morbi leo urna molestie. Pellentesque adipiscing commodo elit at imperdiet dui. Mattis pellentesque id nibh tortor id aliquet. Praesent semper feugiat nibh sed pulvinar proin gravida hendrerit lectus. Enim lobortis scelerisque fermentum dui faucibus in ornare quam. Maecenas pharetra convallis posuere morbi leo. Dolor morbi non arcu risus quis. Sit amet purus gravida quis. Quis varius quam quisque id diam vel quam elementum. Tortor condimentum lacinia quis vel eros donec. Gravida neque convallis a cras semper auctor neque vitae tempus. Odio eu feugiat pretium nibh ipsum consequat nisl vel pretium. Urna et pharetra pharetra massa massa ultricies mi. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Convallis tellus id interdum velit laoreet. Porttitor rhoncus dolor purus non enim praesent. Phasellus egestas tellus rutrum tellus pellentesque. Fringilla urna porttitor rhoncus dolor purus non enim praesent. Proin fermentum leo vel orci porta non pulvinar neque laoreet. Imperdiet nulla malesuada pellentesque elit eget gravida cum. Sit amet nulla facilisi morbi tempus iaculis urna id volutpat. Et molestie ac feugiat sed lectus vestibulum mattis. Tortor at auctor urna nunc id cursus metus. Nisl tincidunt eget nullam non nisi est. Tortor at auctor urna nunc id cursus metus.  Ornare massa eget egestas purus viverra accumsan. Venenatis urna cursus eget nunc scelerisque viverra mauris in. Platea dictumst vestibulum rhoncus est pellentesque elit. Auctor neque vitae tempus quam pellentesque nec. Vivamus arcu felis bibendum ut tristique et egestas quis. At in tellus integer feugiat scelerisque varius morbi. Imperdiet nulla malesuada pellentesque elit eget gravida cum sociis natoque. Ac turpis egestas integer eget aliquet nibh praesent tristique. Senectus et netus et malesuada fames ac turpis egestas. Odio eu feugiat pretium nibh ipsum consequat nisl vel. Elit eget gravida cum sociis natoque penatibus et.  Curabitur gravida arcu ac tortor dignissim convallis aenean et tortor. Tincidunt arcu non sodales neque sodales ut etiam sit. Fringilla phasellus faucibus scelerisque eleifend donec pretium vulputate. Urna cursus eget nunc scelerisque viverra mauris in aliquam. Fames ac turpis egestas sed. Vivamus arcu felis bibendum ut tristique et egestas quis ipsum. Turpis tincidunt id aliquet risus feugiat in ante metus dictum. Leo a diam sollicitudin tempor id eu. Consectetur adipiscing elit ut aliquam. Facilisi nullam vehicula ipsum a arcu. Dignissim suspendisse in est ante in nibh mauris cursus. Faucibus interdum posuere lorem ipsum dolor sit amet consectetur. Neque convallis a cras semper. Pellentesque habitant morbi tristique senectus. Arcu non sodales neque sodales ut etiam. Pretium quam vulputate dignissim suspendisse in. Id velit ut tortor pretium viverra suspendisse potenti. Venenatis urna cursus eget nunc scelerisque. Vestibulum lorem sed risus ultricies tristique. Faucibus purus in massa tempor.  Tellus elementum sagittis vitae et. Ut tortor pretium viverra suspendisse potenti nullam ac. Non arcu risus quis varius. Magna ac placerat vestibulum lectus mauris ultrices eros in. Venenatis lectus magna fringilla urna porttitor. Ultrices gravida dictum fusce ut. Hac habitasse platea dictumst quisque sagittis purus sit. Dictum at tempor commodo ullamcorper a lacus. Mi bibendum neque egestas congue quisque. Lobortis elementum nibh tellus molestie nunc non. Dolor sit amet consectetur adipiscing elit pellentesque habitant morbi tristique.  Pulvinar sapien et ligula ullamcorper malesuada proin libero nunc. Vel elit scelerisque mauris pellentesque pulvinar pellentesque. Eget nulla facilisi etiam dignissim diam quis enim. Convallis posuere morbi leo urna molestie. Pellentesque adipiscing commodo elit at imperdiet dui. Mattis pellentesque id nibh tortor id aliquet. Praesent semper feugiat nibh sed pulvinar proin gravida hendrerit lectus. Enim lobortis scelerisque fermentum dui faucibus in ornare quam. Maecenas pharetra convallis posuere morbi leo. Dolor morbi non arcu risus quis. Sit amet purus gravida quis. Quis varius quam quisque id diam vel quam elementum. Tortor condimentum lacinia quis vel eros donec. Gravida neque convallis a cras semper auctor neque vitae tempus. Odio eu feugiat pretium nibh ipsum consequat nisl vel pretium. Urna et pharetra pharetra massa massa ultricies mi.
line2
line3
`,
			out: []string{"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Convallis tellus id interdum velit laoreet. Porttitor rhoncus dolor purus non enim praesent. Phasellus egestas tellus rutrum tellus pellentesque. Fringilla urna porttitor rhoncus dolor purus non enim praesent. Proin fermentum leo vel orci porta non pulvinar neque laoreet. Imperdiet nulla malesuada pellentesque elit eget gravida cum. Sit amet nulla facilisi morbi tempus iaculis urna id volutpat. Et molestie ac feugiat sed lectus vestibulum mattis. Tortor at auctor urna nunc id cursus metus. Nisl tincidunt eget nullam non nisi est. Tortor at auctor urna nunc id cursus metus.  Ornare massa eget egestas purus viverra accumsan. Venenatis urna cursus eget nunc scelerisque viverra mauris in. Platea dictumst vestibulum rhoncus est pellentesque elit. Auctor neque vitae tempus quam pellentesque nec. Vivamus arcu felis bibendum ut tristique et egestas quis. At in tellus integer feugiat scelerisque varius morbi. Imperdiet nulla malesuada pellentesque elit eget gravida cum sociis natoque. Ac turpis egestas integer eget aliquet nibh praesent tristique. Senectus et netus et malesuada fames ac turpis egestas. Odio eu feugiat pretium nibh ipsum consequat nisl vel. Elit eget gravida cum sociis natoque penatibus et.  Curabitur gravida arcu ac tortor dignissim convallis aenean et tortor. Tincidunt arcu non sodales neque sodales ut etiam sit. Fringilla phasellus faucibus scelerisque eleifend donec pretium vulputate. Urna cursus eget nunc scelerisque viverra mauris in aliquam. Fames ac turpis egestas sed. Vivamus arcu felis bibendum ut tristique et egestas quis ipsum. Turpis tincidunt id aliquet risus feugiat in ante metus dictum. Leo a diam sollicitudin tempor id eu. Consectetur adipiscing elit ut aliquam. Facilisi nullam vehicula ipsum a arcu. Dignissim suspendisse in est ante in nibh mauris cursus. Faucibus interdum posuere lorem ipsum dolor sit amet consectetur. Neque convallis a cras semper. Pellentesque habitant morbi tristique senectus. Arcu non sodales neque sodales ut etiam. Pretium quam vulputate dignissim suspendisse in. Id velit ut tortor pretium viverra suspendisse potenti. Venenatis urna cursus eget nunc scelerisque. Vestibulum lorem sed risus ultricies tristique. Faucibus purus in massa tempor.  Tellus elementum sagittis vitae et. Ut tortor pretium viverra suspendisse potenti nullam ac. Non arcu risus quis varius. Magna ac placerat vestibulum lectus mauris ultrices eros in. Venenatis lectus magna fringilla urna porttitor. Ultrices gravida dictum fusce ut. Hac habitasse platea dictumst quisque sagittis purus sit. Dictum at tempor commodo ullamcorper a lacus. Mi bibendum neque egestas congue quisque. Lobortis elementum nibh tellus molestie nunc non. Dolor sit amet consectetur adipiscing elit pellentesque habitant morbi tristique.  Pulvinar sapien et ligula ullamcorper malesuada proin libero nunc. Vel elit scelerisque mauris pellentesque pulvinar pellentesque. Eget nulla facilisi etiam dignissim diam quis enim. Convallis posuere morbi leo urna molestie. Pellentesque adipiscing commodo elit at imperdiet dui. Mattis pellentesque id nibh tortor id aliquet. Praesent semper feugiat nibh sed pulvinar proin gravida hendrerit lectus. Enim lobortis scelerisque fermentum dui faucibus in ornare quam. Maecenas pharetra convallis posuere morbi leo. Dolor morbi non arcu risus quis. Sit amet purus gravida quis. Quis varius quam quisque id diam vel quam elementum. Tortor condimentum lacinia quis vel eros donec. Gravida neque convallis a cras semper auctor neque vitae tempus. Odio eu feugiat pretium nibh ipsum consequat nisl vel pretium. Urna et pharetra pharetra massa massa ultricies mi. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Convallis tellus id interdum velit laoreet. Porttitor rhoncus dolor purus non enim praesent. Phasellus egestas tellus rutrum tellus pellentesque. Fringilla urna porttitor rhoncus dolor purus non enim praesent. Proin fermentum leo vel orci porta non pulvinar neque laoreet. Imperdiet nulla malesuada pellentesque elit eget gravida cum. Sit amet nulla facilisi morbi tempus iaculis urna id volutpat. Et molestie ac feugiat sed lectus vestibulum mattis. Tortor at auctor urna nunc id cursus metus. Nisl tincidunt eget nullam non nisi est. Tortor at auctor urna nunc id cursus metus.  Ornare massa eget egestas purus viverra accumsan. Venenatis urna cursus eget nunc scelerisque viverra mauris in. Platea dictumst vestibulum rhoncus est pellentesque elit. Auctor neque vitae tempus quam pellentesque nec. Vivamus arcu felis bibendum ut tristique et egestas quis. At in tellus integer feugiat scelerisque varius morbi. Imperdiet nulla malesuada pellentesque elit eget gravida cum sociis natoque. Ac turpis egestas integer eget aliquet nibh praesent tristique. Senectus et netus et malesuada fames ac turpis egestas. Odio eu feugiat pretium nibh ipsum consequat nisl vel. Elit eget gravida cum sociis natoque penatibus et.  Curabitur gravida arcu ac tortor dignissim convallis aenean et tortor. Tincidunt arcu non sodales neque sodales ut etiam sit. Fringilla phasellus faucibus scelerisque eleifend donec pretium vulputate. Urna cursus eget nunc scelerisque viverra mauris in aliquam. Fames ac turpis egestas sed. Vivamus arcu felis bibendum ut tristique et egestas quis ipsum. Turpis tincidunt id aliquet risus feugiat in ante metus dictum. Leo a diam sollicitudin tempor id eu. Consectetur adipiscing elit ut aliquam. Facilisi nullam vehicula ipsum a arcu. Dignissim suspendisse in est ante in nibh mauris cursus. Faucibus interdum posuere lorem ipsum dolor sit amet consectetur. Neque convallis a cras semper. Pellentesque habitant morbi tristique senectus. Arcu non sodales neque sodales ut etiam. Pretium quam vulputate dignissim suspendisse in. Id velit ut tortor pretium viverra suspendisse potenti. Venenatis urna cursus eget nunc scelerisque. Vestibulum lorem sed risus ultricies tristique. Faucibus purus in massa tempor.  Tellus elementum sagittis vitae et. Ut tortor pretium viverra suspendisse potenti nullam ac. Non arcu risus quis varius. Magna ac placerat vestibulum lectus mauris ultrices eros in. Venenatis lectus magna fringilla urna porttitor. Ultrices gravida dictum fusce ut. Hac habitasse platea dictumst quisque sagittis purus sit. Dictum at tempor commodo ullamcorper a lacus. Mi bibendum neque egestas congue quisque. Lobortis elementum nibh tellus molestie nunc non. Dolor sit amet consectetur adipiscing elit pellentesque habitant morbi tristique.  Pulvinar sapien et ligula ullamcorper malesuada proin libero nunc. Vel elit scelerisque mauris pellentesque pulvinar pellentesque. Eget nulla facilisi etiam dignissim diam quis enim. Convallis posuere morbi leo urna molestie. Pellentesque adipiscing commodo elit at imperdiet dui. Mattis pellentesque id nibh tortor id aliquet. Praesent semper feugiat nibh sed pulvinar proin gravida hendrerit lectus. Enim lobortis scelerisque fermentum dui faucibus in ornare quam. Maecenas pharetra convallis posuere morbi leo. Dolor morbi non arcu risus quis. Sit amet purus gravida quis. Quis varius quam quisque id diam vel quam elementum. Tortor condimentum lacinia quis vel eros donec. Gravida neque convallis a cras semper auctor neque vitae tempus. Odio eu feugiat pretium nibh ipsum consequat nisl vel pretium. Urna et pharetra pharetra massa massa ultricies mi.", "line2", "line3"},
		},
	}
	for _, tt := range tests {
		r := bufio.NewReader(strings.NewReader(tt.in))
		for _, exp := range tt.out {
			got, err := readline(r)
			if err != nil {
				t.Error(err)
			}
			if got != exp {
				t.Errorf("got %v, want %v", got, exp)
			}
		}
		{
			if _, err := readline(r); err != io.EOF {
				t.Errorf("got err %v, want io.EOF", err)
			}
		}
	}
}

func TestUnquoteCStyle(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{in: `no need to unquote`, out: `no need to unquote`},
		{in: `"C-escapes \a\b\t\n\v\f\r\"\\"`, out: "C-escapes \a\b\t\n\v\f\r\"\\"},

		// from https://github.com/git/git/blob/041f5ea1cf987a4068ef5f39ba0a09be85952064/t/t3902-quoted.sh#L48-L76
		{in: `Name`, out: `Name`},
		{in: `"Name and a\nLF"`, out: "Name and a\nLF"},
		{in: `"Name and an\tHT"`, out: "Name and an\tHT"},
		{in: `"Name\""`, out: `Name"`},
		{in: `With SP in it`, out: `With SP in it`},
		{in: `"\346\277\261\351\207\216\t\347\264\224"`, out: "濱野\t純"},
		{in: `"\346\277\261\351\207\216\n\347\264\224"`, out: "濱野\n純"},
		{in: `"\346\277\261\351\207\216 \347\264\224"`, out: `濱野 純`},
		{in: `"\346\277\261\351\207\216\"\347\264\224"`, out: "濱野\"純"},
		{in: `"\346\277\261\351\207\216/file"`, out: `濱野/file`},
		{in: `"\346\277\261\351\207\216\347\264\224"`, out: `濱野純`},

		// Edge cases of ill-formed diff file name.
		{in: `\347\264\224`, out: `\347\264\224`}, // no need to unquote
		{in: `"\34a"`, out: "34a"},
		{in: `"\14"`, out: `14`},
	}

	for _, tt := range tests {
		if got := unquoteCStyle(tt.in); got != tt.out {
			t.Errorf("unquoteCStyle(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}
