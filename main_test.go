package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden (.out) files")

// This is the directory where our test fixtures are.
const fixtureDir = "./test-fixtures"

func TestRewrite(t *testing.T) {
	test := []struct {
		cfg  *config
		file string
	}{
		{
			file: "struct_add",
			cfg: &config{
				add:        []string{"json"},
				output:     "source",
				structName: "foo",
				transform:  "snakecase",
			},
		},
		{
			file: "struct_add_existing",
			cfg: &config{
				add:        []string{"json"},
				output:     "source",
				structName: "foo",
				transform:  "snakecase",
			},
		},
		{
			file: "struct_remove",
			cfg: &config{
				remove:     []string{"json"},
				output:     "source",
				structName: "foo",
			},
		},
		{
			file: "struct_clear_tags",
			cfg: &config{
				clear:      true,
				output:     "source",
				structName: "foo",
			},
		},
		{
			file: "struct_clear_options",
			cfg: &config{
				clearOption: true,
				output:      "source",
				structName:  "foo",
			},
		},
		{
			file: "line_add",
			cfg: &config{
				add:       []string{"json"},
				output:    "source",
				line:      "4",
				transform: "snakecase",
			},
		},
		{
			file: "line_add_comment",
			cfg: &config{
				add:       []string{"json"},
				output:    "source",
				line:      "6,7",
				transform: "snakecase",
			},
		},
		{
			file: "line_add_option",
			cfg: &config{
				addOptions: []string{"json=omitempty"},
				output:     "source",
				line:       "4,7",
			},
		},
		{
			file: "line_add_option_existing",
			cfg: &config{
				addOptions: []string{"json=omitempty"},
				output:     "source",
				line:       "6,8",
			},
		},
		{
			file: "line_add_multiple_option",
			cfg: &config{
				addOptions: []string{"json=omitempty", "hcl=squash"},
				add:        []string{"hcl"},
				output:     "source",
				line:       "4,7",
				transform:  "snakecase",
			},
		},
		{
			file: "line_remove",
			cfg: &config{
				remove: []string{"json"},
				output: "source",
				line:   "5,7",
			},
		},
		{
			file: "line_remove_option",
			cfg: &config{
				removeOptions: []string{"hcl=squash"},
				output:        "source",
				line:          "4,8",
			},
		},
		{
			file: "line_remove_options",
			cfg: &config{
				removeOptions: []string{"json=omitempty", "hcl=omitnested"},
				output:        "source",
				line:          "4,7",
			},
		},
		{
			file: "line_multiple_add",
			cfg: &config{
				add:       []string{"json"},
				output:    "source",
				line:      "5,6",
				transform: "camelcase",
			},
		},
		{
			file: "line_camelcase_add",
			cfg: &config{
				add:       []string{"json"},
				output:    "source",
				line:      "4,5",
				transform: "camelcase",
			},
		},
		{
			file: "line_value_add",
			cfg: &config{
				add:    []string{"json:foo"},
				output: "source",
				line:   "4,6",
			},
		},
		{
			file: "offset_add",
			cfg: &config{
				add:       []string{"json"},
				output:    "source",
				offset:    32,
				transform: "snakecase",
			},
		},
		{
			file: "offset_add_duplicate",
			cfg: &config{
				add:       []string{"json"},
				output:    "source",
				offset:    209,
				transform: "snakecase",
			},
		},
	}

	for _, ts := range test {
		t.Run(ts.file, func(t *testing.T) {
			ts.cfg.file = filepath.Join(fixtureDir, fmt.Sprintf("%s.input", ts.file))
			node, err := ts.cfg.rewrite()
			if err != nil {
				t.Fatal(err)
			}

			out, err := ts.cfg.format(node)
			if err != nil {
				t.Fatal(err)
			}
			got := []byte(out)

			// update golden file if necessary
			golden := filepath.Join(fixtureDir, fmt.Sprintf("%s.golden", ts.file))
			if *update {
				err := ioutil.WriteFile(golden, got, 0644)
				if err != nil {
					t.Error(err)
				}
				return
			}

			// get golden file
			want, err := ioutil.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}

			var from []byte
			if ts.cfg.modified != nil {
				from, err = ioutil.ReadAll(ts.cfg.modified)
			} else {
				from, err = ioutil.ReadFile(ts.cfg.file)
			}
			if err != nil {
				t.Fatal(err)
			}

			// compare
			if !bytes.Equal(got, want) {
				t.Errorf("case %s\ngot:\n====\n\n%s\nwant:\n=====\n\n%s\nfrom:\n=====\n\n%s\n",
					ts.file, got, want, from)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	test := []struct {
		cfg  *config
		file string
	}{
		{
			file: "line_camelcase_add_json",
			cfg: &config{
				add:       []string{"json"},
				output:    "json",
				line:      "4,5",
				transform: "camelcase",
			},
		},
	}

	for _, ts := range test {
		t.Run(ts.file, func(t *testing.T) {
			ts.cfg.file = filepath.Join(fixtureDir, fmt.Sprintf("%s.input", ts.file))
			node, err := ts.cfg.rewrite()
			if err != nil {
				t.Fatal(err)
			}

			out, err := ts.cfg.format(node)
			if err != nil {
				t.Fatal(err)
			}
			got := []byte(out)

			// update golden file if necessary
			golden := filepath.Join(fixtureDir, fmt.Sprintf("%s.golden", ts.file))
			if *update {
				err := ioutil.WriteFile(golden, got, 0644)
				if err != nil {
					t.Error(err)
				}
				return
			}

			// get golden file
			want, err := ioutil.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}

			from, err := ioutil.ReadFile(ts.cfg.file)
			if err != nil {
				t.Fatal(err)
			}

			// compare
			if !bytes.Equal(got, want) {
				t.Errorf("case %s\ngot:\n====\n\n%s\nwant:\n=====\n\n%s\nfrom:\n=====\n\n%s\n",
					ts.file, got, want, from)
			}
		})
	}
}

func TestModifiedRewrite(t *testing.T) {
	cfg := &config{
		add:        []string{"json"},
		output:     "source",
		structName: "foo",
		transform:  "snakecase",
		file:       "struct_add_modified",
		modified: strings.NewReader(`struct_add_modified
55
package foo

type foo struct {
	bar string
	t   bool
}
`),
	}
	node, err := cfg.rewrite()
	if err != nil {
		t.Fatal(err)
	}

	got, err := cfg.format(node)
	if err != nil {
		t.Fatal(err)
	}

	golden := filepath.Join(fixtureDir, "struct_add.golden")
	want, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}

	// compare
	if !bytes.Equal([]byte(got), want) {
		t.Errorf("got:\n====\n%s\nwant:\n====\n%s\n", got, want)
	}
}
