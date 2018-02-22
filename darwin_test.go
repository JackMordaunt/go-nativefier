package nativefier

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	fb "github.com/jackmordaunt/filebuilder"
)

func TestBundler_Bundle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc string

		// input
		target    string
		title     string
		url       string
		inferIcon bool
		inferrer  IconInferrer

		// output
		dest     string
		expected []fb.Entry
		wantErr  bool
		wantLog  string
	}{
		{
			desc:      "icon not infered",
			target:    "target/binary.exe",
			title:     "test",
			url:       "https://example.com",
			inferIcon: false,

			dest: "dest",
			expected: []fb.Entry{
				fb.Dir{Path: "test.app", Entries: []fb.Entry{
					fb.Dir{Path: "Contents", Entries: []fb.Entry{
						fb.File{Path: "MacOS/binary.exe"},
						fb.File{Path: "MacOS/config.json"},
						fb.File{Path: "Info.plist"},
						fb.Dir{Path: "Resources"},
					}},
				}},
			},
			wantErr: false,
		},
		{
			desc:      "infer icon with valid inferrer",
			target:    "target/binary.exe",
			title:     "test",
			url:       "https://example.com",
			inferIcon: true,
			inferrer:  mockInferrer{},

			dest: "dest",
			expected: []fb.Entry{
				fb.Dir{Path: "test.app", Entries: []fb.Entry{
					fb.Dir{Path: "Contents", Entries: []fb.Entry{
						fb.File{Path: "MacOS/binary.exe"},
						fb.File{Path: "MacOS/config.json"},
						fb.File{Path: "Resources/icon.icns"},
						fb.File{Path: "Info.plist"},
					}},
				}},
			},
			wantErr: false,
		},
		{
			desc:      "infer icon without valid icon inferrer",
			target:    "target/binary.exe",
			title:     "test",
			url:       "https://example.com",
			inferIcon: true,
			inferrer:  nil,

			dest:    "dest",
			wantLog: errNoInferrer.Error(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(st *testing.T) {
			b := &Darwin{
				Target:    filepath.Join("src", tt.target),
				Title:     tt.title,
				URL:       tt.url,
				InferIcon: tt.inferIcon,
				icon:      tt.inferrer,
				fs:        afero.NewMemMapFs(),
			}

			if _, err := fb.Build(b.fs, "expected", tt.expected...); err != nil {
				st.Fatalf("failed setting up test files: %v", err)
			}
			if _, err := fb.Build(b.fs, "src", fb.File{Path: tt.target}); err != nil {
				st.Fatalf("failed setting up test files: %v", err)
			}

			// Force the bundler's icon inferrer to be nil if the test
			// explicitly calls for a nil inferrer.
			if tt.inferIcon == true && tt.inferrer == nil {
				b.icon = nil
			}
			err := b.Pack(tt.dest)
			if !tt.wantErr && err != nil {
				st.Errorf("unexpected error: %v", err)
				return
			}
			if tt.wantErr && err == nil {
				st.Errorf("want error, got nil")
				return
			}
			diff, ok, err := fb.CompareDirectories(b.fs, "expected", tt.dest)
			if err != nil {
				st.Fatalf("directory comparison failed: %v", err)
			}
			if !ok {
				st.Errorf("want != got: \n%v", diff)
			}
		})
	}
}
