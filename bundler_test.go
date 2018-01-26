package main

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	fb "github.com/jackmordaunt/filebuilder"
)

func TestBundler_Bundle(t *testing.T) {
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
	// Note, Todo: The bundler has a dependency on the native filesystem
	// in `Bundler.convertIcon`.
	// Since the conversion process uses `ioutil.TempDir` the conversions
	// should occur in independent directories.
	// There is still a chance the `ioutil.TempDir` picks the same path
	// twice, which would produce a nice little race condition.
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()
			b := NewBundler(
				filepath.Join("src", tt.target),
				tt.title,
				tt.url,
				tt.inferIcon,
				tt.inferrer,
			)

			// Prepare an independent logger and filesystem.
			logger := testLogger{}
			b.log.Debugf = logger.Logf
			b.log.Errorf = logger.Logf
			b.fs = afero.NewMemMapFs()

			if _, err := fb.Build(b.fs, "expected", tt.expected...); err != nil {
				t.Fatalf("[%s] failed setting up test files: %v", tt.desc, err)
			}
			if _, err := fb.Build(b.fs, "src", fb.File{Path: tt.target}); err != nil {
				t.Fatalf("[%s] failed setting up test files: %v", tt.desc, err)
			}

			// Force the bundler's icon inferrer to be nil if the test
			// explicitly calls for a nil inferrer.
			if tt.inferIcon == true && tt.inferrer == nil {
				b.icon = nil
			}
			err := b.Bundle(tt.dest)
			if !tt.wantErr && err != nil {
				t.Errorf("[%s] unexpected error: %v", tt.desc, err)
				return
			}
			if tt.wantErr && err == nil {
				t.Errorf("[%s] want error, got nil", tt.desc)
				return
			}
			if tt.wantLog != "" {
				if !logger.Contains(tt.wantLog) {
					t.Errorf("[%s] wanted log message: %q", tt.desc, tt.wantLog)
				}
				return
			}
			diff, ok, err := fb.CompareDirectories(b.fs, "expected", tt.dest)
			if err != nil {
				t.Fatalf("[%s] directory comparison failed: %v", tt.desc, err)
			}
			if !ok {
				t.Errorf("[%s] want != got: \n%v", tt.desc, diff)
			}
		})
	}
}
