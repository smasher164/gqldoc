package gqldoc_test

import (
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
