package gqldoc_test

import (
	"errors"
	"io/ioutil"
	"testing"

	"github.com/smasher164/gqldoc"
)

func TestParseFiles(t *testing.T) {
	cases := []struct {
		paths   []string
		wanterr bool
	}{
		{paths: nil, wanterr: true},
		{paths: []string{"testdata/valid.gql"}, wanterr: false},
		{paths: []string{"testdata/valid.graphql"}, wanterr: false},
		{paths: []string{"testdata/invalid.gql"}, wanterr: true},
		{paths: []string{"testdata/doesnotexist.gql"}, wanterr: true},
		{paths: []string{"testdata/invalid.extension"}, wanterr: true},
		{paths: []string{"testdata/1.gql"}, wanterr: true},
		{paths: []string{"testdata/1.gql", "testdata/2.gql"}, wanterr: false},
	}
	for _, testcase := range cases {
		if _, err := gqldoc.ParseFiles(testcase.paths); testcase.wanterr != (err != nil) {
			t.Errorf("ParseFiles(%v) wanterr: %v, got: %v", testcase.paths, testcase.wanterr, err)
		}
	}
}

type errWriter bool

func (w *errWriter) Write(p []byte) (n int, err error) {
	if *w {
		return 0, errors.New("ERROR")
	}
	*w = true
	return 0, nil
}

func TestRenderGraphQL(t *testing.T) {
	schema, err := gqldoc.ParseFiles([]string{"testdata/valid.gql"})
	if err != nil {
		t.Fatal(err)
	}
	if err := gqldoc.FormatGraphQL(new(errWriter), schema); err.Error() != "ERROR" {
		t.Error(err)
	}
}

func TestRenderMarkdown(t *testing.T) {
	schema, err := gqldoc.ParseFiles([]string{"testdata/star-wars.graphql"})
	if err != nil {
		t.Fatal(err)
	}
	if err := gqldoc.FormatMarkdown(ioutil.Discard, schema); err != nil {
		t.Error(err)
	}
}
